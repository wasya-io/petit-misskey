package view

type (
	SimpleView interface {
		View() string
		SetContent(s string)
	}
)
