package main

import (
	"context"
	"log"
	"mfeeder/internal/config"
	"mfeeder/internal/sqlite"
	"mfeeder/internal/watcher/core"
	"mfeeder/internal/watcher/factory"
)

func main() {
	cfg, err := config.LoadConfig(true)
	if err != nil {
		log.Fatal(err)
	}

	watcher := factory.NewWatcher(cfg)

	ctx, _ := context.WithCancel(context.Background())
	_, err = watcher.Snapshot(ctx)
	if err != nil {
		log.Fatal(err)
	}

	ch, err := watcher.Watch(ctx)
	if err != nil {
		return
	}

	for {
		select {
		case event := <-ch:
			switch event.WindowEvent {
			case core.WindowOpened:
				sqlite.WindowOpened(event.Window)
			case core.WindowClosed:
				sqlite.WindowClosed(event.Window)
			case core.WindowFocused:
				sqlite.WindowFocused(event.Window)
			}
		}
	}
}
