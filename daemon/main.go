package main

import (
	"context"
	"log"
	"mfeeder/internal/config"
	"mfeeder/internal/shutdown"
	"mfeeder/internal/watcher/factory"
)

func main() {
	cfg, err := config.LoadConfig(true)
	if err != nil {
		log.Fatal(err)
	}

	watcher := factory.NewWatcher(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	_, err = watcher.Snapshot(ctx)
	if err != nil {
		log.Fatal(err)
	}

	sdManager := shutdown.Manager{Cancel: cancel}

	ch, err := watcher.Watch(&sdManager)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		// todo: save state
	}()

	for {
		select {
		case <-ctx.Done():
			watcher.Close(&sdManager)
			return
		case event := <-ch:
			println(event.WindowEvent, event.Window.Title)
			switch event.WindowEvent {
			//case core.WindowOpened:
			//	sqlite.WindowOpened(event.Window)
			//case core.WindowClosed:
			//	sqlite.WindowClosed(event.Window)
			//case core.WindowFocused:
			//	sqlite.WindowFocused(event.Window)
			}
		}
	}
}
