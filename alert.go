package mqttGather

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// This file contains mechanisms to send notifications in case of violations.
// At the moment this means sending SMS messsages, but in future this may be extend
// to further channels (Ordnungsamt Leitstellen Software, etc)

//	-- log of outgoing alerts
//	CREATE TABLE IF NOT EXISTS alert (
//		alert_id    INTEGER PRIMARY KEY AUTOINCREMENT,
//		device_id   INTEGER REFERENCES device(device_id),
//		ts          INTEGER DEFAULT (STRFTIME('%s','now')),
//		alert_phone  VARCHAR,
//		message     VARCHAR,
//		status      VARCHAR
//	);

// Database mapping of sent alerts
type Alert struct {
	DeviceSignifier string
	Timestamp       int64
	AlertPhone      string
	Message         string
	Status          string
}

// Generic Notifier, currently implemented for SMS and used to mock
// notification for testing (see: alerter_test.go)
type Notifier interface {
	SendAlert(msg, signifier, phone string) (*Alert, error)
}

// Configuration information for SMS notifier:
// `Phone` : target MSISDN
// `Key`   : API key
// `DB`    : database to log alerts to
type SMS struct {
	Key string
}

func normalizePhone(phoneNr string) (string, error) {
	if strings.HasPrefix(phoneNr, "01") {
		phoneNr = "0049" + phoneNr[1:]
	} else if strings.HasPrefix(phoneNr, "+49") {
		phoneNr = "00" + phoneNr[1:]
	} else if !strings.HasPrefix(phoneNr, "0049") {
		return "", fmt.Errorf("unknown phone nr format: %v", phoneNr)
	}
	return phoneNr, nil

}

// Send a notification and stores in `Alert` table
// It's the callers responsibility to keep track of sent alerts, these need
// to be persisted using DB.SaveAlert
func (s *SMS) SendAlert(msg, signifier, phone string) (*Alert, error) {
	phone, err := normalizePhone(phone)
	if err != nil {
		return nil, err
	}
	msgEncoded := url.QueryEscape(msg)
	tmpl := "https://www.smsflatrate.net/schnittstelle.php?key=%s&from=opennoise&to=%s&text=%s&type=10"
	target := fmt.Sprintf(tmpl, s.Key, phone, msgEncoded)

	resp, err := http.Get(target)

	var status string
	if err != nil {
		status = err.Error()
		log.Printf("could not send alert: %v", err)
	} else {
		status = resp.Status
	}
	return &Alert{
		signifier,
		time.Now().Unix(),
		phone,
		msg,
		status,
	}, err
}
