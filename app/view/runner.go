package view

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wasya-io/petit-misskey/domain/core"
)

type Model interface {
	tea.Model
	MsgChannel() chan tea.Msg
}

func Run(model Model, logger core.Logger) {
	p := tea.NewProgram(model, tea.WithAltScreen())
	msgCh := model.MsgChannel()

	go func() {
		for {
			logger.Log("runner", "wait message.")
			p.Send(<-msgCh)
			logger.Log("runner", "message received.")
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Printf("tea runner error. %v", err) // TODO: ちゃんとエラー処理
		os.Exit(1)
	}
}
