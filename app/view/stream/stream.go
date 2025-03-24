package stream

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"strings"
	"text/template"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/wasya-io/petit-misskey/domain/core"
	"github.com/wasya-io/petit-misskey/infrastructure/bubbles"
	"github.com/wasya-io/petit-misskey/infrastructure/setting"
	"github.com/wasya-io/petit-misskey/infrastructure/websocket"
	"github.com/wasya-io/petit-misskey/model/misskey"
)

type Model struct {
	ctx         context.Context
	logger      core.Logger
	cancel      context.CancelFunc
	msgCh       chan tea.Msg
	viewMain    viewport.Model
	viewFooter  viewport.Model
	client      websocket.Client
	notes       []*misskey.Note
	quitting    bool
	connected   bool
	err         error
	instance    *setting.Instance
	viewBuffer  strings.Builder
	width       int
	height      int
	initialized bool
}

var (
	//go:embed template/note.tmpl
	NoteTmpl string
	//go:embed template/renote.tmpl
	RenoteTmpl string
)

func NewModel(instance *setting.Instance, client websocket.Client, logger core.Logger, msgCh chan tea.Msg) *Model {
	ctx, cancel := context.WithCancel(context.Background())

	return &Model{
		ctx:         ctx,
		logger:      logger,
		cancel:      cancel,
		msgCh:       msgCh,
		viewMain:    bubbles.NewViewportFactory().StreamView(),
		viewFooter:  bubbles.NewViewportFactory().FooterView(),
		client:      client,
		notes:       make([]*misskey.Note, 0, 100),
		quitting:    false,
		connected:   false,
		instance:    instance,
		viewBuffer:  strings.Builder{},
		width:       120,
		height:      20,
		initialized: false,
	}
}

func (m Model) Init() tea.Cmd {
	// WebSocketクライアントのgoroutine起動コマンドを返す
	return func() tea.Msg {
		// 別goroutineでWebSocket接続を開始
		go func() {
			if err := m.client.Start(); err != nil {
				m.msgCh <- tea.Msg(websocket.WebSocketErrorMsg{Err: err})
			}
		}()

		// 初期化完了を通知
		return nil
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.logger.Log("stream", fmt.Sprintf("msg: %T", msg))
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			m.client.Stop()
			return m, tea.Quit
		}

	case websocket.NoteMessage:
		m.notes = append([]*misskey.Note{msg.Note}, m.notes...)
		if len(m.notes) > 10 {
			m.notes = m.notes[1:]
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
		m.viewFooter.SetContent("")
		m.viewMain.SetContent("Disconnected")
		return m, nil

	case websocket.WebSocketErrorMsg:
		m.err = msg.Err
		m.viewFooter.SetContent("")
		m.viewMain.SetContent(msg.Err.Error())
		return m, nil

	}
	return m, nil
}

func (m Model) View() string {
	m.viewMain.Width = m.width
	m.viewFooter.Width = m.width

	joinedView := lipgloss.JoinVertical(
		lipgloss.Left,
		m.viewMain.View(),
		m.viewFooter.View(),
	)
	return joinedView

	// if m.quitting {
	// 	return "Goodbye!"
	// }

	// var b strings.Builder

	// if m.connected {
	// 	b.WriteString(fmt.Sprintf("接続中: %s (@%s)\n",
	// 		color.GreenString(m.instance.BaseUrl),
	// 		color.CyanString(m.instance.UserName)))
	// } else {
	// 	b.WriteString(fmt.Sprintf("切断: %s\n",
	// 		color.RedString(m.instance.BaseUrl)))
	// }

	// if m.err != nil {
	// 	b.WriteString(fmt.Sprintf("エラー: %s\n", color.RedString(m.err.Error())))
	// }

	// // ヘルプ表示
	// b.WriteString("--------------------------------\n")
	// b.WriteString("[h] ホームTL [l] ローカルTL [q] 終了\n")
	// b.WriteString("--------------------------------\n\n")

	// // キャッシュされたビューを表示
	// b.WriteString(m.viewBuffer.String())

	// return b.String()
}

func (m *Model) MsgChannel() chan tea.Msg {
	return m.msgCh
}

func (m *Model) refreshViewBuffer() {

	var b strings.Builder
	if m.connected {
		b.WriteString(fmt.Sprintf("接続中: %s (@%s)\n",
			color.GreenString(m.instance.BaseUrl),
			color.CyanString(m.instance.UserName)))
	} else {
		b.WriteString(fmt.Sprintf("切断: %s\n",
			color.RedString(m.instance.BaseUrl)))
	}

	if m.err != nil {
		b.WriteString(fmt.Sprintf("エラー: %s\n", color.RedString(m.err.Error())))
	}

	// ヘルプ表示
	b.WriteString("--------------------------------\n")
	b.WriteString("[h] ホームTL [l] ローカルTL [q] 終了\n")
	b.WriteString("--------------------------------\n\n")

	m.viewFooter.SetContent(b.String())

	m.viewBuffer.Reset()

	m.viewBuffer.WriteString("--------------------------------\n")

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

	m.viewMain = bubbles.NewViewportFactory().StreamView()
	m.viewMain.SetContent(m.viewBuffer.String())
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
