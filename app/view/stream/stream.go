package stream

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/wasya-io/petit-misskey/domain/api"
	"github.com/wasya-io/petit-misskey/domain/core"
	"github.com/wasya-io/petit-misskey/infrastructure/bubbles"
	"github.com/wasya-io/petit-misskey/infrastructure/setting"
	"github.com/wasya-io/petit-misskey/infrastructure/websocket"
	"github.com/wasya-io/petit-misskey/model/misskey"
	"github.com/wasya-io/petit-misskey/view/postnote"
)

type Model struct {
	ctx          context.Context
	logger       core.Logger
	cancel       context.CancelFunc
	msgCh        chan tea.Msg
	viewMain     viewport.Model
	viewStatus   viewport.Model
	textarea     postnote.PostTextarea
	client       websocket.Client
	apiClient    api.Client
	notes        []*misskey.Note
	quitting     bool
	connected    bool
	err          error
	instance     *setting.Instance
	viewBuffer   strings.Builder
	width        int
	height       int
	initialized  bool
	timeline     string
	muViewAll    sync.Mutex
	muViewStatus sync.Mutex
}

var (
	//go:embed template/note.tmpl
	NoteTmpl string
	//go:embed template/renote.tmpl
	RenoteTmpl string
)

func NewModel(instance *setting.Instance, client websocket.Client, apiClient api.Client, logger core.Logger, msgCh chan tea.Msg) *Model {
	ctx, cancel := context.WithCancel(context.Background())

	m := &Model{
		ctx:          ctx,
		logger:       logger,
		cancel:       cancel,
		msgCh:        msgCh,
		viewMain:     bubbles.NewViewportFactory().StreamView(),
		viewStatus:   bubbles.NewViewportFactory().SystemView(),
		client:       client,
		apiClient:    apiClient,
		notes:        make([]*misskey.Note, 0, 100),
		quitting:     false,
		connected:    false,
		instance:     instance,
		viewBuffer:   strings.Builder{},
		width:        120,
		height:       20,
		initialized:  false,
		timeline:     "",
		muViewAll:    sync.Mutex{},
		muViewStatus: sync.Mutex{},
	}
	m.textarea = bubbles.NewViewportFactory().PostView(m.postnoteCallback, logger)
	return m
}

func (m *Model) Init() tea.Cmd {
	// WebSocketクライアントのgoroutine起動コマンドを返す
	return func() tea.Msg {
		// 別goroutineでWebSocket接続を開始
		go func() {
			if err := m.client.Start(); err != nil {
				m.msgCh <- tea.Msg(websocket.WebSocketErrorMsg{Err: err})
			}
		}()

		// 初期化完了を通知
		return textarea.Blink()
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.logger.Log("stream", fmt.Sprintf("msg: %T", msg))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			m.client.Stop()
			m.logger.Log("stream", "終了処理を開始します")
			// ロガーを正しく終了し、残りのログをフラッシュします
			m.logger.Close()
			return m, tea.Quit

		case "ctrl+h":
			err := m.client.ToggleTimeline()
			if err != nil {
				m.logger.Log("stream", fmt.Sprintf("timeline error: %v", err))
				return m, nil
			}
			m.notes = make([]*misskey.Note, 0, 100)
			m.refreshViewBuffer()
			return m, nil
		case "ctrl+l":
			err := m.client.ToggleTimeline()
			if err != nil {
				m.logger.Log("stream", fmt.Sprintf("timeline error: %v", err))
				return m, nil
			}
			m.notes = make([]*misskey.Note, 0, 100)
			m.refreshViewBuffer()
			return m, nil
		default:
			t, cmd := m.textarea.Update(msg)
			m.textarea = t

			return m, cmd
		}

	case websocket.NoteMessage:
		if msg.Note.Body.Body.RenoteID != "" {
			m.logger.Log("stream", fmt.Sprintf("renote: %s", msg.Note.Body.Body.Renote.Text))
		} else {
			m.logger.Log("stream", fmt.Sprintf("note: %s", msg.Note.Body.Body.Text))
		}
		m.notes = append([]*misskey.Note{msg.Note}, m.notes...)
		if len(m.notes) > 10 {
			m.notes = m.notes[:10]
		}
		m.refreshViewBuffer()
		return m, nil

	case tea.WindowSizeMsg:

		m.width = msg.Width
		m.height = msg.Height
		m.refreshViewBuffer()
		return m, nil

	case websocket.WebSocketConnectedMsg:
		m.connected = true
		m.err = nil
		m.timeline = msg.Timeline.String()
		m.refreshStatusView()
		return m, nil

	case websocket.WebSocketPingReceivedMsg:
		// m.viewFooter.SetContent("")
		// m.viewMain.SetContent(msg.Data)
		m.logger.Log("stream", fmt.Sprintf("ping received: %s", msg.Data))
		return m, nil

	case websocket.WebSocketDisconnectedMsg:
		m.connected = false
		if msg.Err != nil {
			m.err = msg.Err
		}
		m.viewStatus.SetContent("")
		m.viewMain.SetContent("Disconnected")
		return m, nil

	case websocket.WebSocketErrorMsg:
		m.err = msg.Err
		m.viewStatus.SetContent("")
		m.viewMain.SetContent(msg.Err.Error())
		return m, nil

	case websocket.TimelineChangedMsg:
		m.timeline = msg.NewTimeline.String()
		m.refreshStatusView()
		return m, nil
	}
	return m, nil
}

