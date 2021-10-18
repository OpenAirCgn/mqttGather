package mqttGather

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

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

type Notifier interface {
	SendAlert(msg string) error
}

type SMS struct {
	Phone string
	Key   string
	DB    DB
}

func NewSMS(phoneNr string, key string, db DB) (*SMS, error) {
	if strings.HasPrefix(phoneNr, "01") {
		phoneNr = "0049" + phoneNr[1:]
	} else if strings.HasPrefix(phoneNr, "+49") {
		phoneNr = "00" + phoneNr[1:]
	} else if !strings.HasPrefix(phoneNr, "0049") {
		return nil, fmt.Errorf("unknown phone nr format: %v", phoneNr)
	}
	return &SMS{
		phoneNr, key, db,
	}, nil
}
func (s *SMS) SendAlert(msg, signifier string) error {
	msgEncoded := url.QueryEscape(msg)
	tmpl := "https://www.smsflatrate.net/schnittstelle.php?key=%s&from=opennoise&to=%s&text=%s&type=10"
	target := fmt.Sprintf(tmpl, s.Key, s.Phone, msgEncoded)

	resp, err := http.Get(target)

	var status string
	if err != nil {
		status = err.Error()
		log.Printf("could not send alert: %v", err)
	} else {
		status = resp.Status
	}
	alert := Alert{
		signifier,
		time.Now().Unix(),
		s.Phone,
		msg,
		status,
	}

	_, err2 := s.DB.SaveAlert(&alert)
	if err2 != nil {
		log.Fatalf("Could not save alert: %v", err2)
	}

	return err

}
