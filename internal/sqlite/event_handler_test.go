package sqlite

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"mfeeder/internal/watcher/core"

	_ "modernc.org/sqlite"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.SetMaxOpenConns(1)

	if _, err = db.Exec(schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close test db: %v", err)
		}
	})

	return db
}

func TestWindowOpenedCreatesAndIsIdempotent(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	window := core.Window{Pid: 10, Exe: "code", Title: "mfeeder"}

	if err := WindowOpened(ctx, window, db); err != nil {
		t.Fatalf("open window: %v", err)
	}
	if err := WindowOpened(ctx, window, db); err != nil {
		t.Fatalf("open window again: %v", err)
	}

	var count int
	var opened bool
	var focused bool
	err := db.QueryRowContext(ctx, "select count(*), opened, focused from sessions").Scan(&count, &opened, &focused)
	if err != nil {
		t.Fatalf("read session: %v", err)
	}
	if count != 1 || !opened || focused {
		t.Fatalf("unexpected session state: count=%d opened=%v focused=%v", count, opened, focused)
	}
}

func TestWindowFocusedCreatesAndResetsPreviousFocus(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	first := core.Window{Pid: 10, Exe: "code", Title: "first"}
	second := core.Window{Pid: 11, Exe: "go", Title: "second"}

	if err := WindowFocused(ctx, first, db); err != nil {
		t.Fatalf("focus first: %v", err)
	}
	setSessionStartedAt(t, db, first, time.Now().Add(-time.Second))
	if err := WindowFocused(ctx, second, db); err != nil {
		t.Fatalf("focus second: %v", err)
	}

	var firstFocused bool
	var firstFocusedMs int64
	err := db.QueryRowContext(ctx, "select focused, time_focused_ms from sessions where pid = ?", first.Pid).
		Scan(&firstFocused, &firstFocusedMs)
	if err != nil {
		t.Fatalf("read first session: %v", err)
	}
	if firstFocused {
		t.Fatalf("first window should not remain focused")
	}
	if firstFocusedMs <= 0 {
		t.Fatalf("first window should have accumulated focus time, got %d", firstFocusedMs)
	}

	var secondOpened bool
	var secondFocused bool
	err = db.QueryRowContext(ctx, "select opened, focused from sessions where pid = ?", second.Pid).
		Scan(&secondOpened, &secondFocused)
	if err != nil {
		t.Fatalf("read second session: %v", err)
	}
	if !secondOpened || !secondFocused {
		t.Fatalf("second window should be opened and focused, opened=%v focused=%v", secondOpened, secondFocused)
	}
}

func TestWindowFocusedAlreadyFocusedIsIdempotent(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	window := core.Window{Pid: 10, Exe: "code", Title: "mfeeder"}

	if err := WindowFocused(ctx, window, db); err != nil {
		t.Fatalf("focus window: %v", err)
	}
	startedAt := time.Now().Add(-time.Second)
	setSessionStartedAt(t, db, window, startedAt)

	if err := WindowFocused(ctx, window, db); err != nil {
		t.Fatalf("focus already focused window: %v", err)
	}

	var focusedStartedAt int64
	if err := db.QueryRowContext(ctx, "select focused_started_at_ms from sessions where pid = ?", window.Pid).
		Scan(&focusedStartedAt); err != nil {
		t.Fatalf("read focused start: %v", err)
	}
	if focusedStartedAt != startedAt.UnixMilli() {
		t.Fatalf("already focused window should keep original start time, got %d want %d", focusedStartedAt, startedAt.UnixMilli())
	}
}

