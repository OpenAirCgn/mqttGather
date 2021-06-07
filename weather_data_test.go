package mqttGather

import (
	"io"
	"os"
	"testing"
)

func TestWeatherDataDownload(t *testing.T) {

	url := "https://opendata.dwd.de/climate_environment/CDC/observations_germany/climate/10_minutes/air_temperature/recent/10minutenwerte_TU_02667_akt.zip"

	reader, err := getZippedAsReader(url)
	if err != nil {
		t.Fatal(err)
	}
	//defer reader.Close()
	io.Copy(os.Stdout, reader)

}

func TestImportWind(t *testing.T) {
	err := ImportWind("weather.sqlite3")
	if err != nil {
		t.Fatal(err)
	}
}
func TestImportSolar(t *testing.T) {
	err := ImportSolar("weather.sqlite3")
	if err != nil {
		t.Fatal(err)
	}
}
func TestImportTemperature(t *testing.T) {
	err := ImportTemperature("weather.sqlite3")
	if err != nil {
		t.Fatal(err)
	}
}
func TestImportPrecipitation(t *testing.T) {
	err := ImportPrecipitation("weather.sqlite3")
	if err != nil {
		t.Fatal(err)
	}
}
