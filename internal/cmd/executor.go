package cmd

import (
	"fmt"
	"mfeeder/internal/config"
	"mfeeder/internal/sqlite"
	"os"
	"strings"
	"time"
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
	case "get":
		return runGet(c)
	case "help":
		Help()
		return nil
	}
	return fmt.Errorf("unknown command: %v", c.cmd)
}

func runDay(c *cmd) error {
	db, err := sqlite.Init()
	if err != nil {
		return err
	}
	defer db.Close()

	var sa []sqlite.Session
	now := time.Now()

	var day time.Time

	if len(c.args) == 0 {
		day = now
		sa, err = sqlite.GetDay(day.Format("2006-01-02"), db)
		if err != nil {
			return err
		}
	} else {
		rawDay, err := time.Parse("01-02", c.args[0])
		if err != nil {
			return err
		}

		if rawDay.Month() > now.Month() {
			day = time.Date(now.AddDate(-1, 0, 0).Year(),
				rawDay.Month(), rawDay.Day(), 0, 0, 0, 0, time.Local)
		} else {
			day = time.Date(now.Year(), rawDay.Month(), rawDay.Day(), 0, 0, 0, 0, time.Local)
		}

		sa, err = sqlite.GetDay(day.Format("2006-01-02"), db)
		if err != nil {
			return err
		}
	}

	if len(c.args) == 1 {
		err = PrintSessions(sa)
		if err != nil {
			return err
		}
	} else if c.args[1] == "p" {
		err = PrintGroupedByProject(sa)
		if err != nil {
			return err
		}
	} else if c.args[1] == "e" {
		err = PrintGroupedByExe(sa)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("unknown format")
	}

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

func runGet(c *cmd) error {
	if c.args[0] == "ex" {
		ex, err := config.GetExclusions()
		if err != nil {
			return err
		}
		fmt.Printf("EXCLUSIONS=%s\n", strings.Join(ex, ", "))
		return nil
	}

	return fmt.Errorf("unknown get command: %v", c.args[0])
}