func TestWindowClosedAccumulatesOpenedAndFocusedTime(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	window := core.Window{Pid: 10, Exe: "code", Title: "mfeeder"}

	if err := WindowFocused(ctx, window, db); err != nil {
		t.Fatalf("focus window: %v", err)
	}
	setSessionStartedAt(t, db, window, time.Now().Add(-time.Second))
	if err := WindowClosed(ctx, window, db); err != nil {
		t.Fatalf("close window: %v", err)
	}

	var opened bool
	var focused bool
	var openedMs int64
	var focusedMs int64
	err := db.QueryRowContext(ctx, `select opened, focused, time_opened_ms, time_focused_ms from sessions where pid = ?`, window.Pid).
		Scan(&opened, &focused, &openedMs, &focusedMs)
	if err != nil {
		t.Fatalf("read closed session: %v", err)
	}
	if opened || focused {
		t.Fatalf("window should be closed and unfocused, opened=%v focused=%v", opened, focused)
	}
	if openedMs <= 0 || focusedMs <= 0 {
		t.Fatalf("window should accumulate opened and focused time, opened=%d focused=%d", openedMs, focusedMs)
	}
}

func TestWindowClosedDoesNotChangeUnfocusedTime(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	window := core.Window{Pid: 10, Exe: "code", Title: "mfeeder"}

	if err := WindowOpened(ctx, window, db); err != nil {
		t.Fatalf("open window: %v", err)
	}
	setSessionStartedAt(t, db, window, time.Now().Add(-time.Second))
	if err := WindowClosed(ctx, window, db); err != nil {
		t.Fatalf("close window: %v", err)
	}

	var openedMs int64
	var focusedMs int64
	err := db.QueryRowContext(ctx, `select time_opened_ms, time_focused_ms from sessions where pid = ?`, window.Pid).
		Scan(&openedMs, &focusedMs)
	if err != nil {
		t.Fatalf("read closed session: %v", err)
	}
	if openedMs <= 0 {
		t.Fatalf("window should accumulate opened time, got %d", openedMs)
	}
	if focusedMs != 0 {
		t.Fatalf("unfocused window should not accumulate focused time, got %d", focusedMs)
	}
}

func TestWindowClosedDoesNotWriteNegativeDurations(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	window := core.Window{Pid: 10, Exe: "code", Title: "mfeeder"}

	if err := WindowOpened(ctx, window, db); err != nil {
		t.Fatalf("open window: %v", err)
	}
	setSessionStartedAt(t, db, window, time.Now().Add(time.Second))
	if err := WindowClosed(ctx, window, db); err != nil {
		t.Fatalf("close window: %v", err)
	}

	var openedMs int64
	if err := db.QueryRowContext(ctx, "select time_opened_ms from sessions where pid = ?", window.Pid).
		Scan(&openedMs); err != nil {
		t.Fatalf("read opened time: %v", err)
	}
	if openedMs < 0 {
		t.Fatalf("opened time should not be negative, got %d", openedMs)
	}
}

func TestWindowClosedAlreadyClosedDoesNotAccumulateAgain(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	window := core.Window{Pid: 10, Exe: "code", Title: "mfeeder"}

	if err := WindowOpened(ctx, window, db); err != nil {
		t.Fatalf("open window: %v", err)
	}
	setSessionStartedAt(t, db, window, time.Now().Add(-time.Second))
	if err := WindowClosed(ctx, window, db); err != nil {
		t.Fatalf("close window: %v", err)
	}

	var openedAfterFirstClose int64
	if err := db.QueryRowContext(ctx, "select time_opened_ms from sessions where pid = ?", window.Pid).
		Scan(&openedAfterFirstClose); err != nil {
		t.Fatalf("read first close time: %v", err)
	}

	setSessionStartedAt(t, db, window, time.Now().Add(-10*time.Second))
	if err := WindowClosed(ctx, window, db); err != nil {
		t.Fatalf("close already closed window: %v", err)
	}

	var openedAfterSecondClose int64
	if err := db.QueryRowContext(ctx, "select time_opened_ms from sessions where pid = ?", window.Pid).
		Scan(&openedAfterSecondClose); err != nil {
		t.Fatalf("read second close time: %v", err)
	}
	if openedAfterSecondClose != openedAfterFirstClose {
		t.Fatalf("closing an already closed window should not add time, got %d want %d", openedAfterSecondClose, openedAfterFirstClose)
	}
}

