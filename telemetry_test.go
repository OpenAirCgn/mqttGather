package mqttGather

import "testing"

func TestTelemetryFromPayloadTime(t *testing.T) {
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

func TestTelemetryFromPayloadFlag(t *testing.T) {
	str := "flg:0000000F"
	tel, err := TelemetryFromPayload(str, "abc")
	if err != nil {
		t.Fatal(err)
	}
	if tel.Client != "abc" {
		t.Fatal("client != abc")
	}

	if tel.Type != Type("flg") {
		t.Fatalf("type != flg")
	}

	should := 15
	if tel.Data != should {
		t.Fatalf("is: %x should:%x", tel.Data, should)

	}

}
