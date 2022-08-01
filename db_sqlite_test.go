package opennoise_daemon

import (
	"database/sql"
	"fmt"
	"math"
	"testing"
	"time"
)

func getTestDB(t *testing.T) *SqliteDB {
	db_, err := NewDatabase("file::memory:?cache=private")
	if err != nil {
		t.Fatal(err)
	}
	db, _ := db_.(*SqliteDB)
	db.db.SetMaxOpenConns(1) // <- this is necessary to avoid some extremly tedious racy conditions
	// which may or may not be a driver bug for in-memory dbs during testing.

	return db
}

const TEST_SIGNIFIER = "aa:bb:cc:dd:ee:ff"

func getTestDBWithDevice(t *testing.T) (*SqliteDB, int64) {
	db := getTestDB(t)
	id, _ := db.lookupDevice(TEST_SIGNIFIER)
	return db, id
}

func getTestDBWithDeviceInfo(t *testing.T) (*SqliteDB, int64) {
	db, id := getTestDBWithDevice(t)
	_, err := db.db.Exec(`INSERT INTO 
				device_info (
					device_id, description, latitude, longitude
				) VALUES (
					:ID, 'bla', 1.0, 2.0);`, id)

	if err != nil {
		t.Fatal(err)
	}
	return db, id
}

func TestInsert(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	stats := RandomDBAStats()
	i, err := db.SaveNow(&stats)
	if err != nil {
		t.Fatal(err)
	}
	if i == -1 {
		t.Fatal("i==-1")
	}

	i, err = db.Save(&stats, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if i == -1 {
		t.Fatal("i==-1")
	}

}

func TestDeviceId(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	var doesntexist = "doesntexist"

	id, err := db.LoadDeviceId(doesntexist)
	if id != -1 || err != sql.ErrNoRows {
		t.Fatalf("can't happen, sanity check 1. id(%d) err(%v)", id, err)
	}

	if db.deviceCache[doesntexist] != 0 {
		t.Fatal("can't happen, sanity check 1")
	}

	id, err = db.lookupDevice(doesntexist)
	if err != nil {
		t.Fatalf("failed to create deviceid: %v", err)
	}

	if id != db.deviceCache[doesntexist] || id == 0 {
		t.Fatalf("something weird: %d", id)
	}

	id, err = db.lookupDevice(doesntexist)
	if err != nil {
		t.Fatalf("failed cached lookup: %v", err)
	}

	// simulate prexisting device in db
	// and not yet in cache

	delete(db.deviceCache, doesntexist)
	id, err = db.lookupDevice(doesntexist)
	if err != nil {
		t.Fatalf("failed db/notcache lookup: %v", err)
	}

}

func TestLoadDeviceInfo(t *testing.T) {
	db, id := getTestDBWithDevice(t)
	defer db.Close()

	info, err := db.LoadDeviceInfo(TEST_SIGNIFIER)

	if info != nil {
		t.Fatalf("loaded nonexisting device info: %v", info)
	}
	if err == nil {
		t.Fatalf("expected err!")
	}

	_, err = db.db.Exec("INSERT INTO device_info (device_id, description, latitude, longitude) VALUES (:ID, 'bla', 1.0, 2.0);", id)

	if err != nil {
		t.Fatalf("%v", err)
	}

	info, err = db.LoadDeviceInfo(TEST_SIGNIFIER)

	if err != nil {
		t.Fatalf("err: %v err:%v", info, err.Error())
	}

	if info.Latitude != 1.0 || info.Longitude != 2.0 || info.Description != "bla" {
		t.Fatalf("Loading Info Failed")
	}

}

func TestLoadDeviceInfoFail(t *testing.T) {}

func TestLoadAlert(t *testing.T) {

	db, id := getTestDBWithDevice(t)
	defer db.Close()

	alert, err := db.LoadLastAlert(TEST_SIGNIFIER)

	if err != nil || alert.Timestamp != 0 {
		t.Fatalf("loaded non existant alert (err=%v): %#v", err, alert)
	}

	for i := 0; i != 3; i += 1 {
		_, err := db.db.Exec("INSERT INTO alert (device_id, ts, alert_phone, message, status) VALUES (:ID, :ts, '123', 'MSG', 'ok' )", id, fmt.Sprintf("%d", i))

		if err != nil {
			t.Fatalf("sanity: %v", err)
		}
	}

	alert, err = db.LoadLastAlert(TEST_SIGNIFIER)
	if err != nil {
		t.Fatalf("could not load last alert: %v", err)
	}
	if alert.Timestamp != 2 {
		t.Fatalf("did not load final alert: %#v", alert)
	}
}

func testThresholdExceeded(t *testing.T, db *SqliteDB, windowsSeconds int64, threshold float64, countShould int64) {
	cnt, err := db.getCountThresholdExceeded(TEST_SIGNIFIER, windowsSeconds, threshold)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if cnt != countShould { // first entry doesn't exceed
		t.Fatalf("incorrect count for (w: %v thresh: %v): %v (should: %v)", windowsSeconds, threshold, cnt, countShould)
	}

}

func TestThresholdExceeded(t *testing.T) {
	db, id := getTestDBWithDevice(t)
	defer db.Close()

	sql := `
INSERT INTO dba_stats
	(device_id, ts, max)
VALUES
	(:ID, :TS, :MAX)
`

	for i := 0; i != 10; i += 1 {
		_, err := db.db.Exec(sql, id, i, i)
		if err != nil {
			t.Fatalf("%v", err)
		}
	}
	var should = [][]int64{
		{9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
		{8, 8, 7, 6, 5, 4, 3, 2, 1, 0},
		{7, 7, 7, 6, 5, 4, 3, 2, 1, 0},
		{6, 6, 6, 6, 5, 4, 3, 2, 1, 0},
		{5, 5, 5, 5, 5, 4, 3, 2, 1, 0},
		{4, 4, 4, 4, 4, 4, 3, 2, 1, 0},
		{3, 3, 3, 3, 3, 3, 3, 2, 1, 0},
		{2, 2, 2, 2, 2, 2, 2, 2, 1, 0},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	for window := 0; window <= 9; window++ {
		for threshold := 0.0; threshold <= 9.5; threshold += 0.5 {
			idx2 := int(math.Floor(threshold))
			shouldCount := should[window][idx2]
			testThresholdExceeded(t, db, int64(window), threshold, shouldCount)
		}
	}

}

func TestSaveAlert(t *testing.T) {
	msg := "TestMsg"
	alert := &Alert{
		TEST_SIGNIFIER,
		0,
		"123",
		msg,
		"ok",
	}

	db, _ := getTestDBWithDevice(t)
	db.SaveAlert(alert)

	alert, err := db.LoadLastAlert(TEST_SIGNIFIER)

	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}
	if alert.DeviceSignifier != TEST_SIGNIFIER {
		t.Fatalf("0 is: %s, should %s", alert.DeviceSignifier, TEST_SIGNIFIER)
	}

	if alert.Message != msg {
		t.Fatalf("1 is: %s, should %s", alert.Message, msg)
	}

}
