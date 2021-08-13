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

	if deviceCache[doesntexist] != 0 {
		t.Fatal("can't happen, sanity check 1")
	}

	id, err = db.lookupDevice(doesntexist)
	if err != nil {
		t.Fatalf("failed to create deviceid: %v", err)
	}

	if id != deviceCache[doesntexist] || id == 0 {
		t.Fatalf("something weird: %d", id)
	}

	id, err = db.lookupDevice(doesntexist)
	if err != nil {
		t.Fatalf("failed cached lookup: %v", err)
	}

	// simulate prexisting device in db
	// and not yet in cache

	delete(deviceCache, doesntexist)

	id, err = db.lookupDevice(doesntexist)
	if err != nil {
		t.Fatalf("failed db/notcache lookup: %v", err)
	}

}
