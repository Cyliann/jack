package tui

import "github.com/lrstanley/go-ytdlp"

type AppState int

const (
	StateInstallingDependencies AppState = iota
	StateUserInput
	StateSearching
	StateDownloading
)

type MsgProgress struct {
	Progress ytdlp.ProgressUpdate
}

type MsgFinished struct {
	Result *ytdlp.Result
	Error  error
}

func (m MsgFinished) Err() error { return m.Error }

type MsgToolsVerified struct {
	Resolved []*ytdlp.ResolvedInstall
	Error    error
}

func (m MsgToolsVerified) Err() error { return m.Error }

type MsgFound struct {
	Items []Item
	Error error
}

func (m MsgFound) Err() error { return m.Error }

type Item struct {
	Title  string
	Artist string
	Url    string
}
