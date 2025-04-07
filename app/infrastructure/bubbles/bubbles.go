package bubbles

import (
	"os"
	"syscall"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/wire"
	"github.com/wasya-io/petit-misskey/domain/core"
	"github.com/wasya-io/petit-misskey/domain/view"
	"github.com/wasya-io/petit-misskey/view/postnote"
	"golang.org/x/term"
)

type (
	SimpleViewFactory interface {
		View() view.SimpleView
	}
	ViewportFactory struct{}
)

var ProviderSet = wire.NewSet(
	NewViewportFactory,
)

func NewViewportFactory() *ViewportFactory {
	return &ViewportFactory{}
}

func (f *ViewportFactory) View() view.SimpleView {
	// ターミナルのサイズを取得
	width, height := getTerminalSize()

	// ターミナルサイズの80%を使用し、最小値を設定
	viewportWidth := max(80, width*80/100)
	viewportHeight := max(5, height*70/100)

	vp := viewport.New(viewportWidth, viewportHeight)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)

	return &vp
}

func (f *ViewportFactory) StreamView() viewport.Model {
	// ターミナルのサイズを取得
	width, height := getTerminalSize()

	// ターミナルサイズの80%を使用し、最小値を設定
	viewportWidth := max(80, width*80/100)
	viewportHeight := max(5, height*70/100)

	vp := viewport.New(120, viewportHeight)
	// ビューポートのサイズを固定化
	vp.Width = viewportWidth
	vp.Height = viewportHeight
	vp.Style = lipgloss.NewStyle().Width(120).
		PaddingRight(2)

	return vp
}

func (f *ViewportFactory) SystemView() viewport.Model {
	// ターミナルのサイズを取得
	width, _ := getTerminalSize()

	// ターミナルサイズの80%を使用し、最小値を設定
	viewportWidth := max(180, width*80/100)
	viewportHeight := 4

	vp := viewport.New(viewportWidth, viewportHeight)
	vp.Style = lipgloss.NewStyle().Width(180).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)

	return vp
}

func (f *ViewportFactory) PostView(submitCallback func(content string) tea.Cmd, logger core.Logger) postnote.PostTextarea {
	return postnote.NewPostTextarea(submitCallback, logger)
}

// getTerminalSize は現在のターミナルの幅と高さを返します
// エラーが発生した場合はデフォルト値(120, 30)を返します
func getTerminalSize() (width, height int) {
	defaultWidth, defaultHeight := 180, 30

	// 標準出力のファイルディスクリプタからターミナルのサイズを取得
	if w, h, err := term.GetSize(int(syscall.Stdout)); err == nil {
		return w, h
	}

	// 標準入力からも試行
	if w, h, err := term.GetSize(int(os.Stdin.Fd())); err == nil {
		return w, h
	}

	return defaultWidth, defaultHeight
}

// max は2つの整数のうち大きい方を返します
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
