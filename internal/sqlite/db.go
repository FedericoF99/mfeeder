package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

const schema = `
	CREATE TABLE IF NOT EXISTS sessions
	(
		id integer primary key autoincrement,
		
		pid integer not null,
		exe text not null,
		title text not null,
		
		time_opened_ms integer not null,
		time_focused_ms integer not null,
		
		opened_started_at_ms integer not null,
		focused_started_at_ms integer not null,
		
		opened boolean not null,
		focused boolean not null,
	
		session_date text not null
	);
	create unique index if not exists pid_index on sessions(pid, exe, title);
	create index if not exists pid_date_index on sessions(pid, session_date);
	
	CREATE TABLE IF NOT EXISTS sessions_history
	(
		id integer primary key autoincrement,
		
		exe text not null,
		title text not null,
		
		time_opened_ms integer not null,
		time_focused_ms integer not null,
	
		session_date text not null
	);
	create index if not exists date_index on sessions_history(session_date);
`

func Init() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "mfeeder.db")
	if err != nil {
		return nil, fmt.Errorf("database init failed: %v", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("database connection failed: %v", err)
	}

	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func StartupClean(ctx context.Context, db *sql.DB) error {
	err := CloseAll(ctx, db)
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = moveToHistory(ctx, tx)
	if err != nil {
		e := tx.Rollback()
		if e != nil {
			return e
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func moveToHistory(ctx context.Context, tx *sql.Tx) error {
	day := time.Now().Format("2006-01-02")

	iRes, err := tx.ExecContext(ctx, `insert into sessions_history (exe, title, time_opened_ms, time_focused_ms, session_date) 
		select exe, title, time_opened_ms, time_focused_ms, session_date from sessions where date(session_date) < date(?) order by date(session_date), id`, day)
	if err != nil {
		return err
	}

	sRes := tx.QueryRowContext(ctx, "select count(id) from sessions where date(session_date) < date(?)", day)

	var count int64
	if err = sRes.Scan(&count); err != nil {
		return err
	}

	if r, er := iRes.RowsAffected(); er != nil {
		return er
	} else if r != count {
		return fmt.Errorf("rows affected mismatch")
	}

	_, err = tx.ExecContext(ctx, "delete from sessions where date(session_date) < date(?)", day)
	if err != nil {
		return err
	}

	return nil
}

func CloseAll(ctx context.Context, db *sql.DB) error {
	now := time.Now()

	_, err := db.ExecContext(ctx, "update sessions set opened = 0, time_opened_ms = (? - opened_started_at_ms) + time_opened_ms where opened = 1", now.UnixMilli())
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "update sessions set focused = 0, time_focused_ms = (? - focused_started_at_ms) + time_focused_ms where focused = 1", now.UnixMilli())
	if err != nil {
		return err
	}

	return nil
}
