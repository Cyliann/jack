package tui

import "github.com/lrstanley/go-ytdlp"

type AppState int

const (
	StateInstallingDependencies AppState = iota
	StateUserInput
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
