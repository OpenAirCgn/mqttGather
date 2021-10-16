package mqttGather

import (
	"database/sql"
	"testing"
	"time"
)

func TestInsert(t *testing.T) {
	db, err := NewDatabase(":memory:")
	if err != nil {
		t.Fatal(err)
	}
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
	db_, err := NewDatabase(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db_.Close()

	db, _ := db_.(*SqliteDB)
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

	db_, err := NewDatabase(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db_.Close()

	db, _ := db_.(*SqliteDB)

	var doesntexist = "doesntexist"
	info, err := db.LoadDeviceInfo(doesntexist)

	if err == nil || "sql: no rows in result set" != err.Error() {
		t.Fatalf("err: %v err:%v", info, err.Error())
	}

	id, err := db.lookupDevice(doesntexist)

	_, err = db.db.Exec("INSERT INTO device_info (device_id, description, latitude, longitude) VALUES (:ID, 'bla', 1.0, 2.0);", id)

	if err != nil {
		t.Fatalf("%v", err)
	}

	info, err = db.LoadDeviceInfo(doesntexist)

	if err != nil {
		t.Fatalf("err: %v err:%v", info, err.Error())
	}

	if info.Latitude != 1.0 || info.Longitude != 2.0 || info.Description != "bla" {
		t.Fatalf("Loading Info Failed")
	}

}
