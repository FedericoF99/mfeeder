package sqlite

import "database/sql"

type Session struct {
	Exe         string
	TimeOpened  int64
	TimeFocused int64
	Title       string
}

func GetDay(t string, db *sql.DB) ([]Session, error) {
	rows, err := db.Query(`
		select exe, pid, title, sum(time_opened_ms), sum(time_focused_ms)
		from sessions where date(session_date) = date(?) 
		group by exe, pid, title order by sum(time_focused_ms) desc, sum(time_opened_ms) desc;`,
		t)

	if err != nil {
		return []Session{}, err
	}
	defer rows.Close()

	var exe string
	var pid int
	var timeOpened int64
	var timeFocused int64
	var titles string
	var sessions []Session

	for rows.Next() {
		err = rows.Scan(&exe, &pid, &titles, &timeOpened, &timeFocused)
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
