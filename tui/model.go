package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	api "github.com/Cyliann/jack/ytmusicapi"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// handle window resizing
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	// handle keypresses
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEscape:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.state == StateUserInput {
				m.state = StateDownloading
				return m, m.search
			}
		}

	// if all tools are present and verified
	case MsgToolsVerified:
		if msg.Error != nil {
			return printErrorAndExit(m, msg, "error installing/verifying tools")
		}
		m.ti.Focus()
		m.state = StateUserInput
		return m, textinput.Blink

	case MsgFound:
		if msg.Error != nil {
			return printErrorAndExit(m, msg, "error searching for items")
		}

		m.items = msg.Items

		// start downloading and listen for progress
		return m, tea.Batch(m.initiateDownload, listenForProgress(m.progressChan))

	case MsgProgress:
		return handleProgress(m, msg)

	case MsgFinished:
		m.done = true
		if msg.Error != nil {
			return printErrorAndExit(m, msg, "error downloading items")
		}
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.progress = newModel
		}
		return m, cmd
	}

	if m.state == StateUserInput {
		var cmd tea.Cmd

		m.ti, cmd = m.ti.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	switch m.state {
	case StateInstallingDependencies:
		return doneStyle.Render(m.spinner.View() + " installing dependencies...\n")

	case StateUserInput:
		return fmt.Sprintf(
			"%s\n\n%s",
			m.ti.View(),
			"(esc to quit)",
		)

	case StateDownloading:
		return handleDownloadingView(m)
	}
	return ""
}

func (m model) initiateDownload() tea.Msg {
	dl := ytdlp.New().
		Format("bestaudio").
		ProgressFunc(100*time.Millisecond, func(prog ytdlp.ProgressUpdate) {
			m.progressChan <- MsgProgress{Progress: prog}
		}).
		RemuxVideo("opus").
		Output("%(artists.0)s - %(title)s.%(ext)s")

	var urls []string
	for _, item := range m.items {
		urls = append(urls, item.Url)
	}

	result, err := dl.Run(context.TODO(), urls...)

	return MsgFinished{Result: result, Error: err}
}

func (m model) search() tea.Msg {
	query := m.ti.Value()
	if query == "" {
		return MsgFound{nil, errors.New("No query provided. Aborting.")}
	}

	albums, err := api.Search(query)
	if err != nil {
		return MsgFound{nil, err}
	}

	items := []Item{}
	for i := range 1 { // TODO: Support downloading multiple items at once
		item := createItemFromAlbum(albums[i])
		items = append(items, item)
	}

	return MsgFound{items, nil}
}

func handleProgress(m model, msg MsgProgress) (model, tea.Cmd) {
	m.lastProgress = msg.Progress
	cmds := []tea.Cmd{m.progress.SetPercent(msg.Progress.Percent() / 100)}

	if m.lastProgress.Status == ytdlp.ProgressStatusFinished {
		cmds = append(cmds, tea.Printf(
			"%s downloaded %s (%s)",
			successStyle,
			fileStyle.Render(fmt.Sprintf("%s - %s", *m.lastProgress.Info.Artist, *m.lastProgress.Info.Title)),
			titleStyle.Render(*m.lastProgress.Info.Filename),
		))
	}
	if m.lastProgress.Status == ytdlp.ProgressStatusError {
		cmds = append(cmds, tea.Printf(
			"%s error downloading: %s",
			errorStyle,
			*m.lastProgress.Info.URL,
		))
	}

	// listenForProgress() exits with a msg every time there's progress
	// so we have to call it again every time.
	// feels like a hack, but I don't know a better way to do it
	cmds = append(cmds, listenForProgress(m.progressChan))

	return m, tea.Sequence(cmds...)
}

func handleDownloadingView(m model) string {

	if m.lastProgress.Status == "" {
		return doneStyle.Render(m.spinner.View() + " searching...\n")
	}

	if m.done {
		return doneStyle.Render(fmt.Sprintf("downloaded %d items.\n", len(m.items)))
	}

	// " <spinner> <status> <file> <GAP> <progress> [eta: <eta>] [size: <size>]"
	spin := m.spinner.View()
	status := string(m.lastProgress.Status)
	prog := m.progress.View()
	eta := m.lastProgress.ETA().Round(time.Second).String()
	eta = "[eta: " + etaStyle.MarginLeft(max(0, 4-len(eta))).Render(eta) + "]"
	size := "[size: " + sizeStyle.Render(humanize.Bytes(uint64(m.lastProgress.TotalBytes))) + "]"

	cellsAvail := max(0, m.width-lipgloss.Width(spin+" "+status+" "+prog+" "+eta+" "+size))

	file := fileStyle.MaxWidth(cellsAvail).Render(m.lastProgress.Filename)

	cellsRemaining := max(0, cellsAvail-lipgloss.Width(file))
	gap := strings.Repeat(" ", cellsRemaining)

	return spin + " " + status + " " + file + gap + prog + " " + eta + " " + size
}
