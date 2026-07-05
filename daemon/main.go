package main

import (
	"context"
	"log"
	"mfeeder/internal/config"
	"mfeeder/internal/shutdown"
	"mfeeder/internal/sqlite"
	"mfeeder/internal/watcher/core"
	"mfeeder/internal/watcher/factory"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	logFile, err := setupLogger()
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	cfg, err := config.LoadConfig(true)
	if err != nil {
		log.Fatal(err)
	}

	watcher := factory.NewWatcher(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	sdManager := shutdown.Manager{Cancel: cancel}

	db, err := sqlite.Init()
	if err != nil {
		log.Fatal(err)
	}
	err = sqlite.StartupClean(ctx, db)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err = sqlite.CloseAll(ctx, db)
		if err != nil {
			log.Println(err)
		}
		db.Close()
	}()

	wArr, err := watcher.Snapshot(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// todo: dichiarare a priori le finestre aperte potrebbe non essere una buona idea
	// potenzialmente una finestra minimizzata per molte ore risulterebbe aperta se non arrivano eventi che la riguardano
	for _, w := range wArr {
		if w.Focused {
			err = sqlite.WindowFocused(ctx, w, db)
			if err != nil {
				log.Println(err)
			}
			continue
		}
		err = sqlite.WindowOpened(ctx, w, db)
		if err != nil {
			log.Println(err)
		}
	}

	ch, err := watcher.Watch(&sdManager)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Watcher started")

	for {
		select {
		case <-ctx.Done():
			watcher.Close(&sdManager)
			return
		case event := <-ch:
			switch event.WindowEvent {
			case core.WindowOpened:
				err = sqlite.WindowOpened(ctx, event.Window, db)
			case core.WindowClosed:
				err = sqlite.WindowClosed(ctx, event.Window, db)
			case core.WindowFocused:
				err = sqlite.WindowFocused(ctx, event.Window, db)
			}
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func setupLogger() (*os.File, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	logDir := filepath.Join(dir, "mfeeder", "logs")
	if err = os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	if err = cleanupOldLogs(logDir, 7); err != nil {
		return nil, err
	}

	logFile := filepath.Join(logDir, "mfeederd-"+time.Now().Format("2006-01-02")+".log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	return f, nil
}

func cleanupOldLogs(logDir string, retentionDays int) error {
	if retentionDays <= 0 {
		return nil
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	entries, err := os.ReadDir(logDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, "mfeederd-") || !strings.HasSuffix(name, ".log") {
			continue
		}

		datePart := strings.TrimSuffix(strings.TrimPrefix(name, "mfeederd-"), ".log")
		logDate, err := time.Parse("2006-01-02", datePart)
		if err != nil {
			continue
		}

		if logDate.Before(cutoff) {
			if err := os.Remove(filepath.Join(logDir, name)); err != nil {
				return err
			}
		}
	}

	return nil
}
