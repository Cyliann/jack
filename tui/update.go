package tui

import (
	"context"
	"errors"
	"fmt"
	"time"

	api "github.com/Cyliann/jack/ytmusicapi"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lrstanley/go-ytdlp"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	// handle window resizing
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	// handle keypresses
	case tea.KeyMsg:
		m, cmd = m.handleKeyboardInput(msg)
		return m, cmd

	// all tools are present and verified
	case MsgToolsVerified:
		if msg.Error != nil {
			return printErrorAndExit(m, msg, "error installing/verifying tools")
		}
		m, cmd = m.proceedToUserInput()
		return m, cmd

	// got the queried item from innertube
	case MsgFound:
		if msg.Error != nil {
			return printErrorAndExit(m, msg, "error searching for items")
		}

		m.items = msg.Items
		m.state = StateDownloading

		// start downloading and listen for progress
		return m, tea.Batch(m.initiateDownload, listenForProgress(m.progressChan))

	// got progress from yt-dlp
	case MsgProgress:
		return updateDownloadProgress(m, msg)

	// finished downloading
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

	// update textinput only if it's visible
	if m.state == StateUserInput {
		var cmd tea.Cmd

		m.ti, cmd = m.ti.Update(msg)
		return m, cmd
	}
	return m, nil
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

func (m model) handleKeyboardInput(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEscape:
		return m, tea.Quit
	case tea.KeyEnter:
		if m.state == StateUserInput {
			m.state = StateDownloading
			return m, m.search
		}
	}
	var cmd tea.Cmd
	m.ti, cmd = m.ti.Update(msg)
	return m, cmd
}

func (m model) proceedToUserInput() (model, tea.Cmd) {
	m.ti.Focus()
	m.state = StateUserInput
	return m, textinput.Blink
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

func updateDownloadProgress(m model, msg MsgProgress) (model, tea.Cmd) {
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
