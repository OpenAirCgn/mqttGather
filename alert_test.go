package mqttGather

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestSendAlert(t *testing.T) {
	key := os.Getenv("SMSKEY")
	if key == "" {
		t.Skip("skipping test. Set SMSKEY env variable to run.")
	}
	db, _ := getTestDBWithDevice(t)
	defer db.Close()

	sms := SMS{key}
	msg := fmt.Sprintf("Test SMS: %s", time.Now().String())
	alert, err := sms.SendAlert(msg, TEST_SIGNIFIER, "01791001709")
	if err != nil {
		t.Fatalf("Failed to send sms: %v", err)
	}
	// to honor deadtime, alert would need to be saved!
	if alert.AlertPhone != "+491791001709" {
		t.Fatalf("alert info transfered incorrectly: %#v", alert)
	}

}
