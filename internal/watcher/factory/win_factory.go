//go:build windows

package factory

import (
	"mfeeder/internal/config"
	"mfeeder/internal/watcher/core"
)

func newWatcher(c *config.Conf) core.Watcher {
	return WinWatcher{cfg: c}
}
