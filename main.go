// example from: https://github.com/lrstanley/go-ytdlp/blob/master/_examples/bubble-dl/main.go
package main

import (
	"fmt"
	"os"

	"github.com/Cyliann/jack/tui"
	tea "github.com/charmbracelet/bubbletea"
)

var core *tea.Program

func main() {
	core = tea.NewProgram(tui.NewModel())
	_, err := core.Run()
	if err != nil {
		fmt.Printf("error running program: %v\n", err) //nolint:forbidigo
		os.Exit(1)
	}
}
