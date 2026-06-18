//go:build windows

package factory

import (
	"mfeeder/internal/config"
	"mfeeder/internal/watcher/core"
	"mfeeder/internal/watcher/windows"
)

func newWatcher(c *config.Conf) core.Watcher {
	return &windows.WinWatcher{Cfg: c}
}
