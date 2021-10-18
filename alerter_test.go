package mqttGather

import (
	"testing"
)

// SMS Mock.
type notifyFunc func(msg, signifier string) error

func (n notifyFunc) SendAlert(msg, signifier string) error {
	return n(msg, signifier)
}

func TestAlerter(t *testing.T) {
	db, _ := getTestDBWithDeviceInfo(t)
	defer db.Close()

	var m_is, s_is string
	notifier := notifyFunc(func(msg, signifier string) error {
		m_is = msg
		s_is = signifier
		return nil
	})
	channel := make(chan DBAStats)
	done := make(chan bool)

	alerter := Alerter{
		DB:           db,
		Notifier:     notifier,
		Threshold:    2.0,
		DelayMS:      10000,
		Count:        5,
		StatsChannel: channel,
		Done:         done,
	}

	alerter.Start()

	// insert 5 Stats below threshold
	for i := 0; i != 5; i = i + 1 {
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
	// insert 4 stats above threshold
	for i := 5; i != 9; i = i + 1 {
		s := DBAStats{
			Signifier: TEST_SIGNIFIER,
			Max:       2.5,
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
		Max:       2.5,
	}
	db.SaveNow(&s)
	channel <- s
	close(channel)
	<-done
	if m_is != "Lautstaerkeueberschreitung an Strassenmusik-Messgeraet bla" || s_is != TEST_SIGNIFIER {
		t.Fatalf("no notification! >%v< >%v<", m_is, s_is == TEST_SIGNIFIER)
	}

}
