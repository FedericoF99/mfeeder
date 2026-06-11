package sqlite

import (
	"database/sql"
	"fmt"
	"log"
)

const schema = `
	CREATE TABLE IF NOT EXISTS sessions
	(
		id integer primary key autoincrement,
		
		pid integer not null,
		exe text not null,
		title text not null,
		
		start timestamp not null,
		end timestamp not null
	);
	
	--create index if not exist start_index on sessions(start);
`

func Init() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "mfeeder.db")
	if err != nil {
		return nil, fmt.Errorf("database init failed: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("database connection failed", err)
	}

	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}

	return db, nil
}
