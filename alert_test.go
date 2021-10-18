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

	sms, err := NewSMS("01791001709", key, db)
	if err != nil {
		t.Fatalf("%v", err)
	}
	msg := fmt.Sprintf("Test SMS: %s", time.Now().String())
	if err := sms.SendAlert(msg, TEST_SIGNIFIER); err != nil {
		t.Fatalf("Failed to send sms: %v", err)
	}

	alert, err := db.LoadLastAlert(TEST_SIGNIFIER)

	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}
	if alert.DeviceSignifier != TEST_SIGNIFIER {
		t.Fatalf("0 is: %s, should %s", alert.DeviceSignifier, TEST_SIGNIFIER)
	}

	if alert.Message != msg {
		t.Fatalf("1 is: %s, should %s", alert.Message, msg)
	}

	// t.Fatalf("%#v", alert)
}
