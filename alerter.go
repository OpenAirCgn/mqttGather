package mqttGather

import (
	"fmt"
	"log"
	"time"
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
	DB           *SqliteDB
	Notifier     Notifier
	StatsChannel <-chan DBAStats
	Done         chan<- bool
}

func (a *Alerter) Start() {

	go func() {
		for stats := range a.StatsChannel {

			cfg, err := a.DB.LoadDeviceInfo(stats.Signifier)

			if err != nil {
				// TODO die here?
				log.Printf("E: could not load configuration for device: %s (%v)", stats.Signifier, err)
				continue
			}

			// TODO check no alert has been send in DeadTime
			if stats.Max < cfg.AlertThreshold {
				continue
			}

			lastAlert, err := a.DB.LoadLastAlert(stats.Signifier)

			if lastAlert.Timestamp+cfg.AlertDeadtime > time.Now().Unix() {
				continue
			}

			cnt, err := a.DB.GetCountThresholdExceeded(stats.Signifier, cfg.AlertDuration, cfg.AlertThreshold)
			if err != nil {
				log.Printf("E: could not retrieve count for: %s (%#v): %v", stats.Signifier, a, err)
				continue
			}

			if cnt >= cfg.AlertCount {
				device, err := a.DB.LoadDeviceInfo(stats.Signifier)
				if err != nil {
					log.Printf("Could not load Device Info for (%s): %v", stats.Signifier, err)
					continue
				}
				msg := fmt.Sprintf("Lautstaerkeueberschreitung an Strassenmusik-Messgeraet %s", device.Description)

				// no need to handle error, sendAlert either takes down system
				// or logs failure. There's nothing more we can do at the moment.
				alert, err := a.Notifier.SendAlert(msg, stats.Signifier)

				if _, err = a.DB.SaveAlert(alert); err != nil {
					log.Printf("E: could no save alert: %#v (%v)", alert, err)
				}
			}
		}
		a.Done <- true
	}()
}
