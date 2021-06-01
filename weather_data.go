package mqttGather

// utilities to enrich the gathered MQTT data with weather statistics
// from DWD https://www.dwd.de/DE/leistungen/klimadatendeutschland/klimadatendeutschland.html


const BASE_DWD_CDC_URL = 'https://opendata.dwd.de/climate_environment/CDC/observations_germany/climate/10_minutes/'

// Station information here:
// https://opendata.dwd.de/climate_environment/CDC/observations_germany/climate/10_minutes/
// cologne is station_id 2667
// KZ		ID	ICAO	NAME		alt	LAT	LONG	Automated since since
// 10513	2667	EDDK	Köln-Bonn	92	50° 51'	07° 09'	01.12.1993	1957

// https://opendata.dwd.de/climate_environment/CDC/observations_germany/climate/10_minutes/air_temperature/recent/10minutenwerte_TU_02667_akt.zip
//STATIONS_ID;MESS_DATUM;  QN;PP_10;TT_10;TM5_10;RF_10;TD_10;eor
//        617;201911290000;    3;   -999;   7.4;   6.6;  89.5;   5.8;eor
// PP_10 Luftdruck -- apparently always inalid for 10 minute readings
// TT_10 Lufttemperatur in 2m Hoehe
// TM5_10 Temp in 5cm 
// RF_10 relative Feuchtigkeit in 2m
// TD_10 Taupunkt
type Temperatur struct {}
// https://opendata.dwd.de/climate_environment/CDC/observations_germany/climate/10_minutes/precipitation/recent/10minutenwerte_nieder_02667_akt.zip

// STATIONS_ID;MESS_DATUM;  QN;RWS_DAU_10;RWS_10;RWS_IND_10;eor
//       2667;201911290000;    3;  10;   0.13;   1;eor
// RWS_DAU_10 Niederschlagsdauer
// RWS_10 Summe des Niedeschlags
// RWS_IND_10 Niederschlagsindikator
type Precipitation struct {}
// STATIONS_ID;MESS_DATUM;  QN;DS_10;GS_10;SD_10;LS_10;eor
//       2667;201911290000;    3;   0.0;   0.0;   0.000;-999;eor
type Solar struct {}
type Wind struct {}


