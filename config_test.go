package mqttGather

import "testing"

func TestLoad(t *testing.T) {
	rc, err := LoadFromFile("test_data/config.json")
	if err != nil {
		t.Fatal(err)
	}
	if rc.SqlLiteConnect != ":memory:" {
		t.Fatalf("wrong connect string: %s", rc.SqlLiteConnect)
	}
	if rc.Host != "tcp://test.mosquitto.org:1883" {
		t.Fatalf("wrong host: %s", rc.Host)

	}
	if rc.Topic != "/opennoise/+/dba_stats" {
		t.Fatalf("wrong topic: %s", rc.Topic)

	}
	if rc.TelemetryTopic != "/opennoise/+/telemetry" {
		t.Fatalf("wrong telemetry topic: %s", rc.Topic)

	}
	if rc.ClientId != "mqttTest" {
		t.Fatalf("wrong clientId: %s", rc.ClientId)

	}
}
