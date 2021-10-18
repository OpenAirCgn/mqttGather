package mqttGather

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
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

// SQL Helper functions
// the following functions are intended to cut down on/ centralize
// sql boilerplate code.
// These functions assume:
// - on open db connection
// - that a sql statement is compiled to a PreparedStatement
// - which is either an INSERT or a SELECT

// `execFunction`s carry out the work needed to be done
// with the compiled PreparedStatement, i.e. extract results
// or Scan values into an object.
type execFunc func(*sql.Stmt) (interface{}, error)

// Compiles the passed SQL statement to a PreparedStatement which
// is passed off to the provided execFunc and closed once that
// function returns.
func (s *SqliteDB) execute(sql string, exec execFunc) (interface{}, error) {
	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return exec(stmt)
}

// Same as `execute, but assumes that each INSERT statement is assigned
// an automated primary key which is retrieved via Result.LastInsertId()
func (s *SqliteDB) insert(sqls string, exec execFunc) (int64, error) {
	result_, err := s.execute(sqls, exec)
	result := result_.(sql.Result)
	if err != nil {
		return -1, err
	}

	return result.LastInsertId()
}

// 'Public' laoding of DeviceId given a signifier. If no such mapping exists
// this funciton returns an error.
// TODO: creating device mappings needs to be rethought. Currently mappings
// are created the first time a new device is encountered using the private `loadDevice`
// function, see below.

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

func (s *SqliteDB) insertDevice(signifier string) (int64, error) {
	exec := func(stmt *sql.Stmt) (interface{}, error) {
		return stmt.Exec(
			signifier,
		)
	}
	sql := `INSERT INTO device ( device_signifier ) VALUES (:DEVICE);`
	return s.insert(sql, exec)
}

// Persist Stats to DB.
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

// Persists Stats to DB using the current time as the timestamp.
// This is the usual mode of saving as we have no idea when the sample originated
// only when it was received.
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

func (s *SqliteDB) SaveTelemetryNow(t *Telemetry) (int64, error) {
	return s.SaveTelemetry(t, time.Now())
}
func (s *SqliteDB) SaveTelemetry(t *Telemetry, ti time.Time) (int64, error) {
	if t.IsMemory() {
		return s.saveMemory(t, ti)
	} else if t.IsVersion() {
		return s.saveVersion(t, ti)
	} else {
		return s.saveMisc(t, ti)
		// IsSignalQuality() bool { return t.Type.IsSignalQuality() } ** TODO
		// IsFlag() bool          { return t.Type.IsFlag() }
		// IsResetReason() bool   { return t.Type.IsResetReason() }
		// unknown
	}
}
func (s *SqliteDB) saveMemory(t *Telemetry, ti time.Time) (int64, error) {
	device_id, err := s.lookupDevice(t.Client)
	if err != nil {
		return -1, err
	}
	exec := func(stmt *sql.Stmt) (interface{}, error) {
		return stmt.Exec(
			device_id,
			t.Type,
			t.Data,
			ti.Unix(),
		)
	}
	sql := `INSERT INTO tele_mem (
		device_id, type, free_mem, ts
	) VALUES (
		:DEVICE_ID,:TYPE, :FREE_MEM, TS
	);`

	return s.insert(sql, exec)
}

func (s *SqliteDB) saveVersion(t *Telemetry, ti time.Time) (int64, error) {
	device_id, err := s.lookupDevice(t.Client)
	if err != nil {
		return -1, err
	}
	exec := func(stmt *sql.Stmt) (interface{}, error) {
		return stmt.Exec(
			device_id,
			t.Type,
			t.Data,
			ti.Unix(),
		)
	}
	sql := `INSERT INTO tele_ver (
		device_id, type, info, ts
	) VALUES (
		:DEVICE_ID,:TYPE, :INFO, :TS
	);`

	return s.insert(sql, exec)
}

func (s *SqliteDB) saveMisc(t *Telemetry, ti time.Time) (int64, error) {
	device_id, err := s.lookupDevice(t.Client)
	if err != nil {
		return -1, err
	}

	exec := func(stmt *sql.Stmt) (interface{}, error) {
		return stmt.Exec(
			device_id,
			t.Type,
			t.Data,
			ti.Unix(),
		)

	}
	sql := `INSERT INTO tele_misc (
		device_id, type, data, ts
	) VALUES (
		:DEVICE_ID,:TYPE, :DATA, :TS
	);`

	return s.insert(sql, exec)
}

// Save an Alert (typically SMS) we sent in response to a violation.
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

// Load Configuration information for a device
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

// Load the latest Alert (typically SMS) sent to a device.
// generally to control dead times.
// TODO: possibly regard "dead time" in query ?
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

// Retrieve the number of times the `threshold` (value MAX) was exceeded by device `signifier` in the last `seconds` seconds.
func (s *SqliteDB) GetCountThresholdExceeded(signifier string, seconds int64, threshold float64) (int64, error) {
	return s.getCountThresholdExceeded(
		signifier,
		time.Now().Unix()-seconds,
		threshold)
}

// Retrieve the numnber of times the 'threshold' was exceeded by device `signifier` since the timestamp `windowBginTS`.
// This (redundant) function exists primarily to facilitate testing.
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
	if err != nil {
		fmt.Printf("%v", err)
	}
	return id_.(int64), err
}

// Closes the underlying database connection.
func (s *SqliteDB) Close() {
	s.db.Close()
}

// Creates all necessary database objects.
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
