package mqttGather

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

type SqliteDB struct {
	db          *sql.DB
	deviceCache map[string]int64
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
		make(map[string]int64),
	}, nil

}

type execFunc func(*sql.Stmt) (interface{}, error)

func (s *SqliteDB) execute(sql string, exec execFunc) (interface{}, error) {
	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return exec(stmt)
}

func (s *SqliteDB) insert(sqls string, exec execFunc) (int64, error) {

	result_, err := s.execute(sqls, exec)
	result := result_.(sql.Result)
	if err != nil {
		return -1, err
	}

	return result.LastInsertId()

}

func (s *SqliteDB) insertDevice(signifier string) (int64, error) {

	exec := func(stmt *sql.Stmt) (interface{}, error) {
		return stmt.Exec(
			signifier,
		)
	}
	sql := `INSERT INTO device ( device_signifier ) VALUES (:DEVICE);`
	return s.insert(sql, exec)
}

func (s *SqliteDB) lookupDevice(device_mac string) (deviceId int64, err error) {
	if deviceId, ok := s.deviceCache[device_mac]; ok {
		return deviceId, nil
	}

	deviceId, err = s.LoadDeviceId(device_mac)

	if err == nil {
		s.deviceCache[device_mac] = deviceId
		return deviceId, err
	}

	if err != nil && err != sql.ErrNoRows {
		return -1, err
	}
	// an ErrNoRows error indicates we don't have an entry yet.
	// create one and add to cache.

	if id, err := s.insertDevice(device_mac); err != nil {
		return -1, err
	} else {
		s.deviceCache[device_mac] = id
		return id, nil
	}
}

func (s *SqliteDB) LoadDeviceId(device_signifier string) (int64, error) {
	exec := func(stmt *sql.Stmt) (interface{}, error) {
		var device_id int64
		err := stmt.QueryRow(device_signifier).Scan(&device_id)
		return device_id, err
	}

	sql := "SELECT device_id FROM device WHERE device_signifier = :DEVICE"

	id, err := s.execute(sql, exec)
	return id.(int64), err
}

