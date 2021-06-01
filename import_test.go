package mqttGather

import (
	"database/sql"
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// cheap trick to run winddaten import

const fn = "test_data/produkt_ff_stunde_20191127_20210529_02667.txt"
const dbFn = "windaten.sqlite3"

// sample:
//STATIONS_ID;MESS_DATUM;QN_3;   F;   D;eor
//       2667;2019112700;   10;   4.2; 150;eor

const (
	STATIONS_ID_IDX = iota
	MESS_DATUM_IDX  // YYYYMMDDHH
	QN_3_IDX        // donno
	F_IDX           // f is for Force?
	D_IDX           // Direction?
	EOR_IDX         // no clue
)

const (
	DATE_FMT = "2006010215MST"
)

func TestImportWindDaten(t *testing.T) {

	db, err := sql.Open("sqlite3", dbFn)
	if err != nil {
		t.Fatal(err)
	}

	sql := `DROP TABLE wind; 
		CREATE TABLE wind (
			wind_id INTEGER PRIMARY KEY AUTOINCREMENT,
			f_ms FLOAT,
			dir  INTEGER, -- compass heading
			ts INTEGER -- add unique constraint, location, etc.
		)`

	if _, err := db.Exec(sql); err != nil {
		t.Fatal(err)
	}

	csvF, err := os.Open(fn)
	if err != nil {
		t.Fatal(err)
	}
	defer csvF.Close()

	reader := csv.NewReader(csvF)
	reader.Comma = ';'

	insert := `
		INSERT INTO wind
			(f_ms, dir, ts)
		VALUES
			(:F, :D, :T);
	`
	stmt, err := db.Prepare(insert)
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()

	if _, err := db.Exec("BEGIN TRANSACTION"); err != nil {
		t.Fatal(err)
	}
	defer db.Exec("COMMIT")
	i := 0
	for {
		println(i)
		i += 1
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)

		}
		if record[0] == "STATIONS_ID" {
			continue
		}

		_hour := record[MESS_DATUM_IDX]
		_windSpeed := record[F_IDX]
		_direction := record[D_IDX]

		hour, err := time.Parse(DATE_FMT, _hour+"MEZ")
		if err != nil {
			t.Fatal(err)
		}

		windSpeed, err := strconv.ParseFloat(strings.TrimSpace(_windSpeed), 32)
		if err != nil {
			t.Fatal(err)
		}

		direction, err := strconv.ParseInt(strings.TrimSpace(_direction), 10, 32)

		if _, err := stmt.Exec(windSpeed, direction, hour); err != nil {
			t.Fatal(err)
		}

	}

}
