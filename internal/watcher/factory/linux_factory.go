//go:build linux

package factory

import (
	"log"
	"mfeeder/internal/config"
	"mfeeder/internal/watcher/core"
	"os"
	"strings"
)

func newWatcher(c *config.Conf) core.Watcher {
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "" {

	} else if os.Getenv("XDG_SESSION_TYPE") == "wayland" &&
		strings.Contains(os.Getenv("XDG_CURRENT_DESKTOP"), "GNOME") {

	} else if os.Getenv("XDG_SESSION_TYPE") == "x11" ||
		os.Getenv("DISPLAY") != "" {

	}

	log.Fatal("Unsupported platform")
	return nil
}
