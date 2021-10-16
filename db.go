package mqttGather

import "time"

type DB interface {
	Save(*DBAStats, time.Time) (int64, error)
	SaveNow(*DBAStats) (int64, error)
	SaveTelemetry(*Telemetry, time.Time) (int64, error)
	SaveTelemetryNow(*Telemetry) (int64, error)
	LoadDeviceInfo(string) (*DeviceInfo, error)
	LoadLastAlert(string) (*Alert, error)
	Close()
}
