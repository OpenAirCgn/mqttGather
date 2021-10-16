package mqttGather

//	-- log of outgoing alerts
//	CREATE TABLE IF NOT EXISTS alert (
//		alert_id    INTEGER PRIMARY KEY AUTOINCREMENT,
//		device_id   INTEGER REFERENCES device(device_id),
//		ts          INTEGER DEFAULT (STRFTIME('%s','now')),
//		alert_phone  VARCHAR,
//		message     VARCHAR,
//		status      VARCHAR
//	);
type Alert struct {
	DeviceSignifier string
	Timestamp       int64
	AlertPhone      string
	Message         string
	Status          string
}
