package opennoise_daemon

import (
	"encoding/json"
	"io"
	"os"
	"errors"
)

type RunConfig struct {
	SqlLiteConnect     string `json:"sqlite"`
	MqttHost           string `json:"mqtt_host"`
	MqttClientId       string `json:"mqtt_client_id"`
	MqttNoiseTopic     string `json:"mqtt_noise_topic"`
	MqttTelemetryTopic string `json:"mqtt_telemetry_topic"`
	LogDir             string `json:"log_dir"`
	SMSKey             string `json:"sms_key"`
}

// Load a runtime configuration using a reader
func Load(reader io.Reader) (*RunConfig, error) {
	decoder := json.NewDecoder(reader)
	var cfg RunConfig
	err := decoder.Decode(&cfg)
	return &cfg, err
}

// Load a runtime configuration from a file
func LoadFromFile(fn string) (*RunConfig, error) {
	if file, err := os.Open(fn); err != nil {
		return nil, err
	} else {
		return Load(file)
	}
}

// Checks for basic completeness of config. Only tests
// for absolutely necessary params
func (rc RunConfig) Check() error {
	if (rc.SqlLiteConnect == "") {
		return errors.New("sqlite config missing")
	}
	if (rc.MqttHost == "") {
		return errors.New("MQTT host missing")
	}
	if (rc.MqttNoiseTopic == "") {
		return errors.New("MQTT noise topic missing")
	}
	if (rc.MqttTelemetryTopic == "") {
		return errors.New("MQTT telemetry topic missing")
	}
	if (rc.MqttClientId == "") {
		return errors.New("MQTT client ID missing")
	}
	if (rc.SMSKey == "") {
		return errors.New("No SMS Key")
	}
	return nil
}