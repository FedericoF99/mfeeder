package cmd

import (
	"fmt"
	"mfeeder/internal/sqlite"
	"os"
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

func Help() {
	fmt.Println("Metric Feeder help")
	fmt.Println("available commands:")
	fmt.Println("	mfeeder day <date>")
	fmt.Println("		prints the feed for the given date, if no date is given, prints the feed for today")
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

	for _, s := range sessions {
		_, err = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.Exe, s.Titles, formatMs(s.TimeOpened), formatMs(s.TimeFocused))
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
	var exeMap = make(map[string][]sqlite.Session)

	for _, s := range sessions {
		exe := s.Exe
		if _, ok := exeMap[exe]; ok {
			exeMap[exe] = append(exeMap[exe], s)
		} else {
			exeMap[exe] = []sqlite.Session{s}
		}
	}

	err := printMap(exeMap, "EXE\tTIME OPENED\tTIME FOCUSED")
	if err != nil {
		return err
	}

	return nil
}

func PrintGroupedByProject(sessions []sqlite.Session) error {
	var projectMap = make(map[string][]sqlite.Session)

	for _, s := range sessions {
		ide := isIde(s.Exe)
		if !ide {
			continue
		}

		project := projectFromTitle(s.Titles)
		if _, ok := projectMap[project]; ok {
			projectMap[project] = append(projectMap[project], s)
		} else {
			projectMap[project] = []sqlite.Session{s}
		}
	}

	err := printMap(projectMap, "PROJECT\tTIME OPENED\tTIME FOCUSED")
	if err != nil {
		return err
	}

	return nil
}

func printMap(m map[string][]sqlite.Session, title string) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, err := fmt.Fprintln(w, title)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, "\t\t")
	if err != nil {
		return err
	}

	for k, v := range m {
		var open int64 = 0
		var focus int64 = 0
		buffer := make([]string, 0, len(v))

		for i, s := range v {
			prefix := "  ├─ "
			if i == len(v)-1 {
				prefix = "  └─ "
			}

			buffer = append(buffer, fmt.Sprintf("%s\t%s\t%s\n", prefix+" "+s.Titles, formatMs(s.TimeOpened), formatMs(s.TimeFocused)))
			open += s.TimeOpened
			focus += s.TimeFocused
		}

		_, err = fmt.Fprintf(w, "%s\t%s\t%s\n", k, formatMs(open), formatMs(focus))
		if err != nil {
			return err
		}

		for _, s := range buffer {
			_, err = fmt.Fprint(w, s)
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

func isIde(exe string) bool {
	key := strings.ToLower(exe)
	if _, ok := exeGroups[key]; ok {
		return true
	}
	return false
}

func projectFromTitle(title string) string {
	for _, sep := range []rune{'–', '—', '-'} {
		if i := strings.Index(title, string(sep)); i > 0 {
			return strings.TrimSpace(title[:i])
		}
	}
	return "undefined"
}

func formatMs(ms int64) string {
	d := time.Duration(ms) * time.Millisecond
	m := int(d.Minutes())

	hours := m / 60
	minutes := m - (hours * 60)
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
