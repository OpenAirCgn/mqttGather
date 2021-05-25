package mqttGather

import (
	"testing"
	"time"
)

func TestInsert(t *testing.T) {
	db, err := NewDatabase("test.sqlite3")
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
