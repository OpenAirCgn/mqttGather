package opennoise_daemon

// Database mapping of sent alerts
type Alert struct {
	DeviceSignifier string
	Timestamp       int64
	AlertPhone      string
	Message         string
	Status          string
}

