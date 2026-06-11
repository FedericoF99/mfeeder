package watcher

import "mfeeder/internal/config"

type Info struct {
	Pid     int
	Exe     string
	Title   string
	Focused bool
}

type Watcher interface {
	Watch(c *config.Conf) ([]Info, error)
}
