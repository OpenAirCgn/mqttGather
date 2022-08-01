package opennoise_daemon

import(
	"strings"
	"fmt"
	"errors"
	"net/http"
	"net/url"
)

// Generic Notifier, currently implemented for SMS and used to mock
// notification for testing (see: alerter_test.go)
type Notifier interface {
  SendAlert(msg string, phone string) error
}

// This is a notifier that sends via SMS
type SMSNotifier struct {
  Key string
}

func NewSMSNotifier (key string) (*SMSNotifier, error) {
	sms := SMSNotifier {
		Key: key,
	}
	return &sms, nil
}

func normalizePhone(phoneNr string) (string, error) {
  switch {
  case strings.HasPrefix(phoneNr, "01"):
    phoneNr = "0049" + phoneNr[1:]
  case strings.HasPrefix(phoneNr, "+49"):
    phoneNr = "00" + phoneNr[1:]
  case !strings.HasPrefix(phoneNr, "0049"):
    return "", fmt.Errorf("unknown phone nr format: %v", phoneNr)
  }
  return phoneNr, nil
}

// Send a notification and stores in `Alert` table
// It's the callers responsibility to keep track of sent alerts, these need
// to be persisted using DB.SaveAlert
func (s *SMSNotifier) SendAlert(msg string, phone string) error {
  if s.Key == "" {
    return errors.New("not sent, no sms key")
  }
	normPhone, err := normalizePhone(phone)
	if err != nil {
		return err
	}
	msgEncoded := url.QueryEscape(msg)
	tmpl := "https://www.smsflatrate.net/schnittstelle.php?key=%s&from=opennoise&to=%s&text=%s&type=10"
	target := fmt.Sprintf(tmpl, s.Key, normPhone, msgEncoded)
	_ , err = http.Get(target)
	//TODO: check returned response? 
  return err
}
