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
//STATIONS_ID;MESS_DATUM;  QN;PP_10;TT_10;TM5_10;RF_10;TD_10;eor
//        617;201911290000;    3;   -999;   7.4;   6.6;  89.5;   5.8;eor
// PP_10 Luftdruck -- apparently always inalid for 10 minute readings
// TT_10 Lufttemperatur in 2m Hoehe
// TM5_10 Temp in 5cm
// RF_10 relative Feuchtigkeit in 2m
// TD_10 Taupunkt

type Temperature struct {
	Observation
	Temp2m     float32
	Temp5cm    float32
	Humidity2m float32
	DewPoint   float32
}

const DB_TABLE_TEMPERATURE = `
	CREATE TABLE IF NOT EXISTS temperature (
		temperature_id INTEGER PRIMARY KEY AUTOINCREMENT,
		station        VARCHAR,
		ts             INTEGER,
		temp2m         FLOAT,
		temp5cm        FLOAT,
		humidity2m     FLOAT,
		dewPoint       FLOAT
	)
`
const DB_INSERT_TEMPERATURE = `
	INSERT INTO precipitation (
		station, ts, temp2m, temp5cm, humidity2m, dewPoint
	) VALUE (
		:station, :ts, :temp2m, :temp5cm, :humidity2m, :dewPoint
	)
`

// https://opendata.dwd.de/climate_environment/CDC/observations_germany/climate/10_minutes/precipitation/recent/10minutenwerte_nieder_02667_akt.zip

// STATIONS_ID;MESS_DATUM;  QN;RWS_DAU_10;RWS_10;RWS_IND_10;eor
//       2667;201911290000;    3;  10;   0.13;   1;eor
// RWS_DAU_10 Niederschlagsdauer
// RWS_10 Summe des Niedeschlags
// RWS_IND_10 Niederschlagsindikator
type Precipitation struct {
	Observation
	Duration10m int
	Sum10       float32
	Indicator   bool
}

const DB_TABLE_PRECIPITATION = `
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
	) VALUE (
		:station, :ts, :duration, :sum_10, :indicator
	)
`

// STATIONS_ID;MESS_DATUM;  QN;DS_10;GS_10;SD_10;LS_10;eor
//       2667;201911290000;    3;   0.0;   0.0;   0.000;-999;eor
// DS_10 diffuse himmelstrahlung Joule
// GS_10 Globalstrahlung Joule
// SD_10 Sunshine duration

type Solar struct {
	Observation
	DiffuseIrradiation float32
	GlobalIrradiation  float32
	Duration10m        float32
}

const DB_TABLE_SOLAR = `
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
	) VALUE (
		:station, :ts, :diffuse, :global, :duration
	)
`

const WIND_URL = "https://opendata.dwd.de/climate_environment/CDC/observations_germany/climate/10_minutes/wind/recent/10minutenwerte_wind_02667_akt.zip"

// STATIONS_ID;MESS_DATUM;  QN;FF_10;DD_10;eor
//       5404;201911300000;    3;   3.9; 250;eor
// FF_10 average windspeed
// DD_10 direction
type Wind struct {
	Observation
	Windspeed float32
	Direction int
}

const (
	WIND_STATIONS_ID_IDX = iota
	WIND_MESS_DATUM_IDX  // YYYYMMDDHH
	WIND_QN_3_IDX        // quality level ...
	WIND_F_IDX           // f is for Force?
	WIND_D_IDX           // Direction?
	WIND_EOR_IDX         // no clue

	WIND_DATEFMT = "200601021504MST"
)

const DB_TABLE_WIND = `
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
		w.Windspeed,
		w.Direction,
	)
	return err
}

func ImportWind(db_fn string) error {
	db, err := sql.Open("sqlite3", db_fn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(DB_TABLE_WIND)
	if err != nil {
		panic(err)
	}

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Commit()

	stmt, err := tx.Prepare(DB_INSERT_WIND)
	if err != nil {
		tx.Rollback()
		panic(err)
	}
	defer stmt.Close()

	reader, err := getZippedAsReader(WIND_URL)
	if err != nil {
		tx.Rollback()
		panic(err)

	}

	csvReader := csv.NewReader(reader)
	csvReader.Comma = ';'
	i := 0
	var wind Wind
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
		if record[0] == "STATIONS_ID" {
			continue // header
		}

		_station := strings.TrimSpace(record[WIND_STATIONS_ID_IDX])
		_hour := record[WIND_MESS_DATUM_IDX]
		_windSpeed := record[WIND_F_IDX]
		_direction := record[WIND_D_IDX]

		hour, err := time.Parse(WIND_DATEFMT, _hour+"MEZ")
		if err != nil {
			tx.Rollback()
			panic(err)
		}

		windSpeed, err := strconv.ParseFloat(strings.TrimSpace(_windSpeed), 32)
		if err != nil {
			tx.Rollback()
			panic(err)
		}

		direction, err := strconv.ParseInt(strings.TrimSpace(_direction), 10, 32)

		if err != nil {
			tx.Rollback()
			panic(err)
		}

		wind.Station = Station(_station)
		wind.Timestamp = hour
		wind.Windspeed = float32(windSpeed)
		wind.Direction = int(direction)
		if err = wind.Insert(stmt); err != nil {
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
