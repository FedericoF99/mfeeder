package factory

import (
	"mfeeder/internal/config"
	"mfeeder/internal/watcher/core"
	"mfeeder/internal/watcher/windows"
	"runtime"
)

func NewWatcher(c *config.Conf) core.Watcher {
	if runtime.GOOS == "windows" {
		return windows.NewWatcher(c)
	}
	return nil
}
