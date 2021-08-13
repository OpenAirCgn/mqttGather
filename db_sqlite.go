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

var deviceCache map[string]int64 = make(map[string]int64)

func (s *SqliteDB) lookupDevice(device_mac string) (deviceId int64, err error) {
	if deviceId, ok := deviceCache[device_mac]; ok {
		return deviceId, nil
	}

	deviceId, err = s.LoadDeviceId(device_mac)

	if err == nil {
		deviceCache[device_mac] = deviceId
		return deviceId, err
	}

	if err != nil && err != sql.ErrNoRows {
		return -1, err
	}

	// an ErrNoRows error indicates we don't have an entry yet.
	sql := `INSERT INTO device ( device_signifier ) VALUES (:DEVICE);`
	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(device_mac)
	if err != nil {
		return -1, err
	}

	if id, err := res.LastInsertId(); err != nil {
		return -1, err
	} else {
		deviceCache[device_mac] = id
		return id, nil
	}
}

func (s *SqliteDB) LoadDeviceId(device_signifier string) (int64, error) {
	sql := "SELECT device_id FROM device WHERE device_signifier = :DEVICE"
	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	var device_id int64
	err = stmt.QueryRow(device_signifier).Scan(&device_id)
	return device_id, err
}

func (s *SqliteDB) Save(stats *DBAStats, t time.Time) (int64, error) {
	device_id, err := s.lookupDevice(stats.Client)
	if err != nil {
		return -1, err
	}

	sql := `INSERT INTO dba_stats (
		device_id, min, max, average, averageVar, mean, num, ts
	) VALUES (
		:DEVICE_ID,:MIN,:MAX,:AVG, :AVG_VAR, :MEAN, :NUM, :TS
	);`

	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		//stats.Client,
		device_id,
		stats.Min,
		stats.Max,
		stats.Average,
		stats.AverageVar,
		stats.Mean,
		stats.Num,
		t.Unix(),
	)
	if err != nil {
		return -1, err
	}
	return res.LastInsertId()

}

func (s *SqliteDB) SaveNow(stats *DBAStats) (int64, error) {
	device_id, err := s.lookupDevice(stats.Client)
	if err != nil {
		return -1, err
	}

	sql := `INSERT INTO dba_stats (
		device_id, min, max, average, averageVar, mean, num
	) VALUES (
		:DEVICE_ID,:MIN,:MAX,:AVG, :AVG_VAR, :MEAN, :NUM
	);`

	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		device_id,
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
	device_id, err := s.lookupDevice(t.Client)
	if err != nil {
		return -1, err
	}
	sql := `INSERT INTO tele_mem (
		device_id, type, free_mem
	) VALUES (
		:DEVICE_ID,:TYPE, :FREE_MEM
	);`

	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		device_id,
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
	device_id, err := s.lookupDevice(t.Client)
	if err != nil {
		return -1, err
	}
	sql := `INSERT INTO tele_ver (
		device_id, type, info
	) VALUES (
		:DEVICE_ID,:TYPE, :INFO
	);`

	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		device_id,
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
	device_id, err := s.lookupDevice(t.Client)
	if err != nil {
		return -1, err
	}

	sql := `INSERT INTO tele_misc (
		device_id, type, data
	) VALUES (
		:DEVICE_ID,:TYPE, :DATA
	);`

	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(
		device_id,
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
	CREATE TABLE IF NOT EXISTS device (
		device_id        INTEGER PRIMARY KEY AUTOINCREMENT,
		device_signifier VARCHAR UNIQUE -- this is the MAC addr of the openoise device
	);

	CREATE TABLE IF NOT EXISTS dba_stats (
		dba_stats_id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id    INTEGER REFERENCES device(device_id),
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
		device_id   INTEGER REFERENCES device(device_id),
		type        VARCHAR,
		free_mem    INTEGER,
		ts          INTEGER DEFAULT (STRFTIME('%s','now'))

	);

	CREATE TABLE IF NOT EXISTS tele_ver (
		tele_ver_id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id    INTEGER REFERENCES device(device_id),
		type        VARCHAR,
		info        VARCHAR,
		ts          INTEGER DEFAULT (STRFTIME('%s','now'))

	);

	CREATE TABLE IF NOT EXISTS tele_misc (
		tele_misc_id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id    INTEGER REFERENCES device(device_id),
		type         VARCHAR,
		data         VARCHAR,
		ts           INTEGER DEFAULT (STRFTIME('%s','now'))
	);
	`
	_, err := db.Exec(sql)
	return err
}
