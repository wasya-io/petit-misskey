package meta

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wasya-io/petit-misskey/domain/view"
	"github.com/wasya-io/petit-misskey/infrastructure/bubbles"
	"github.com/wasya-io/petit-misskey/service/meta"
	"github.com/wasya-io/petit-misskey/util"
)

type (
	Model struct {
		view       view.SimpleView
		service    *meta.Service
		ctx        context.Context
		quitting   bool
		teaProgram *tea.Program
	}
	initMsg tea.Msg // Updateに起動処理を要求するMsg
)

func NewModel(service *meta.Service, viewFactory bubbles.SimpleViewFactory) *Model {
	return &Model{
		view:     viewFactory.View(),
		service:  service,
		ctx:      context.Background(),
		quitting: false,
	}
}

func (m *Model) Init() tea.Cmd {
	return func() tea.Msg { return new(initMsg) } // 起動処理要求を返す
}

// Update is called when a message is received. Use it to inspect messages
// and, in response, update the model and/or send a command.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg: // 終了コマンドの割り込みを処理
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		default:
			return m, nil
		}

	case initMsg: // 起動処理
		j, err := m.service.Do(m.ctx)
		if err != nil {
			// TODO: viewにエラーメッセージを詰めて返す
			return m, nil
		}

		m.view.SetContent(util.PrittyJson(j)) // metaの実行結果をviewに渡す

		return m, nil

	default:
		return m, nil
	}
}

// View renders the program's UI, which is just a string. The view is
// rendered after every Update.
func (m *Model) View() string {
	view := m.view.View()
	if m.quitting {
		view += "\n" // NOTE: 終了時に最後の行がつぶれないようにする
	}
	return view
}

func (m *Model) MsgChannel() chan tea.Msg {
	return nil
}
