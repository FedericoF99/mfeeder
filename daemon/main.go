package main

import (
	"context"
	"log"
	"mfeeder/internal/config"
	"mfeeder/internal/shutdown"
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

	ctx, cancel := context.WithCancel(context.Background())
	sdManager := shutdown.Manager{Cancel: cancel}

	db, err := sqlite.Init(ctx)
	if err != nil {
		log.Fatal(err)
	}
	err = sqlite.StartupClean(ctx, db)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		_ = sqlite.CloseAll(ctx, db)
		db.Close()
	}()

	wArr, err := watcher.Snapshot(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, w := range wArr {
		if w.Focused {
			_ = sqlite.WindowFocused(ctx, w, db)
			continue
		}
		_ = sqlite.WindowOpened(ctx, w, db)
	}

	ch, err := watcher.Watch(&sdManager)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-ctx.Done():
			watcher.Close(&sdManager)
			return
		case event := <-ch:
			println(event.WindowEvent, event.Window.Title)
			switch event.WindowEvent {
			case core.WindowOpened:
				_ = sqlite.WindowOpened(ctx, event.Window, db)
			case core.WindowClosed:
				_ = sqlite.WindowClosed(ctx, event.Window, db)
			case core.WindowFocused:
				_ = sqlite.WindowFocused(ctx, event.Window, db)
			}
		}
	}
}
