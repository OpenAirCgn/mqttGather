package opennoise_daemon

import "time"

type DB interface {
	Save(*DBAStats, time.Time) (int64, error)
	SaveNow(*DBAStats) (int64, error)
	SaveTelemetry(*Telemetry, time.Time) (int64, error)
	SaveTelemetryNow(*Telemetry) (int64, error)
	LoadDeviceInfo(signifier string) (*DeviceInfo, error)
	SaveAlert(*Alert) (int64, error)
	LoadLastAlert(signifier string) (*Alert, error)
	GetCountThresholdExceeded(signifier string, seconds int64, threshold float64) (int64, error)
	Close()
	Reconnect() error

}
