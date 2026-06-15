package core

import (
	"context"
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
	Watch(ctx context.Context) (<-chan WindowEvent, error)
	Close() error
}
