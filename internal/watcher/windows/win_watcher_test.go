//go:build windows

package windows

import (
	"context"
	"mfeeder/internal/config"
	"testing"
)

func TestWatcher(t *testing.T) {
	watcher := NewWatcher(&config.Conf{})
	watcher.Watch(context.Background())
}
