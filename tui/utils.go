package tui

import (
	"os"
	"regexp"

	tea "github.com/charmbracelet/bubbletea"
)

func containstString(str, filepath string) bool {
	b, err := os.ReadFile(filepath)
	if err != nil {
		panic(err)
	}

	isExist, err := regexp.Match(str, b)
	if err != nil {
		panic(err)
	}
	return isExist
}

func listenForProgress(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}
