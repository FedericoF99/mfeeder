package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mfeeder/internal/watcher/core"
	"time"
)

const createQuery = `
	insert into sessions 
	    (
	     pid, exe, title, time_opened_ms, time_focused_ms,
	     opened_started_at_ms, focused_started_at_ms, opened, focused, session_date
	    )
		values 
	(
	 ?, ?, ?, 0, 0, ?, ?, true, ?, ?
	)`

func WindowOpened(ctx context.Context, window core.Window, db *sql.DB) error {
	now := time.Now()

	row := db.QueryRowContext(ctx,
		`select id, title, opened from sessions 
            	where pid = ? and exe = ? and title = ?`,
		window.Pid, window.Exe, window.Title)

	var id int
	var title string
	var opened bool

	err := row.Scan(&id, &title, &opened)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = createWindow(false, ctx, window, db)
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}

	if opened {
		return nil
	}

	_, err = db.ExecContext(ctx,
		`update sessions set opened = 1, opened_started_at_ms = ? where id = ?`,
		now.UnixMilli(), id)

	if err != nil {
		return err
	}

	return nil
}

func createWindow(focused bool, ctx context.Context, window core.Window, db *sql.DB) error {
	now := time.Now()
	nowMs := now.UnixMilli()

	var focusedStartedAtMs int64 = 0
	if focused {
		focusedStartedAtMs = nowMs
	}

	res, err := db.ExecContext(ctx, createQuery, window.Pid, window.Exe, window.Title, nowMs, focusedStartedAtMs, focused, now.Format("2006-01-02"))
	if err != nil {
		return err
	}
	if r, e := res.RowsAffected(); e != nil {
		return e
	} else if r != 1 {
		return fmt.Errorf("unexpected number of rows affected: %d", r)
	}

	return nil
}

func WindowClosed(ctx context.Context, window core.Window, db *sql.DB) error {
	now := time.Now()

	row := db.QueryRowContext(ctx,
		`select id, title, time_opened_ms, opened_started_at_ms, time_focused_ms, focused_started_at_ms, opened 
				from sessions where pid = ? and exe = ? and title = ?`,
		window.Pid, window.Exe, window.Title)

	var id int
	var title string
	var timeOpenedMs int64
	var openedStartedAtMs int64
	var timeFocusedMs int64
	var focusedStartedAtMs int64
	var opened bool

	err := row.Scan(&id, &title, &timeOpenedMs, &openedStartedAtMs, &timeFocusedMs, &focusedStartedAtMs, &opened)
	if err != nil {
		return err
	}

	if !opened {
		return nil
	}

	msOpened := (now.UnixMilli() - openedStartedAtMs) + timeOpenedMs
	msFocused := (now.UnixMilli() - focusedStartedAtMs) + timeFocusedMs
	_, err = db.ExecContext(ctx,
		`update sessions set opened = 0, focused = 0, time_opened_ms = ?, 
                time_focused_ms = case when focused = 1 then ? else time_focused_ms end 
                where id = ?`, msOpened, msFocused, id)
	if err != nil {
		return err
	}

	return nil
}

func WindowFocused(ctx context.Context, window core.Window, db *sql.DB) error {
	now := time.Now()

	row := db.QueryRowContext(ctx,
		`select id, title, time_focused_ms, focused_started_at_ms, focused 
				from sessions where pid = ? and exe = ? and title = ?`,
		window.Pid, window.Exe, window.Title)

	var id int
	var title string
	var timeFocusedMs int64
	var focusedStartedAtMs int64
	var focused bool

	err := row.Scan(&id, &title, &timeFocusedMs, &focusedStartedAtMs, &focused)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = resetAllFocused(now, ctx, db)
			if err != nil {
				return err
			}
			err = createWindow(true, ctx, window, db)
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}

	if focused {
		return nil
	}

	err = resetAllFocused(now, ctx, db)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx,
		`update sessions set opened = 1, focused = 1, focused_started_at_ms = ? 
                where id = ?`, now.UnixMilli(), id)

	if err != nil {
		return err
	}

	return nil
}

func resetAllFocused(now time.Time, ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx,
		`update sessions set focused = 0, time_focused_ms = (? - focused_started_at_ms) + time_focused_ms 
                where focused = 1`, now.UnixMilli())

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	return nil
}
