package mqttGather

import "time"

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
		precipitation INTEGER PRIMARY KEY AUTOINCREMENT,
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

// https://opendata.dwd.de/climate_environment/CDC/observations_germany/climate/10_minutes/wind/recent/10minutenwerte_wind_05490_akt.zip
// STATIONS_ID;MESS_DATUM;  QN;FF_10;DD_10;eor
//       5404;201911300000;    3;   3.9; 250;eor
// FF_10 average windspeed
// DD_10 direction
type Wind struct {
	Observation
	Windspeed float32
	Direction int
}

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
	) VALUE (
		:station, :ts, :wind_speed, :direction
	)
`
