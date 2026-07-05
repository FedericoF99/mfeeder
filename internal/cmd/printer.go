package cmd

import (
	"fmt"
	"mfeeder/internal/sqlite"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

var exeGroups = map[string]struct{}{
	"code": {}, "code-insiders": {}, "goland": {}, "goland64": {},
	"idea": {}, "idea64": {}, "pycharm": {}, "pycharm64": {},
	"webstorm": {}, "webstorm64": {}, "rider": {}, "rider64": {},
	"clion": {}, "clion64": {}, "rustrover": {}, "rustrover64": {},
	"phpstorm": {}, "phpstorm64": {}, "datagrip": {}, "datagrip64": {},
	"nvim": {}, "vim": {}, "sublime_text": {}, "zed": {},
}

type group struct {
	key         string
	sessions    []sqlite.Session
	timeFocused int64
	timeOpened  int64
}

func Help() {
	fmt.Println("Metric Feeder help")
	fmt.Println("available commands:")
	fmt.Println("	mfeeder day <date> <optional flags>")
	fmt.Println("		prints the feed for the given date, if no date is given, prints the feed for today")
	fmt.Println("		available flags:")
	fmt.Println("			-e: groups the feed by the exe name")
	fmt.Println("			-p: groups the feed by the project name (only for ides)")
	fmt.Println("")
	fmt.Println("	mfeeder ex <option> <value>")
	fmt.Println("		options available:")
	fmt.Println("			- add: adds an exclusion to the config file")
	fmt.Println("			- rm: removes an exclusion from the config file")
	fmt.Println("		exluded processes will not be recorder and printed in the feed")
	fmt.Println("")
	fmt.Println("	mfeeder get <option>")
	fmt.Println("		available options:")
	fmt.Println("			- ex: prints the current exclusions")
	fmt.Println("")
	fmt.Println("	mfeeder --help")
	fmt.Println("		prints this help message")
	fmt.Println("")
}

func PrintSessions(sessions []sqlite.Session) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	_, err := fmt.Fprintln(w, "EXE\tTITLES\tTIME OPENED\tTIME FOCUSED")
	if err != nil {
		return err
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].TimeFocused > sessions[j].TimeFocused
	})

	for _, s := range sessions {
		_, err = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.Exe, s.Title, formatMs(s.TimeOpened), formatMs(s.TimeFocused))
		if err != nil {
			return err
		}
	}

	err = w.Flush()
	if err != nil {
		return err
	}

	return nil
}

func PrintGroupedByExe(sessions []sqlite.Session) error {
	var groups = make([]group, 0)

	for _, s := range sessions {
		exe := s.Exe
		groups = insert(exe, s, groups)
	}

	err := printGroups(groups, "EXE\tTIME OPENED\tTIME FOCUSED")
	if err != nil {
		return err
	}

	return nil
}

func PrintGroupedByProject(sessions []sqlite.Session) error {
	var groups = make([]group, 0)

	for _, s := range sessions {
		ide := isIde(s.Exe)
		if !ide {
			continue
		}

		project, cleanTitle := projectFromTitle(s.Title)
		if cleanTitle != "" {
			s.Title = cleanTitle
		}

		groups = insert(project, s, groups)
	}

	err := printGroups(groups, "PROJECT\tTIME OPENED\tTIME FOCUSED")
	if err != nil {
		return err
	}

	return nil
}

func printGroups(gs []group, title string) error {
	for i := range gs {
		var tf int64 = 0
		var to int64 = 0
		for _, s := range gs[i].sessions {
			tf += s.TimeFocused
			to += s.TimeOpened
		}

		gs[i].timeFocused = tf
		gs[i].timeOpened = to
	}

	sort.Slice(gs, func(i, j int) bool {
		if gs[i].timeFocused == gs[j].timeFocused {
			return gs[i].timeOpened > gs[j].timeOpened
		}

		return gs[i].timeFocused > gs[j].timeFocused
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, err := fmt.Fprintln(w, title)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, "\t\t")
	if err != nil {
		return err
	}

	for _, v := range gs {
		_, err = fmt.Fprintf(w, "%s\t%s\t%s\n", v.key, formatMs(v.timeOpened), formatMs(v.timeFocused))
		if err != nil {
			return err
		}

		for i, s := range v.sessions {
			prefix := "  ├─ "
			if i == len(v.sessions)-1 {
				prefix = "  └─ "
			}

			_, err = fmt.Fprintf(w, "%s\t%s\t%s\n", prefix+" "+s.Title, formatMs(s.TimeOpened), formatMs(s.TimeFocused))
			if err != nil {
				return err
			}
		}

		_, err = fmt.Fprintln(w, "\t\t")
		if err != nil {
			return err
		}
	}

	err = w.Flush()
	if err != nil {
		return err
	}

	return nil
}

func insert(key string, s sqlite.Session, gs []group) []group {
	for i := range gs {
		if gs[i].key == key {
			gs[i].sessions = append(gs[i].sessions, s)
			return gs
		}
	}

	gs = append(gs, group{key: key, sessions: []sqlite.Session{s}})
	return gs
}

func isIde(exe string) bool {
	key := strings.ToLower(exe)
	if _, ok := exeGroups[key]; ok {
		return true
	}
	return false
}

func projectFromTitle(title string) (string, string) {
	for _, sep := range []string{" – ", " — ", " - "} {
		if i := strings.Index(title, sep); i > 0 {
			return strings.TrimSpace(title[:i]), strings.TrimSpace(title[i+len(sep):])
		}
	}
	return "undefined", ""
}

func formatMs(ms int64) string {
	d := time.Duration(ms) * time.Millisecond
	m := int(d.Minutes())

	hours := m / 60
	minutes := m - (hours * 60)
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
