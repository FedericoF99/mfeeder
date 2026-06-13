package watcher

import (
	"mfeeder/internal/config"
	"testing"
)

func TestWatcher(t *testing.T) {
	watcher := NewWatcher()
	watcher.Snapshot(nil, &config.Conf{})
}
