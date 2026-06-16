package factory

import (
	"mfeeder/internal/config"
	"mfeeder/internal/watcher/core"
)

func NewWatcher(c *config.Conf) core.Watcher {
	return newWatcher(c)
}