func TestWindowClosedReturnsErrorForUnknownWindow(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	window := core.Window{Pid: 404, Exe: "missing", Title: "missing"}

	if err := WindowClosed(ctx, window, db); err == nil {
		t.Fatal("expected error closing unknown window")
	}
}

func TestWindowClosedClearsFocusedEvenWhenOpenedFlagIsFalse(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	window := core.Window{Pid: 10, Exe: "code", Title: "mfeeder"}

	insertSession(t, db, window.Pid, window.Exe, window.Title, time.Now().Format("2006-01-02"), false, true, 100, 50)

	if err := WindowClosed(ctx, window, db); err != nil {
		t.Fatalf("close inconsistent window: %v", err)
	}

	var focused bool
	if err := db.QueryRowContext(ctx, "select focused from sessions where pid = ?", window.Pid).Scan(&focused); err != nil {
		t.Fatalf("read focused state: %v", err)
	}
	if focused {
		t.Fatal("closed window should not remain focused")
	}
}

func TestWindowOpenedAfterCloseDoesNotOvercountOpenedTime(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	window := core.Window{Pid: 10, Exe: "code", Title: "mfeeder"}

	if err := WindowOpened(ctx, window, db); err != nil {
		t.Fatalf("open window: %v", err)
	}
	setSessionStartedAt(t, db, window, time.Now().Add(-10*time.Second))
	if err := WindowClosed(ctx, window, db); err != nil {
		t.Fatalf("close window: %v", err)
	}

	var openedAfterFirstClose int64
	if err := db.QueryRowContext(ctx, "select time_opened_ms from sessions where pid = ?", window.Pid).
		Scan(&openedAfterFirstClose); err != nil {
		t.Fatalf("read first close time: %v", err)
	}

	if err := WindowOpened(ctx, window, db); err != nil {
		t.Fatalf("reopen window: %v", err)
	}
	setSessionStartedAt(t, db, window, time.Now().Add(-time.Second))
	if err := WindowClosed(ctx, window, db); err != nil {
		t.Fatalf("close reopened window: %v", err)
	}

	var openedAfterSecondClose int64
	if err := db.QueryRowContext(ctx, "select time_opened_ms from sessions where pid = ?", window.Pid).
		Scan(&openedAfterSecondClose); err != nil {
		t.Fatalf("read second close time: %v", err)
	}

	increment := openedAfterSecondClose - openedAfterFirstClose
	if increment < 900 || increment > 1100 {
		t.Fatalf("expected reopened interval near 1s, got increment %dms", increment)
	}
}

