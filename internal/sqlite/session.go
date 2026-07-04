package sqlite

import (
	"database/sql"
	"time"
)

type Session struct {
	Exe         string
	TimeOpened  int64
	TimeFocused int64
	Title       string
}

func GetDay(t time.Time, db *sql.DB) ([]Session, error) {
	now := time.Now()

	var rows *sql.Rows
	var err error

	if t.Day() == now.Day() && t.Month() == now.Month() && t.Year() == now.Year() {
		rows, err = db.Query(`
			select exe, title, sum(time_opened_ms), sum(time_focused_ms)
			from sessions where date(session_date) = date(?) 
			group by exe, title order by sum(time_focused_ms) desc, sum(time_opened_ms) desc;`,
			t.Format("2006-01-02"))
	} else {
		rows, err = db.Query(`
			select exe, title, sum(time_opened_ms), sum(time_focused_ms)
			from sessions_history where date(session_date) = date(?) 
			group by exe, title order by sum(time_focused_ms) desc, sum(time_opened_ms) desc;`,
			t.Format("2006-01-02"))
	}

	if err != nil {
		return []Session{}, err
	}
	defer rows.Close()

	var exe string
	var timeOpened int64
	var timeFocused int64
	var titles string
	var sessions []Session

	for rows.Next() {
		err = rows.Scan(&exe, &titles, &timeOpened, &timeFocused)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, Session{exe, timeOpened, timeFocused, titles})
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return sessions, nil
}