func (s *SqliteDB) Save(stats *DBAStats, t time.Time) (int64, error) {
	device_id, err := s.lookupDevice(stats.Signifier)
	if err != nil {
		return -1, err
	}

	exec := func(stmt *sql.Stmt) (interface{}, error) {
		return stmt.Exec(
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
	}

	sql := `INSERT INTO dba_stats (
		device_id, min, max, average, averageVar, mean, num, ts
	) VALUES (
		:DEVICE_ID,:MIN,:MAX,:AVG, :AVG_VAR, :MEAN, :NUM, :TS
	);`

	return s.insert(sql, exec)

}

func (s *SqliteDB) SaveNow(stats *DBAStats) (int64, error) {
	device_id, err := s.lookupDevice(stats.Signifier)
	if err != nil {
		return -1, err
	}

	exec := func(stmt *sql.Stmt) (interface{}, error) {
		return stmt.Exec(
			device_id,
			stats.Min,
			stats.Max,
			stats.Average,
			stats.AverageVar,
			stats.Mean,
			stats.Num,
			// NOW is added as the deafult value
		)
	}
	sql := `INSERT INTO dba_stats (
		device_id, min, max, average, averageVar, mean, num
	) VALUES (
		:DEVICE_ID,:MIN,:MAX,:AVG, :AVG_VAR, :MEAN, :NUM
	);`

	return s.insert(sql, exec)

}

func (s *SqliteDB) SaveTelemetry(te *Telemetry, ti time.Time) (int64, error) {
	panic("not implemented")
}
func (s *SqliteDB) saveMemory(t *Telemetry) (int64, error) {
	device_id, err := s.lookupDevice(t.Client)
	if err != nil {
		return -1, err
	}
	exec := func(stmt *sql.Stmt) (interface{}, error) {
		return stmt.Exec(
			device_id,
			t.Type,
			t.Data,
			// NOW is added as the deafult value
		)
	}
	sql := `INSERT INTO tele_mem (
		device_id, type, free_mem
	) VALUES (
		:DEVICE_ID,:TYPE, :FREE_MEM
	);`

	return s.insert(sql, exec)
}

func (s *SqliteDB) saveVersion(t *Telemetry) (int64, error) {
	device_id, err := s.lookupDevice(t.Client)
	if err != nil {
		return -1, err
	}
	exec := func(stmt *sql.Stmt) (interface{}, error) {
		return stmt.Exec(
			device_id,
			t.Type,
			t.Data,
			// NOW is added as the default value
		)
	}
	sql := `INSERT INTO tele_ver (
		device_id, type, info
	) VALUES (
		:DEVICE_ID,:TYPE, :INFO
	);`

	return s.insert(sql, exec)
}

func (s *SqliteDB) saveMisc(t *Telemetry) (int64, error) {
	device_id, err := s.lookupDevice(t.Client)
	if err != nil {
		return -1, err
	}

	exec := func(stmt *sql.Stmt) (interface{}, error) {
		return stmt.Exec(
			device_id,
			t.Type,
			t.Data,
			// NOW is added as the default value
		)

	}
	sql := `INSERT INTO tele_misc (
		device_id, type, data
	) VALUES (
		:DEVICE_ID,:TYPE, :DATA
	);`

	return s.insert(sql, exec)
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

func (s *SqliteDB) SaveAlert(alert *Alert) (int64, error) {
	exec := func(stmt *sql.Stmt) (interface{}, error) {
		return stmt.Exec(
			alert.DeviceSignifier,
			alert.Timestamp,
			alert.Message,
			alert.Status,
		)
	}
	sql := `INSERT INTO alert
			(device_id, alert_phone, message, status)
		VALUES
			( (SELECT DISTINCT device_id FROM device WHERE device_signifier = :SIGNIFIER),
			  :PHONE,
			  :MESSAGE,
			  :STATUS
			)
		`
	return s.insert(sql, exec)
}

func (s *SqliteDB) LoadDeviceInfo(signifier string) (*DeviceInfo, error) {
	exec := func(stmt *sql.Stmt) (interface{}, error) {
		var info DeviceInfo

		err := stmt.QueryRow(signifier).Scan(
			&info.DeviceSignifier,
			&info.Description,
			&info.Latitude,
			&info.Longitude,
			&info.AlertThreshold,
			&info.AlertDuration,
			&info.AlertCount,
			&info.AlertDeadtime,
			&info.AlertPhone,
			&info.AlertActive,
			&info.TurnOnTime,
		)
		return &info, err
	}

	sql := `
SELECT
	d.device_signifier,
	description,
	latitude,
	longitude,
	alert_threshold,
	alert_duration,
	alert_count,
	alert_deadtime,
	alert_phone,
	alert_active,
	turn_on_time

FROM
	device_info di
JOIN
	device d
ON
	di.device_id = d.device_id
WHERE
	device_signifier = :SIGNIFIER
`
	info_, err := s.execute(sql, exec)
	return info_.(*DeviceInfo), err
}

func (s *SqliteDB) LoadLastAlert(signifier string) (*Alert, error) {
	exec := func(stmt *sql.Stmt) (interface{}, error) {
		var alert Alert
		err := stmt.QueryRow(signifier).Scan(
			&alert.DeviceSignifier,
			&alert.Timestamp,
			&alert.AlertPhone,
			&alert.Message,
			&alert.Status,
		)
		return &alert, err
	}

	sql := `
SELECT
	d.device_signifier,
	ts,
	alert_phone,
	message,
	status
FROM
	alert a
JOIN
	device d
ON
	d.device_id = a.device_id
WHERE
	d.device_signifier = :SIGNIFIER
ORDER BY
	ts
DESC
LIMIT 1
`
	alert_, err := s.execute(sql, exec)
	return alert_.(*Alert), err
}

func (s *SqliteDB) GetCountThresholdExceeded(signifier string, seconds int64, threshold float64) (int64, error) {
	return s.getCountThresholdExceeded(
		signifier,
		time.Now().Unix()-seconds,
		threshold)
}
func (s *SqliteDB) getCountThresholdExceeded(signifier string, windowBeginTS int64, threshold float64) (int64, error) {
	exec := func(stmt *sql.Stmt) (interface{}, error) {
		var count int64
		err := stmt.QueryRow(windowBeginTS, signifier, threshold).Scan(&count)
		return count, err
	}
	sql := `
SELECT
	count(*)
FROM
	dba_stats s
JOIN
	device d
ON
	s.device_id = d.device_id
WHERE
	ts > :SECONDS
AND
	d.device_signifier = :SIGNIFIER
AND
	s.max > :THRESHOLD
	`

	id_, err := s.execute(sql, exec)
	return id_.(int64), err
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

	CREATE TABLE IF NOT EXISTS device_info (
		deviceinfo_id   INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id       INTEGER NOT NULL UNIQUE REFERENCES device(device_id),
		description     VARCHAR NOT NULL DEFAULT 'Unbekanntes Geraet', 
		latitude        FLOAT NOT NULL,                 
		longitude       FLOAT NOT NULL,                 
		alert_threshold  FLOAT NOT NULL DEFAULT 100,     
		alert_duration   FLOAT NOT NULL DEFAULT 60,      
		alert_count      INTEGER NOT NULL DEFAULT 3,     
		alert_deadtime   FLOAT NOT NULL DEFAULT 1800,    
		alert_phone      VARCHAR NOT NULL DEFAULT "",    
		alert_active     BOOLEAN NOT NULL DEFAULT FALSE, 
		turn_on_time      INTEGER NOT NULL DEFAULT 0      
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

	-- log of outgoing alerts
	CREATE TABLE IF NOT EXISTS alert (
		alert_id    INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id   INTEGER REFERENCES device(device_id),
		ts          INTEGER DEFAULT (STRFTIME('%s','now')),
		alert_phone  VARCHAR,
		message     VARCHAR,
		status      VARCHAR
	);

	`
	_, err := db.Exec(sql)
	return err
}
