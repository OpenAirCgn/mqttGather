package mqttGather

import "testing"
import "time"

func testMQTT(t *testing.T) {
	if m, err := NewMQTT(nil); err != nil {
		t.Fatal(err)
	} else {
		time.Sleep(30 * time.Second)
		m.Disconnect()
	}

}

func TestClientDissection(t *testing.T) {
	str := "/opennoise/c4:dd:57:66:95:60/dba_stats"
	client := str[11 : 11+17]
	if client != "c4:dd:57:66:95:60" {
		t.Fatal(client)
	}
}
