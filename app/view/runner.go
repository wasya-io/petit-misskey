package view

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type Model interface {
	tea.Model
	MsgChannel() chan tea.Msg
}

func Run(model Model) {
	p := tea.NewProgram(model)
	msgCh := model.MsgChannel()

	go func() {
		for {
			fmt.Println("wait message.")
			p.Send(<-msgCh)
			fmt.Println("message received.")
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Printf("tea runner error. %v", err) // TODO: ちゃんとエラー処理
		os.Exit(1)
	}
}
