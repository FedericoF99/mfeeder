package main

import (
	"context"
	"fmt"
	"log"
	"mfeeder/internal/config"
	"mfeeder/internal/watcher/factory"
)

func main() {
	cfg, err := config.LoadConfig(true)
	if err != nil {
		log.Fatal(err)
	}

	watcher := factory.NewWatcher(cfg)

	ctx := context.Background()
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
			fmt.Printf("%s - %s\n", event.Window.Exe, event.WindowEvent)
		}
	}
}
