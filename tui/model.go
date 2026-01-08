package tui

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lrstanley/go-ytdlp"
)

type model struct {
	items        []Item
	width        int
	height       int
	spinner      spinner.Model
	progress     progress.Model
	done         bool
	progressChan chan tea.Msg
	lastProgress ytdlp.ProgressUpdate
	state        AppState
	ti           textinput.Model
}

func NewModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter your search query"
	ti.Width = 100

	m := model{
		spinner: spinner.New(),
		progress: progress.New(
			progress.WithDefaultGradient(),
			progress.WithWidth(40),
		),
		progressChan: make(chan tea.Msg),
		state:        StateInstallingDependencies,
		ti:           ti,
	}

	m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		installDeps,
		m.spinner.Tick,
	)
}
