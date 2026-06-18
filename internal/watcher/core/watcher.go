package core

import (
	"context"
	"mfeeder/internal/shutdown"
)

type Window struct {
	Pid     int
	Exe     string
	Title   string
	Focused bool
}

type WindowEventType string

const (
	WindowOpened  WindowEventType = "WINDOW_OPENED"
	WindowClosed  WindowEventType = "WINDOW_CLOSED"
	WindowFocused WindowEventType = "WINDOW_FOCUSED"
)

type WindowEvent struct {
	Window      Window
	WindowEvent WindowEventType
}

type Watcher interface {
	Snapshot(ctx context.Context) ([]Window, error)
	Watch(sdManager *shutdown.Manager) (<-chan WindowEvent, error)
	Close(sdManager *shutdown.Manager)
}
