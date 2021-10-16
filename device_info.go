package mqttGather

type DeviceInfo struct {
	DeviceSignifier string
	Description     string
	Latitude        float64
	Longitude       float64
	AlertThreshold  float64
	AlertDuration   float64
	AlertCount      int
	AlertDeadtime   float64
	AlertPhone      string
	AlertActive     bool
	TurnOnTime      int
}
