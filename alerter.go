package mqttGather

import (
	"fmt"
	"log"
	//	"time"
)

// This file contains functionality to monitor for noise violations:
// - each message received per MQTT to passed to the `Alerter` via a channel after being persisted tp the DB
// - if
//     -     the message's max field exceeds the threshold
//     - AND the threshold has been exceeds at least Count number of times
//     -     in the past `DelayMS` ms
//     - AND no previous Alert has been send in the past DeadTime
// - an SMS Alert is send and persisted.

type Alerter struct {
	DB              *SqliteDB
	Notifier        Notifier
	Threshold       float64
	ViolationWindow int64 // seconds
	ViolationCount  int64
	DeadTime        int64
	StatsChannel    <-chan DBAStats
	Done            chan<- bool
}

func (a *Alerter) Start() {
	go func() {
		for stats := range a.StatsChannel {
			// TODO check no alert has been send in DeadTime
			if stats.Max < a.Threshold {
				continue
			}

			cnt, err := a.DB.GetCountThresholdExceeded(stats.Signifier, a.ViolationWindow, a.Threshold)
			if err != nil {
				log.Printf("could not retrieve count for: %s (%#v): %v", stats.Signifier, a, err)
			}

			if cnt >= a.ViolationCount {
				device, err := a.DB.LoadDeviceInfo(stats.Signifier)
				if err != nil {
					log.Printf("Could not load Device Info for (%s): %v", stats.Signifier, err)
					continue
				}
				msg := fmt.Sprintf("Lautstaerkeueberschreitung an Strassenmusik-Messgeraet %s", device.Description)

				// no need to handle error, sendAlert either takes down system
				// or logs failure. There's nothing more we can do at the moment.
				a.Notifier.SendAlert(msg, stats.Signifier)
			}
		}
		a.Done <- true
	}()
}
