package bubbles

import (
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/wire"
	"github.com/wasya-io/petit-misskey/domain/view"
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
	vp := viewport.New(120, 7)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)

	return &vp
}
