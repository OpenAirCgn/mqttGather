package mqttGather

import "testing"

func TestTelemetryFromPayload(t *testing.T) {
	str := "tme:Jul 31 202119:24:44"
	tel, err := TelemetryFromPayload(str, "abc")
	if err != nil {
		t.Fatal(err)
	}
	if tel.Client != "abc" {
		t.Fatal("client != abc")
	}

	if tel.Type != Type("tme") {
		t.Fatalf("type != tme")
	}

	should := "Jul 31 202119:24:44"
	if tel.Data != should {
		t.Fatalf("is: %s should:%s", tel.Data, should)

	}
}
