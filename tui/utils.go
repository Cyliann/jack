package tui

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"runtime"

	api "github.com/Cyliann/jack/ytmusicapi"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lrstanley/go-ytdlp"
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

// If yt-dlp/ffmpeg/ffprobe isn't installed yet, download and cache the binaries for further use.
// Note that the download/installation of ffmpeg/ffprobe is only supported on a handful of platforms,
// and so it is still recommended to install ffmpeg/ffprobe via other means.
func installDeps() tea.Msg {
	// FIX: automatic dependency install
	// for some reason InstallAll() doesn't pull the newest yt-dlp binary
	return MsgToolsVerified{nil, nil}
	// On NixOS you cannot run dynamically linked binaries, so we can't install yt-dlp from github.
	// We rely on it being installed externally and only install ffmpeg

	// Check if the system is NixOs
	if runtime.GOOS == "linux" && containstString("nixos", "/etc/os-release") {
		// Install only FFmpeg
		resolved, err := ytdlp.InstallFFmpeg(context.TODO(), nil)
		return MsgToolsVerified{Resolved: []*ytdlp.ResolvedInstall{resolved}, Error: err}

	} else {
		// Otherwise install all tools
		resolved, err := ytdlp.InstallAll(context.TODO())
		return MsgToolsVerified{Resolved: resolved, Error: err}
	}
}

type hasErr interface {
	Err() error
}

func printErrorAndExit[T hasErr](m model, msg T, text string) (model, tea.Cmd) {
	return m, tea.Sequence(
		tea.Printf("%s %s: %s", errorStyle, text, msg.Err()),
		tea.Quit,
	)
}

func createItemFromAlbum(albumResult api.AlbumResult) Item {
	id := albumResult.Id

	url := fmt.Sprintf("https://music.youtube.com/playlist?list=%s", id)

	return Item{Title: albumResult.Album, Artist: albumResult.Artist, Url: url}
}
