package cmd

import (
	"fmt"
	"time"
)

type cmd struct {
	cmd  string
	opt  map[string]string
	args []string
}

func parseCmd(args []string) (*cmd, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("command not found, try: mfeeder --help")
	}

	c := cmd{
		cmd:  args[1],
		opt:  make(map[string]string),
		args: make([]string, 0),
	}

	var err error

	switch c.cmd {
	case "day":
		err = validateDay(args, &c)
	case "ex":
		err = validateEx(args, &c)
	case "get":
		err = validateGet(args, &c)
	case "--help":
		if len(args) > 2 {
			err = fmt.Errorf("unexpected arguments for help command, try: mfeeder --help")
		}
		c.cmd = "help"
	default:
		return nil, fmt.Errorf("command not found, try: mfeeder --help")
	}

	return &c, err
}

func validateDay(args []string, c *cmd) error {
	aLen := len(args)

	if aLen > 4 {
		return fmt.Errorf("unexpected arguments for day command, try: mfeeder --help")
	}

	for i := 2; i < aLen; i++ {
		if _, err := time.Parse("01-02", args[i]); err == nil {
			if len(c.args) != 0 {
				return fmt.Errorf("multiple dates not allowed, try: mfeeder --help")
			}

			c.args = append(c.args, args[i])
		} else if args[i] == "-e" || args[i] == "-p" {
			if len(c.opt) != 0 {
				return fmt.Errorf("multiple flags not allowed, try: mfeeder --help")
			}

			c.opt[args[i][1:]] = ""
		} else {
			return fmt.Errorf("invalid argument %s for day command, try: mfeeder --help", args[i])
		}
	}

	c.cmd = "day"
	return nil
}

func validateEx(args []string, c *cmd) error {
	aLen := len(args)

	if aLen < 4 {
		return fmt.Errorf("missing arguments for ex command, try: mfeeder --help")
	}
	if len(args) > 4 {
		return fmt.Errorf("unexpected arguments for ex command, try: mfeeder --help")
	}

	if args[2] != "add" && args[2] != "rm" {
		return fmt.Errorf("invalid option for ex command, try: mfeeder --help")
	}
	if args[3] == "" {
		return fmt.Errorf("missing argument for ex command, try: mfeeder --help")
	}

	c.opt[args[2]] = args[3]

	c.cmd = "ex"
	return nil
}

func validateGet(args []string, c *cmd) error {
	aLen := len(args)

	if aLen < 3 {
		return fmt.Errorf("missing arguments for get command, try: mfeeder --help")
	}
	if aLen > 3 {
		return fmt.Errorf("unexpected arguments for get command, try: mfeeder --help")
	}

	if args[2] != "ex" {
		return fmt.Errorf("invalid option for get command, try: mfeeder --help")
	}

	c.args = append(c.args, args[2])

	c.cmd = "get"
	return nil
}