func TestCloseAllClosesOpenAndFocusedSessions(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	focused := core.Window{Pid: 10, Exe: "code", Title: "focused"}
	opened := core.Window{Pid: 11, Exe: "go", Title: "opened"}

	if err := WindowFocused(ctx, focused, db); err != nil {
		t.Fatalf("focus window: %v", err)
	}
	if err := WindowOpened(ctx, opened, db); err != nil {
		t.Fatalf("open window: %v", err)
	}
	setSessionStartedAt(t, db, focused, time.Now().Add(-time.Second))
	setSessionStartedAt(t, db, opened, time.Now().Add(-time.Second))
	if err := CloseAll(ctx, db); err != nil {
		t.Fatalf("close all: %v", err)
	}

	rows, err := db.QueryContext(ctx, "select opened, focused, time_opened_ms from sessions")
	if err != nil {
		t.Fatalf("query sessions: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var opened bool
		var focused bool
		var openedMs int64
		if err := rows.Scan(&opened, &focused, &openedMs); err != nil {
			t.Fatalf("scan session: %v", err)
		}
		if opened || focused || openedMs <= 0 {
			t.Fatalf("unexpected closed session state: opened=%v focused=%v openedMs=%d", opened, focused, openedMs)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate sessions: %v", err)
	}
}

func TestCloseAllDoesNotWriteNegativeDurations(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	window := core.Window{Pid: 10, Exe: "code", Title: "future"}

	insertSession(t, db, window.Pid, window.Exe, window.Title, time.Now().Format("2006-01-02"), true, true, 0, 0)
	setSessionStartedAt(t, db, window, time.Now().Add(time.Second))

	if err := CloseAll(ctx, db); err != nil {
		t.Fatalf("close all: %v", err)
	}

	var openedMs int64
	var focusedMs int64
	if err := db.QueryRowContext(ctx, "select time_opened_ms, time_focused_ms from sessions where pid = ?", window.Pid).
		Scan(&openedMs, &focusedMs); err != nil {
		t.Fatalf("read times: %v", err)
	}
	if openedMs < 0 || focusedMs < 0 {
		t.Fatalf("durations should not be negative, opened=%d focused=%d", openedMs, focusedMs)
	}
}

func TestStartupCleanMovesOnlyPreviousDaysToHistory(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	insertSession(t, db, 10, "old", "old title", yesterday, false, false, 100, 50)
	insertSession(t, db, 11, "new", "new title", today, true, true, 0, 0)

	if err := StartupClean(ctx, db); err != nil {
		t.Fatalf("startup clean: %v", err)
	}

	var historyCount int
	if err := db.QueryRowContext(ctx, "select count(*) from sessions_history where exe = 'old'").Scan(&historyCount); err != nil {
		t.Fatalf("count history: %v", err)
	}
	if historyCount != 1 {
		t.Fatalf("expected old session in history, got %d", historyCount)
	}

	var currentCount int
	var opened bool
	var focused bool
	if err := db.QueryRowContext(ctx, "select count(*), opened, focused from sessions where exe = 'new'").
		Scan(&currentCount, &opened, &focused); err != nil {
		t.Fatalf("read current session: %v", err)
	}
	if currentCount != 1 || opened || focused {
		t.Fatalf("today session should remain but be closed, count=%d opened=%v focused=%v", currentCount, opened, focused)
	}
}

func TestStartupCleanMovesUpdatedOpenOldSessionToHistory(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	window := core.Window{Pid: 10, Exe: "old", Title: "old title"}

	insertSession(t, db, window.Pid, window.Exe, window.Title, yesterday, true, true, 100, 50)
	setSessionStartedAt(t, db, window, time.Now().Add(-time.Second))

	if err := StartupClean(ctx, db); err != nil {
		t.Fatalf("startup clean: %v", err)
	}

	var sessionCount int
	if err := db.QueryRowContext(ctx, "select count(*) from sessions").Scan(&sessionCount); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if sessionCount != 0 {
		t.Fatalf("old session should be removed from current sessions, got %d", sessionCount)
	}

	var openedMs int64
	var focusedMs int64
	if err := db.QueryRowContext(ctx, "select time_opened_ms, time_focused_ms from sessions_history where exe = ?", window.Exe).
		Scan(&openedMs, &focusedMs); err != nil {
		t.Fatalf("read history session: %v", err)
	}
	if openedMs < 1000 || focusedMs < 1000 {
		t.Fatalf("history should include closed open/focus intervals, opened=%d focused=%d", openedMs, focusedMs)
	}
}

func TestGetDayReturnsSessionsMovedToHistory(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	yesterday := time.Now().AddDate(0, 0, -1)

	insertSession(t, db, 10, "code", "old project", yesterday.Format("2006-01-02"), false, false, 1200, 600)

	if err := StartupClean(ctx, db); err != nil {
		t.Fatalf("startup clean: %v", err)
	}

	sessions, err := GetDay(yesterday, db)
	if err != nil {
		t.Fatalf("get historical day: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected historical session, got %d: %#v", len(sessions), sessions)
	}
	if sessions[0].Exe != "code" || sessions[0].Title != "old project" || sessions[0].TimeOpened != 1200 || sessions[0].TimeFocused != 600 {
		t.Fatalf("unexpected historical session: %#v", sessions[0])
	}
}

func TestGetDayAggregatesAndOrdersSessions(t *testing.T) {
	db := newTestDB(t)
	day := time.Now()
	dayStr := day.Format("2006-01-02")
	insertSession(t, db, 10, "code", "project", dayStr, false, false, 400, 60)
	insertSession(t, db, 11, "browser", "docs", dayStr, false, false, 500, 10)
	insertSession(t, db, 12, "terminal", "shell", dayStr, false, false, 900, 10)
	insertSession(t, db, 13, "old", "old", "2026-07-01", false, false, 1000, 1000)

	sessions, err := GetDay(day, db)
	if err != nil {
		t.Fatalf("get day: %v", err)
	}
	if len(sessions) != 3 {
		t.Fatalf("expected 3 sessions, got %d: %#v", len(sessions), sessions)
	}
	if sessions[0].Exe != "code" || sessions[0].TimeOpened != 400 || sessions[0].TimeFocused != 60 {
		t.Fatalf("unexpected first aggregate: %#v", sessions[0])
	}
	if sessions[1].Exe != "terminal" || sessions[1].TimeOpened != 900 || sessions[1].TimeFocused != 10 {
		t.Fatalf("unexpected second aggregate: %#v", sessions[1])
	}
	if sessions[2].Exe != "browser" || sessions[2].TimeOpened != 500 || sessions[2].TimeFocused != 10 {
		t.Fatalf("unexpected third aggregate: %#v", sessions[2])
	}
}

func TestGetDayAggregatesSameWindowAcrossPIDs(t *testing.T) {
	db := newTestDB(t)
	day := time.Now()
	dayStr := day.Format("2006-01-02")
	insertSession(t, db, 10, "code", "same project", dayStr, false, false, 400, 60)
	insertSession(t, db, 11, "code", "same project", dayStr, false, false, 600, 40)

	sessions, err := GetDay(day, db)
	if err != nil {
		t.Fatalf("get day: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected one aggregate for same exe/title across pids, got %d: %#v", len(sessions), sessions)
	}
	if sessions[0].TimeOpened != 1000 || sessions[0].TimeFocused != 100 {
		t.Fatalf("unexpected aggregate: %#v", sessions[0])
	}
}

func insertSession(t *testing.T, db *sql.DB, pid int, exe, title, day string, opened, focused bool, openedMs, focusedMs int64) {
	t.Helper()

	now := time.Now().UnixMilli()
	_, err := db.Exec(createQuery, pid, exe, title, now, now, focused, day)
	if err != nil {
		t.Fatalf("insert session: %v", err)
	}

	_, err = db.Exec(
		`update sessions
		 set opened = ?, focused = ?, time_opened_ms = ?, time_focused_ms = ?
		 where pid = ? and exe = ? and title = ?`,
		opened, focused, openedMs, focusedMs, pid, exe, title,
	)
	if err != nil {
		t.Fatalf("update session: %v", err)
	}
}

func setSessionStartedAt(t *testing.T, db *sql.DB, window core.Window, startedAt time.Time) {
	t.Helper()

	_, err := db.Exec(
		`update sessions
		 set opened_started_at_ms = ?, focused_started_at_ms = ?
		 where pid = ? and exe = ? and title = ?`,
		startedAt.UnixMilli(), startedAt.UnixMilli(), window.Pid, window.Exe, window.Title,
	)
	if err != nil {
		t.Fatalf("set session started at: %v", err)
	}
}

func setFocusedStartedAt(t *testing.T, db *sql.DB, window core.Window, startedAt time.Time) {
	t.Helper()

	_, err := db.Exec(
		`update sessions
		 set focused_started_at_ms = ?
		 where pid = ? and exe = ? and title = ?`,
		startedAt.UnixMilli(), window.Pid, window.Exe, window.Title,
	)
	if err != nil {
		t.Fatalf("set focused started at: %v", err)
	}
}
