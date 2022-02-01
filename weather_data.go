package mqttGather

import (
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// utilities to enrich the gathered MQTT data with weather statistics
// from DWD https://www.dwd.de/DE/leistungen/klimadatendeutschland/klimadatendeutschland.html

const BASE_DWD_CDC_URL = "https://opendata.dwd.de/climate_environment/CDC/observations_germany/climate/10_minutes/"

// Station information here:
// https://opendata.dwd.de/climate_environment/CDC/observations_germany/climate/10_minutes/
// cologne is station_id 2667
// KZ		ID	ICAO	NAME		alt	LAT	LONG	Automated since since
// 10513	2667	EDDK	Köln-Bonn	92	50° 51'	07° 09'	01.12.1993	1957

type Station string

type Observation struct {
	Station   Station
	Timestamp time.Time
}

// https://opendata.dwd.de/climate_environment/CDC/observations_germany/climate/10_minutes/air_temperature/recent/10minutenwerte_TU_02667_akt.zip
// STATIONS_ID;MESS_DATUM;  QN;PP_10;TT_10;TM5_10;RF_10;TD_10;eor
//        617;201911290000;    3;   -999;   7.4;   6.6;  89.5;   5.8;eor
// PP_10 Luftdruck -- apparently always inalid for 10 minute readings
// TT_10 Lufttemperatur in 2m Hoehe
// TM5_10 Temp in 5cm
// RF_10 relative Feuchtigkeit in 2m
// TD_10 Taupunkt

const (
	TEMP_STATIONS_ID_IDX = iota
	TEMP_MESS_DATUM_IDX
	TEMP_QN_IDX
	TEMP_PP_10_IDX
	TEMP_TT_10_IDX
	TEMP_TM5_10_IDX
	TEMP_RF_10_IDX
	TEMP_TD_10_IDX
	TEMP_eor_IDX

	DATE_FMT      = "200601021504MST"
	TEMP_DATE_FMT = DATE_FMT

	TEMP_URL = "https://opendata.dwd.de/climate_environment/CDC/observations_germany/climate/10_minutes/air_temperature/recent/10minutenwerte_TU_02667_akt.zip"
)

type Temperature struct {
	Observation
	Temp2m     float32
	Temp5cm    float32
	Humidity2m float32
	DewPoint   float32
}

const DB_TABLE_TEMPERATURE = `
	DROP TABLE IF EXISTS temperature;
	CREATE TABLE IF NOT EXISTS temperature (
		temperature_id INTEGER PRIMARY KEY AUTOINCREMENT,
		station        VARCHAR,
		ts             INTEGER,
		temp2m         FLOAT,
		temp5cm        FLOAT,
		humidity2m     FLOAT,
		dewPoint       FLOAT
	)`

const DB_INSERT_TEMPERATURE = `
	INSERT INTO temperature (
		station, ts, temp2m, temp5cm, humidity2m, dewPoint
	) VALUES (
		:station, :ts, :temp2m, :temp5cm, :humidity2m, :dewPoint
	)`

func (t *Temperature) Insert(stmt *sql.Stmt) error {
	_, err := stmt.Exec(
		t.Station,
		t.Timestamp.Unix(),
		t.Temp2m,
		t.Temp5cm,
		t.Humidity2m,
		t.DewPoint,
	)
	return err
}

func ImportTemperature(db_fn string) error {
	insert := func(record []string, stmt *sql.Stmt) error {
		if record[0] == "STATIONS_ID" {
			return nil // header
		}

		var temp Temperature
		var err error
		temp.Station = Station(record[TEMP_STATIONS_ID_IDX])

		if temp.Timestamp, err = time.Parse(TEMP_DATE_FMT, record[TEMP_MESS_DATUM_IDX]+"MEZ"); err != nil {
			return err
		}
		if temp2m, err := strconv.ParseFloat(record[TEMP_TT_10_IDX], 32); err != nil {
			return err
		} else {
			temp.Temp2m = float32(temp2m)
		}
		if temp5cm, err := strconv.ParseFloat(record[TEMP_TM5_10_IDX], 32); err != nil {
			return err
		} else {
			temp.Temp5cm = float32(temp5cm)
		}
		if hum, err := strconv.ParseFloat(record[TEMP_RF_10_IDX], 32); err != nil {
			return err
		} else {
			temp.Humidity2m = float32(hum)
		}
		if dew, err := strconv.ParseFloat(record[TEMP_RF_10_IDX], 32); err != nil {
			return err
		} else {
			temp.DewPoint = float32(dew)
		}

		return temp.Insert(stmt)
	}

	return importDWDData(
		db_fn,
		DB_TABLE_TEMPERATURE,
		DB_INSERT_TEMPERATURE,
		TEMP_URL,
		insert,
	)
}

const PRECIPITATION_URL = "https://opendata.dwd.de/climate_environment/CDC/observations_germany/climate/10_minutes/precipitation/recent/10minutenwerte_nieder_02667_akt.zip"

// STATIONS_ID;MESS_DATUM;  QN;RWS_DAU_10;RWS_10;RWS_IND_10;eor
//       2667;201911290000;    3;  10;   0.13;   1;eor
// RWS_DAU_10 Niederschlagsdauer
// RWS_10 Summe des Niedeschlags
// RWS_IND_10 Niederschlagsindikator

const (
	PRECIP_STATIONS_ID_IDX = iota
	PRECIP_MESS_DATUM_IDX
	PRECIP_QN_IDX
	PRECIP_RWS_DAU_10_IDX
	PRECIP_RWS_10_IDX
	PRECIP_RWS_IND_10_IDX
	PRECIP_eor_IDX

	PRECIP_DATE_FMT = DATE_FMT
)

type Precipitation struct {
	Observation
	Duration10m int
	Sum10m      float32
	Indicator   bool
}

const DB_TABLE_PRECIPITATION = `
	DROP TABLE IF EXISTS       precipitation;
	CREATE TABLE IF NOT EXISTS precipitation (
		precipitation_id INTEGER PRIMARY KEY AUTOINCREMENT,
		station VARCHAR,
		ts INTEGER,
		duration INTEGER,
		sum_10 FLOAT,
		indicator BOOLEAN
	)
`
const DB_INSERT_PRECIPITATION = `
	INSERT INTO precipitation (
		station, ts, duration, sum_10, indicator
	) VALUES (
		:station, :ts, :duration, :sum_10, :indicator
	)
`

func (p *Precipitation) Insert(stmt *sql.Stmt) error {
	_, err := stmt.Exec(
		p.Station,
		p.Timestamp.Unix(),
		p.Duration10m,
		p.Sum10m,
		p.Indicator,
	)
	return err
}
func ImportPrecipitation(db_fn string) error {
	insert := func(record []string, stmt *sql.Stmt) error {
		if record[0] == "STATIONS_ID" {
			return nil // header
		}

		var precip Precipitation
		var err error
		precip.Station = Station(record[PRECIP_STATIONS_ID_IDX])

		if precip.Timestamp, err = time.Parse(PRECIP_DATE_FMT, record[PRECIP_MESS_DATUM_IDX]+"MEZ"); err != nil {
			return err
		}
		if sum10m, err := strconv.ParseFloat(record[PRECIP_RWS_10_IDX], 32); err != nil {
			return err
		} else {
			precip.Sum10m = float32(sum10m)
		}

		return precip.Insert(stmt)
	}

	return importDWDData(
		db_fn,
		DB_TABLE_PRECIPITATION,
		DB_INSERT_PRECIPITATION,
		PRECIPITATION_URL,
		insert,
	)
}

const SOLAR_URL = "https://opendata.dwd.de/climate_environment/CDC/observations_germany/climate/10_minutes/solar/recent/10minutenwerte_SOLAR_02667_akt.zip"

// STATIONS_ID;MESS_DATUM;  QN;DS_10;GS_10;SD_10;LS_10;eor
//       2667;201911290000;    3;   0.0;   0.0;   0.000;-999;eor
// DS_10 diffuse himmelstrahlung Joule
// GS_10 Globalstrahlung Joule
// SD_10 Sunshine duration

const (
	SOLAR_STATIONS_ID_IDX = iota
	SOLAR_MESS_DATUM_IDX
	SOLAR_QN_IDX
	SOLAR_DS_10_IDX
	SOLAR_GS_10_IDX
	SOLAR_SD_10_IDX
	SOLAR_LS_10_IDX
	SOLAR_eor_IDX

	SOLAR_DATEFMT = DATE_FMT
)

type Solar struct {
	Observation
	DiffuseIrradiation float32
	GlobalIrradiation  float32
	Duration10m        float32
}

const DB_TABLE_SOLAR = `
	DROP TABLE IF EXISTS solar;
	CREATE TABLE IF NOT EXISTS solar (
		solar_id INTEGER PRIMARY KEY AUTOINCREMENT,
		station VARCHAR,
		ts INTEGER,
		diffuse FLOAT,
		global FLOAT,
		duration FLOAT
	)
`
const DB_INSERT_SOLAR = `
	INSERT INTO solar (
		station, ts, diffuse, global, duration
	) VALUES (
		:station, :ts, :diffuse, :global, :duration
	)
`

func (w *Solar) Insert(stmt *sql.Stmt) error {
	_, err := stmt.Exec(
		w.Station,
		w.Timestamp.Unix(),
		w.DiffuseIrradiation,
		w.GlobalIrradiation,
		w.Duration10m,
	)
	return err
}

func ImportSolar(db_fn string) error {
	insert := func(record []string, stmt *sql.Stmt) error {
		if record[0] == "STATIONS_ID" {
			return nil // header
		}

		var solar Solar
		var err error
		solar.Station = Station(record[SOLAR_STATIONS_ID_IDX])

		if solar.Timestamp, err = time.Parse(SOLAR_DATEFMT, record[SOLAR_MESS_DATUM_IDX]+"MEZ"); err != nil {
			return err
		}
		if diffuse, err := strconv.ParseFloat(record[SOLAR_DS_10_IDX], 32); err != nil {
			return err
		} else {
			solar.DiffuseIrradiation = float32(diffuse)
		}

		if global, err := strconv.ParseFloat(record[SOLAR_GS_10_IDX], 32); err != nil {
			return err
		} else {
			solar.GlobalIrradiation = float32(global)
		}

		if duration, err := strconv.ParseFloat(record[SOLAR_SD_10_IDX], 32); err != nil {
			return err
		} else {
			solar.Duration10m = float32(duration)
		}

		return solar.Insert(stmt)
	}

	return importDWDData(
		db_fn,
		DB_TABLE_SOLAR,
		DB_INSERT_SOLAR,
		SOLAR_URL,
		insert,
	)
}

const WIND_URL = "https://opendata.dwd.de/climate_environment/CDC/observations_germany/climate/10_minutes/wind/recent/10minutenwerte_wind_02667_akt.zip"

// STATIONS_ID;MESS_DATUM;  QN;FF_10;DD_10;eor
//       5404;201911300000;    3;   3.9; 250;eor
// FF_10 average windspeed
// DD_10 direction
type Wind struct {
	Observation
	WindSpeed float32
	Direction int
}

const (
	WIND_STATIONS_ID_IDX = iota
	WIND_MESS_DATUM_IDX  // YYYYMMDDHH
	WIND_QN_3_IDX        // quality level ...
	WIND_F_IDX           // f is for Force?
	WIND_D_IDX           // Direction?
	WIND_EOR_IDX         // no clue

	WIND_DATE_FMT = DATE_FMT
)

const DB_TABLE_WIND = `
	DROP TABLE IF EXISTS wind;
	CREATE TABLE IF NOT EXISTS wind (
		wind_id INTEGER PRIMARY KEY AUTOINCREMENT,
		station VARCHAR,
		ts INTEGER,
		wind_speed FLOAT,
		direction INTEGER
	)`

const DB_INSERT_WIND = `
	INSERT INTO wind (
		station, ts, wind_speed, direction
	) VALUES (
		:station, :ts, :wind_speed, :direction
	)
`

func (w *Wind) Insert(stmt *sql.Stmt) error {
	_, err := stmt.Exec(
		w.Station,
		w.Timestamp.Unix(),
		w.WindSpeed,
		w.Direction,
	)
	return err
}

func ImportWind(db_fn string) error {
	insert := func(record []string, stmt *sql.Stmt) error {
		if record[0] == "STATIONS_ID" {
			return nil // header
		}

		var wind Wind
		var err error
		wind.Station = Station(record[WIND_STATIONS_ID_IDX])

		if wind.Timestamp, err = time.Parse(WIND_DATE_FMT, record[WIND_MESS_DATUM_IDX]+"MEZ"); err != nil {
			return err
		}
		if windSpeed, err := strconv.ParseFloat(record[WIND_F_IDX], 32); err != nil {
			return err
		} else {
			wind.WindSpeed = float32(windSpeed)
		}
		if direction, err := strconv.ParseInt(record[WIND_D_IDX], 10, 32); err != nil {
			return err
		} else {
			wind.Direction = int(direction)
		}

		return wind.Insert(stmt)
	}

	return importDWDData(
		db_fn,
		DB_TABLE_WIND,
		DB_INSERT_WIND,
		WIND_URL,
		insert,
	)
}

func importDWDData(
	db_fn string,
	db_create_stmt string,
	db_insert_stmt string,
	url string,
	insert func(csvRecord []string, db *sql.Stmt) error) error {
	db, err := sql.Open("sqlite3", db_fn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(db_create_stmt)
	if err != nil {
		panic(err)
	}

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare(db_insert_stmt)
	if err != nil {
		tx.Rollback()
		panic(err)
	}
	defer stmt.Close()

	reader, err := getZippedAsReader(url)
	if err != nil {
		tx.Rollback()
		panic(err)

	}

	csvReader := csv.NewReader(reader)
	csvReader.Comma = ';'
	i := 0
	for {
		println(i)
		i += 1
		record, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break // done
			}
			tx.Rollback()
			panic(err)
		}
		var trimmed []string
		for _, csv_entry := range record {
			trimmed = append(trimmed, strings.TrimSpace(csv_entry))
		}
		if err := insert(trimmed, stmt); err != nil {
			tx.Rollback()
			panic(err)
		}

	}
	return nil
}

// TODO figure out how to close all the intermediate readers.
func getZippedAsReader(url string) (io.Reader, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	tmpFile, err := ioutil.TempFile(os.TempDir(), "dwd-")
	if err != nil {
		return nil, err
	}

	defer os.Remove(tmpFile.Name())

	size, err := io.Copy(tmpFile, resp.Body)
	if err != nil {
		return nil, err
	}

	// go docs are very cavalier about what the size parameter is meant to mean.
	// Do I have to check the lenght before decoding? Or can I just put something random in?

	zipreader, err := zip.NewReader(tmpFile, size)

	for _, f := range zipreader.File {
		fmt.Printf("reading: %s\n", f.Name)
		if fileReader, err := f.Open(); err != nil {
			return nil, err
		} else {
			return fileReader, nil
		}
	}
	return nil, nil // can't happen

}
