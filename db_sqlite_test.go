package mqttGather

import (
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func getTestDB(t *testing.T) *SqliteDB {
	db_, err := NewDatabase(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	db, _ := db_.(*SqliteDB)
	return db
}

const TEST_SIGNIFIER = "aa:bb:cc:dd:ee;ff"

func getTestDBWithDevice(t *testing.T) (*SqliteDB, int64) {
	db := getTestDB(t)
	id, _ := db.lookupDevice(TEST_SIGNIFIER)
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
	if id != 0 || err != sql.ErrNoRows {
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

	_, err := db.db.Exec("INSERT INTO device_info (device_id, description, latitude, longitude) VALUES (:ID, 'bla', 1.0, 2.0);", id)

	if err != nil {
		t.Fatalf("%v", err)
	}

	info, err := db.LoadDeviceInfo(TEST_SIGNIFIER)

	if err != nil {
		t.Fatalf("err: %v err:%v", info, err.Error())
	}

	if info.Latitude != 1.0 || info.Longitude != 2.0 || info.Description != "bla" {
		t.Fatalf("Loading Info Failed")
	}

}

func TestLoadAlert(t *testing.T) {

	db, id := getTestDBWithDevice(t)
	defer db.Close()

	for i := 0; i != 3; i += 1 {
		_, err := db.db.Exec("INSERT INTO alert (device_id, ts, alert_phone, message, status) VALUES (:ID, :ts, '123', 'MSG', 'ok' )", id, fmt.Sprintf("%d", i))

		if err != nil {
			t.Fatalf("sanity: %v", err)
		}
	}

	alert, err := db.LoadLastAlert(TEST_SIGNIFIER)
	if err != nil {
		t.Fatalf("could not load last alert: %v", err)
	}
	if alert.Timestamp != 2 {
		t.Fatalf("did not load final alert: %#v", alert)
	}
}
