package opennoise_daemon

type DeviceInfo struct {
	DeviceSignifier string
	Description     string
	Latitude        float64
	Longitude       float64
	AlertThreshold  float64
	AlertDuration   int64
	AlertCount      int64
	AlertDeadtime   int64
	AlertPhone      string
	AlertActive     bool
	TurnOnTime      int64
}
