package opennoise_daemon

type DeviceObserver interface {
  SensorDBAStats(*DBAStats)
  SensorTelemetry(*Telemetry)
}
