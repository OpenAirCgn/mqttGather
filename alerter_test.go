package opennoise_daemon

import (
	"bytes"
	"log"
	"os"
	"testing"
	"time"
)

// SMS Mock.
type notifyFunc func(msg, signifier, phone string) error

func (n notifyFunc) SendAlert(msg, signifier, phone string) (*Alert, error) {
	err := n(msg, signifier, phone)
	return &Alert{
		signifier,
		time.Now().Unix(),
		phone,
		msg,
		"ok",
	}, err
}

func TestAlerter(t *testing.T) {
	db, _ := getTestDBWithDeviceInfo(t)
	defer db.Close()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	var m_is, s_is string
	notifier := notifyFunc(func(msg, signifier, phone string) error {
		m_is = msg
		s_is = signifier
		return nil
	})
	channel := make(chan DBAStats)
	done := make(chan bool)

	alerter := Alerter{
		DB:           db,
		Notifier:     notifier,
		StatsChannel: channel,
		Done:         done,
	}

	alerter.Start()

	// insert 5 Stats below threshold
	for i := 0; i != 5; i++ {
		s := DBAStats{
			Signifier: TEST_SIGNIFIER,
			Max:       1.5,
		}
		if _, err := db.SaveNow(&s); err != nil {
			t.Fatalf("%v", err)
		}
		channel <- s
	}
	// send stats, check
	if m_is != "" || s_is != "" {
		t.Fatalf("received unwarranted alert: %s %s", m_is, s_is)
	}
	// insert 2 stats above threshold
	for i := 5; i != 7; i++ {
		s := DBAStats{
			Signifier: TEST_SIGNIFIER,
			Max:       102.0,
		}
		db.SaveNow(&s)
		channel <- s
	}
	// aend stats, check
	if m_is != "" || s_is != "" {
		t.Fatalf("received unwarranted alert (2): %s %s", m_is, s_is)
	}
	// insert on more, check woohoo!
	s := DBAStats{
		Signifier: TEST_SIGNIFIER,
		Max:       102.5,
	}
	db.SaveNow(&s)
	channel <- s

	if m_is != "Lautstaerkeueberschreitung an Strassenmusik-Messgeraet bla" || s_is != TEST_SIGNIFIER {
		t.Fatalf("no notification! >%v< >%v<", m_is, s_is == TEST_SIGNIFIER)
	}

	m_is = "nothing"

	db.SaveNow(&s)
	channel <- s

	if m_is != "nothing" {
		t.Fatalf("received unwarranted alert (3:deadtime) %s", m_is)
	}

	close(channel)
	<-done

}
