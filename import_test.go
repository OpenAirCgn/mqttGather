package l

import (
	"database/sql"
	"testing"
)

// cheap trick to run winddaten import

const fn = "test_data/produkt_ff_stunde_20191127_20210529_02667.txt"
const dbFn = "windaten.sqlite3"

func TestImportWindDaten(t *testing.T) {

	db, err := sql.Open("sqlite3", dbFn)
	if err != nil {
		t.Fatal(err)
	}

	sql := `CREATE TABLE wind (
		wind_id INTEGER PRIMARY KEY AUTOINCREMENT,
		f_ms FLOAT,
		dir  INTEGER, -- compass heading
		ts INTEGER
)`

	if _, err := db.Exec(sql); err != nil {
		t.Fatal(err)
	}

}
