package postnote

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wasya-io/petit-misskey/domain/core"
)

type (
	PostTextarea struct {
		textarea.Model
		logger         core.Logger
		CallbackSubmit func(content string) tea.Cmd
	}

	PostKeyMap struct {
		textarea.KeyMap
		Submit key.Binding
	}
)

// 新しいPostKeyMapを生成します
func NewPostKeyMap() PostKeyMap {
	return PostKeyMap{
		Submit: key.NewBinding(
			key.WithKeys("ctrl+s", "cmd+s"),
			key.WithHelp("Ctrl+s/Cmd+s", "送信"),
		),
	}
}

// NewPostTextarea は新しいPostTextareaを生成します
func NewPostTextarea(submitCallback func(content string) tea.Cmd, logger core.Logger) PostTextarea {
	ta := textarea.New()
	placeholder := "いまどうしてる？"
	ta.Placeholder = " " + placeholder
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 280
	ta.SetWidth(lipgloss.Width(placeholder) + 8)
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	return PostTextarea{
		Model:          ta,
		logger:         logger,
		CallbackSubmit: submitCallback,
	}
}

// PostKeyMap は現在のキーマップを返します
func (pt PostTextarea) PostKeyMap() PostKeyMap {
	return NewPostKeyMap()
}

// Update はキーイベントを処理します
func (pt PostTextarea) Update(msg tea.Msg) (PostTextarea, tea.Cmd) {
	var cmds []tea.Cmd
	pt.logger.Log("postarea", fmt.Sprintf("updated %T", msg))
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Command+Enterが押されたかチェック
		// キー情報をより詳細にログ出力する例
		pt.logger.Log("postarea", fmt.Sprintf("key message: %s, Alt: %v", msg.String(), msg.Alt))
		if key.Matches(msg, pt.PostKeyMap().Submit) {
			content := pt.Value()
			pt.logger.Log("postarea", fmt.Sprintf("content: %s", content))
			if content != "" && pt.CallbackSubmit != nil {
				cmds = append(cmds, pt.CallbackSubmit(content))
				// 入力をクリア
				pt.Reset()
				return pt, tea.Batch(cmds...)
			}
		}
	default:
		pt.logger.Log("postarea", fmt.Sprintf("key message??: %v", msg))
	}

	// 標準のtextarea更新処理を実行
	mdl, cmd := pt.Model.Update(msg)
	pt.Model = mdl
	cmds = append(cmds, cmd)

	return pt, tea.Batch(cmds...)
}
