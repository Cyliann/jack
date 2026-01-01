package main

import (
	"log"

	"github.com/Cyliann/jack/tui"
	tea "github.com/charmbracelet/bubbletea"
)

var core *tea.Program

func main() {
	core = tea.NewProgram(tui.NewModel())
	_, err := core.Run()
	if err != nil {
		log.Fatalf("error running program: %v\n", err)
	}
}
