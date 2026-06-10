package cmd

import (
	"fmt"
	"mfeeder/internal/config"
	"os"
	"strconv"
	"strings"
)

func Run() error {
	c, err := parseCmd(os.Args)
	if err != nil {
		return fmt.Errorf("error parsing command: %v\n", err)
	}

	err = run(c)
	if err != nil {
		return fmt.Errorf("error executing %s command: %v\n", c.cmd, err)
	}

	return nil
}

func run(c *cmd) error {
	switch c.cmd {
	case "day":
		return runDay(c)
	case "ex":
		return runEx(c)
	case "sys":
		return runSys()
	case "get":
		return runGet(c)
	case "help":
		help()
		return nil
	}
	return fmt.Errorf("unknown command: %v", c.cmd)
}

func runDay(c *cmd) error {
	// todo: implement day
	return nil
}

func runEx(c *cmd) error {
	var err error

	if c.args[0] == "add" {
		err = config.AddExclusion(c.opt["add"])
	} else if c.args[0] == "rm" {
		err = config.RmExclusion(c.opt["rm"])
	}

	return err
}

func runSys() error {
	err := config.ToggleSystem()
	return err
}

func runGet(c *cmd) error {
	if c.args[0] == "ex" {
		ex, err := config.GetExclusions()
		if err != nil {
			return err
		}
		fmt.Printf("EXCLUSIONS=%s\n", strings.Join(ex, ", "))
		return nil
	} else if c.args[0] == "sys" {
		sys, err := config.GetSystem()
		if err != nil {
			return err
		}
		fmt.Printf("SYSTEM=%s\n", strconv.FormatBool(sys))
		return nil
	}

	return fmt.Errorf("unknown get command: %v", c.args[0])
}
