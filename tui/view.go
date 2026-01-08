package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
)

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
		return getDownloadView(m)
	}
	return ""
}

func getDownloadView(m model) string {

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
