package mqttGather

import (
	"encoding/json"
	"io"
	"os"
)

type RunConfig struct {
	SqlLiteConnect string `json:"sqlite"`
	Host           string `json:"host"`
	Topic          string `json:"topic"`
	TelemetryTopic string `json:"telemetry_topic"`
	ClientId       string `json:"client_id"`
	LogDir         string `json:"log_dir"`
	SMSKey         string `json:"sms_key"`
}

func Load(reader io.Reader) (*RunConfig, error) {
	decoder := json.NewDecoder(reader)
	var cfg RunConfig
	err := decoder.Decode(&cfg)
	return &cfg, err
}

func LoadFromFile(fn string) (*RunConfig, error) {
	if file, err := os.Open(fn); err != nil {
		return nil, err
	} else {
		return Load(file)
	}
}