func (m *Model) View() string {
	m.viewMain.Width = m.width
	m.viewStatus.Width = m.width

	joinedView := lipgloss.JoinVertical(
		lipgloss.Left,
		m.viewStatus.View(),
		m.viewMain.View(),
		m.textarea.View(),
	)
	return joinedView
}

func (m *Model) MsgChannel() chan tea.Msg {
	return m.msgCh
}

func (m *Model) refreshStatusView() {
	m.muViewStatus.Lock()
	defer m.muViewStatus.Unlock()

	m.logger.Log("stream", "refresh status started")
	var b strings.Builder
	if m.connected {
		b.WriteString(fmt.Sprintf("接続中: %s (@%s) [%s]\n",
			color.GreenString(m.instance.BaseUrl),
			color.CyanString(m.instance.UserName),
			m.timeline))
	} else {
		b.WriteString(fmt.Sprintf("切断: %s [ - ]\n",
			color.RedString(m.instance.BaseUrl)))
	}

	if m.err != nil {
		b.WriteString(fmt.Sprintf("エラー: %s\n", color.RedString(m.err.Error())))
	}

	// ヘルプ表示
	b.WriteString("--------------------------------\n")
	b.WriteString("[ctrl+h] ホームTL [ctrl+l] ローカルTL [ctrl+c] 終了\n")
	b.WriteString("--------------------------------\n\n")

	m.viewStatus.SetContent(b.String())
	m.logger.Log("stream", "refresh status finished")
}

func (m *Model) refreshViewBuffer() {
	m.muViewAll.Lock()
	defer m.muViewAll.Unlock()

	m.logger.Log("stream", "refresh started")
	m.refreshStatusView()

	m.viewBuffer.Reset()

	maxNotes := m.height - 6
	if maxNotes < 0 {
		maxNotes = 10
	}

	startIdx := len(m.notes) - maxNotes
	if startIdx < 0 {
		startIdx = 0
	}

	for i := startIdx; i < len(m.notes); i++ {
		note := m.notes[i]
		m.viewBuffer.WriteString(formatNote(note))
		m.viewBuffer.WriteString("\n")
	}

	// m.viewMain = bubbles.NewViewportFactory().StreamView()
	m.viewMain.SetContent(m.viewBuffer.String())
	m.logger.Log("stream", fmt.Sprintf("view buffer: %s", m.viewBuffer.String()))

	m.logger.Log("stream", "refresh finished")
}

// PostnoteCallback は投稿ノートのコールバック関数です
func (m *Model) postnoteCallback(content string) tea.Cmd {

	ret, err := m.apiClient.CreateNote(context.Background(), misskey.VisibilityHome, content)
	if err != nil {
		m.logger.Log("stream", fmt.Sprintf("note error: %v", err))
		return nil
	}
	m.logger.Log("stream", fmt.Sprintf("note: %s", ret.CreatedNote.Body.ID))

	return nil
}

// formatNote はノートを表示用にフォーマットします
func formatNote(note *misskey.Note) string {
	var buf strings.Builder
	var data map[string]interface{}
	if note.Body.Body.RenoteID != "" {
		t, err := template.New("note").Parse(RenoteTmpl)
		if err != nil {
			log.Printf("template error: %v", err)
		}
		data = map[string]interface{}{
			"renotedName":     color.HiBlackString(note.Body.Body.User.Name),
			"renotedUsername": color.HiBlackString(note.Body.Body.User.Username),
			"name":            color.HiGreenString(note.Body.Body.Renote.User.Name),
			"username":        color.HiBlueString(note.Body.Body.Renote.User.Username),
			"text":            note.Body.Body.Renote.Text,
			"createdAt":       note.Body.Body.Renote.CreatedAt.Format(time.RFC3339),
		}
		if err := t.Execute(&buf, data); err != nil {
			log.Printf("template execute error: %v", err)
		}
	} else {
		t, err := template.New("note").Parse(NoteTmpl)
		if err != nil {
			log.Printf("template error: %v", err)
		}
		data = map[string]interface{}{
			"name":      color.HiGreenString(note.Body.Body.User.Name),
			"username":  color.HiBlueString(note.Body.Body.User.Username),
			"text":      note.Body.Body.Text,
			"createdAt": note.Body.Body.CreatedAt.String(),
		}
		if err := t.Execute(&buf, data); err != nil {
			log.Printf("template execute error: %v", err)
		}
	}
	return buf.String()
}
