//go:build !linux && !windows

package factory

import (
	"log"
	"mfeeder/internal/config"
	"mfeeder/internal/watcher/core"
)

func newWatcher(c *config.Conf) core.Watcher {
	log.Fatal("Unsupported platform")
}
