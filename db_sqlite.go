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
		// NOW is added as the deafult value
	)
	if err != nil {
		return -1, err
	}
	return res.LastInsertId()
}

func (s *SqliteDB) SaveTelemetry(te *Telemetry, ti time.Time) (int64, error) {
	panic("not implemented")
}
func (s *SqliteDB) saveMemory(t *Telemetry) (int64, error) {
	sql := `INSERT INTO tele_mem (
		client, type, free_mem
	) VALUES (
		:CLIENT,:TYPE, :FREE_MEM
	);`

	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return -1, err
	}

	res, err := stmt.Exec(
		t.Client,
		t.Type,
		t.Data,
		// NOW is added as the deafult value
	)
	if err != nil {
		return -1, err
	}
	return res.LastInsertId()

}
func (s *SqliteDB) saveVersion(t *Telemetry) (int64, error) {
	sql := `INSERT INTO tele_ver (
		client, type, info
	) VALUES (
		:CLIENT,:TYPE, :INFO
	);`

	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return -1, err
	}

	res, err := stmt.Exec(
		t.Client,
		t.Type,
		t.Data,
		// NOW is added as the default value
	)
	if err != nil {
		return -1, err
	}
	return res.LastInsertId()

}
func (s *SqliteDB) saveMisc(t *Telemetry) (int64, error) {
	sql := `INSERT INTO tele_misc (
		client, type, data
	) VALUES (
		:CLIENT,:TYPE, :DATA
	);`

	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return -1, err
	}

	res, err := stmt.Exec(
		t.Client,
		t.Type,
		t.Data,
		// NOW is added as the default value
	)
	if err != nil {
		return -1, err
	}
	return res.LastInsertId()

}
func (s *SqliteDB) SaveTelemetryNow(t *Telemetry) (int64, error) {
	if t.IsMemory() {
		return s.saveMemory(t)
	} else if t.IsVersion() {
		return s.saveVersion(t)
	} else {
		return s.saveMisc(t)
		// IsSignalQuality() bool { return t.Type.IsSignalQuality() } ** TODO
		// IsFlag() bool          { return t.Type.IsFlag() }
		// IsResetReason() bool   { return t.Type.IsResetReason() }
		// unknown
	}
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
	);

	CREATE TABLE IF NOT EXISTS tele_mem (
		tele_mem_id INTEGER PRIMARY KEY AUTOINCREMENT,
		client      VARCHAR,
		type        VARCHAR,
		free_mem    INTEGER,
		ts          INTEGER DEFAULT (STRFTIME('%s','now'))

	);

	CREATE TABLE IF NOT EXISTS tele_ver (
		tele_ver_id INTEGER PRIMARY KEY AUTOINCREMENT,
		client      VARCHAR,
		type        VARCHAR,
		info        VARCHAR,
		ts          INTEGER DEFAULT (STRFTIME('%s','now'))

	);

	CREATE TABLE IF NOT EXISTS tele_misc (
		tele_misc_id INTEGER PRIMARY KEY AUTOINCREMENT,
		client       VARCHAR,
		type         VARCHAR,
		data         VARCHAR,
		ts           INTEGER DEFAULT (STRFTIME('%s','now'))
	);
	`
	_, err := db.Exec(sql)
	return err
}
