package mqttGather

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SqliteDB struct {
	db *sql.DB
}

func NewDatabase(connectString string) (DB, error) {
	db, err := sql.Open("sqlite3", connectString)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	if err := setupDB(db); err != nil {
		db.Close()
		return nil, err
	}

	return &SqliteDB{
		db,
	}, nil

}

func (s *SqliteDB) Save(stats *DBAStats, t time.Time) (int64, error) {
	sql := `INSERT INTO dba_stats (
		client, min, max, average, averageVar, mean, num, ts
	) VALUES (
		:CLIENT,:MIN,:MAX,:AVG, :AVG_VAR, :MEAN, :NUM, :TS
	);`

	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return -1, err
	}

	res, err := stmt.Exec(
		stats.Client,
		stats.Min,
		stats.Max,
		stats.Average,
		stats.AverageVar,
		stats.Mean,
		stats.Num,
		t.Unix(),
	)
	if err != nil {
		return -1, nil
	}
	return res.LastInsertId()

}

func (s *SqliteDB) SaveNow(stats *DBAStats) (int64, error) {

	sql := `INSERT INTO dba_stats (
		client, min, max, average, averageVar, mean, num
	) VALUES (
		:CLIENT,:MIN,:MAX,:AVG, :AVG_VAR, :MEAN, :NUM
	);`

	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return -1, err
	}

	res, err := stmt.Exec(
		stats.Client,
		stats.Min,
		stats.Max,
		stats.Average,
		stats.AverageVar,
		stats.Mean,
		stats.Num,
	)
	if err != nil {
		return -1, err
	}
	return res.LastInsertId()
}

func (s *SqliteDB) Close() {
	s.db.Close()
}

func setupDB(db *sql.DB) error {
	sql := `
	CREATE TABLE IF NOT EXISTS dba_stats (
		dba_stats_id INTEGER PRIMARY KEY AUTOINCREMENT,
		client       VARCHAR,
		min          FLOAT,
		max          FLOAT,
		average      FLOAT,
		averageVar   FLOAT,
		mean         FLOAT,
		num          INTEGER,
		ts           INTEGER DEFAULT (STRFTIME('%s','now'))
	)
	`
	_, err := db.Exec(sql)
	return err
}
