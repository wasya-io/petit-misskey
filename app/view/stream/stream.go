package stream

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/wasya-io/petit-misskey/infrastructure/setting"
	"github.com/wasya-io/petit-misskey/infrastructure/websocket"
	"github.com/wasya-io/petit-misskey/model/misskey"
)

type Model struct {
	ctx        context.Context
	cancel     context.CancelFunc
	msgCh      chan tea.Msg
	client     websocket.Client
	notes      []*misskey.Note
	quitting   bool
	connected  bool
	err        error
	instance   *setting.Instance
	viewBuffer strings.Builder
	width      int
	height     int
}

func NewModel(instance *setting.Instance, client websocket.Client, msgCh chan tea.Msg) *Model {
	ctx, cancel := context.WithCancel(context.Background())

	return &Model{
		ctx:        ctx,
		cancel:     cancel,
		msgCh:      msgCh,
		client:     client,
		notes:      make([]*misskey.Note, 0, 100),
		quitting:   false,
		connected:  false,
		instance:   instance,
		viewBuffer: strings.Builder{},
		width:      80,
		height:     20,
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			m.client.Stop()
			return m, tea.Quit
		}

	case websocket.NoteMessage:
		m.notes = append(m.notes, msg.Note)
		if len(m.notes) > 100 {
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

	case websocket.WebSocketDisconnectedMsg:
		m.connected = false
		if msg.Err != nil {
			m.err = msg.Err
		}
		m.refreshViewBuffer()
		return m, nil

	case websocket.WebSocketErrorMsg:
		m.err = msg.Err
		return m, nil

	}
	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return "Goodbye!"
	}

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

	// キャッシュされたビューを表示
	b.WriteString(m.viewBuffer.String())

	return b.String()
}

func (m *Model) MsgChannel() chan tea.Msg {
	return m.msgCh
}

func (m *Model) refreshViewBuffer() {
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
}

// formatNote はノートを表示用にフォーマットします
func formatNote(note *misskey.Note) string {
	var b strings.Builder

	if note.Body.Body.RenoteID != "" {
		// リノートの場合
		b.WriteString(fmt.Sprintf("%s (@%s) がリノート:\n",
			color.HiBlackString(note.Body.Body.User.Name),
			color.HiBlackString(note.Body.Body.User.Username)))

		b.WriteString(fmt.Sprintf("  %s (@%s): %s\n",
			color.HiGreenString(note.Body.Body.Renote.User.Name),
			color.HiBlueString(note.Body.Body.Renote.User.Username),
			note.Body.Body.Renote.Text))

		b.WriteString(fmt.Sprintf("  %s\n",
			color.HiBlackString(note.Body.Body.Renote.CreatedAt.Format(time.RFC3339))))
	} else {
		// 通常投稿の場合
		b.WriteString(fmt.Sprintf("%s (@%s): %s\n",
			color.HiGreenString(note.Body.Body.User.Name),
			color.HiBlueString(note.Body.Body.User.Username),
			note.Body.Body.Text))

		b.WriteString(fmt.Sprintf("  %s\n",
			color.HiBlackString(note.Body.Body.CreatedAt.Format(time.RFC3339))))
	}

	return b.String()
}
