package cmd

import (
	"strings"
	"testing"

	"mfeeder/internal/sqlite"
)

func TestPrintSessionsSortsAndFormatsRows(t *testing.T) {
	sessions := []sqlite.Session{
		{Exe: "browser", Title: "docs", TimeOpened: 30 * 60 * 1000, TimeFocused: 5 * 60 * 1000},
		{Exe: "code", Title: "project", TimeOpened: 2 * 60 * 60 * 1000, TimeFocused: 70 * 60 * 1000},
	}

	out := captureStdout(t, func() {
		if err := PrintSessions(sessions); err != nil {
			t.Fatalf("print sessions: %v", err)
		}
	})

	if !strings.Contains(out, "EXE") || !strings.Contains(out, "TIME OPENED") {
		t.Fatalf("missing header: %q", out)
	}
	codeIndex := strings.Index(out, "code")
	browserIndex := strings.Index(out, "browser")
	if codeIndex == -1 || browserIndex == -1 || codeIndex > browserIndex {
		t.Fatalf("sessions not sorted by focus time: %q", out)
	}
	if !strings.Contains(out, "2h 0m") || !strings.Contains(out, "1h 10m") {
		t.Fatalf("formatted durations missing: %q", out)
	}
}

func TestPrintGroupedByExeAggregatesAndSorts(t *testing.T) {
	sessions := []sqlite.Session{
		{Exe: "code", Title: "one", TimeOpened: 10_000, TimeFocused: 10_000},
		{Exe: "code", Title: "two", TimeOpened: 20_000, TimeFocused: 20_000},
		{Exe: "browser", Title: "docs", TimeOpened: 90_000, TimeFocused: 1_000},
	}

	out := captureStdout(t, func() {
		if err := PrintGroupedByExe(sessions); err != nil {
			t.Fatalf("print grouped by exe: %v", err)
		}
	})

	if !strings.Contains(out, "EXE") || !strings.Contains(out, "code") || !strings.Contains(out, "browser") {
		t.Fatalf("missing grouped output: %q", out)
	}
	if strings.Index(out, "code") > strings.Index(out, "browser") {
		t.Fatalf("groups not sorted by focus time: %q", out)
	}
}

func TestPrintGroupedByProjectFiltersIDEsAndCleansTitles(t *testing.T) {
	sessions := []sqlite.Session{
		{Exe: "code", Title: "mfeeder - internal/cmd", TimeOpened: 120_000, TimeFocused: 60_000},
		{Exe: "browser", Title: "mfeeder - docs", TimeOpened: 120_000, TimeFocused: 60_000},
	}

	out := captureStdout(t, func() {
		if err := PrintGroupedByProject(sessions); err != nil {
			t.Fatalf("print grouped by project: %v", err)
		}
	})

	if !strings.Contains(out, "PROJECT") || !strings.Contains(out, "mfeeder") || !strings.Contains(out, "internal/cmd") {
		t.Fatalf("missing project output: %q", out)
	}
	if strings.Contains(out, "browser") || strings.Contains(out, "docs") {
		t.Fatalf("non-IDE session should be filtered out: %q", out)
	}
}

func TestIsIde(t *testing.T) {
	if !isIde("GoLand64") {
		t.Fatal("expected GoLand64 to be recognized as IDE")
	}
	if isIde("firefox") {
		t.Fatal("did not expect firefox to be recognized as IDE")
	}
}

func TestProjectFromTitleSplitsKnownSeparators(t *testing.T) {
	project, title := projectFromTitle("mfeeder – internal/sqlite")
	if project != "mfeeder" || title != "internal/sqlite" {
		t.Fatalf("unexpected project split: project=%q title=%q", project, title)
	}
}

func TestProjectFromTitlePreservesHyphenatedProjectName(t *testing.T) {
	project, title := projectFromTitle("my-project - internal/sqlite")
	if project != "my-project" || title != "internal/sqlite" {
		t.Fatalf("unexpected hyphenated project split: project=%q title=%q", project, title)
	}
}

func TestProjectFromTitleFallback(t *testing.T) {
	project, title := projectFromTitle("plain title")
	if project != "undefined" || title != "" {
		t.Fatalf("unexpected undefined project split: project=%q title=%q", project, title)
	}
}

func TestFormatMs(t *testing.T) {
	if got := formatMs(90 * 60 * 1000); got != "1h 30m" {
		t.Fatalf("unexpected formatted duration: %q", got)
	}
}
