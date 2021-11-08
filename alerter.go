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

func NewAlerter(cfg *RunConfig, mqtt *Mqtt, done chan<- bool) *Alerter {
	return &Alerter{
		mqtt.db.(*SqliteDB),
		&SMS{cfg.SMSKey},
		mqtt.statsChannel,
		done,
	}
}

func (a *Alerter) Start() {

	go func() {
		i := 0
		log.Printf("started alerter.")
		for stats := range a.StatsChannel {
			cfg, err := a.DB.LoadDeviceInfo(stats.Signifier)

			if err != nil {
				if i%30 == 0 {
					log.Printf("E: could not load configuration for device: %s (%v)", stats.Signifier, err)
				}
				i += 1
				continue
			}

			if stats.Max < cfg.AlertThreshold {
				continue
			}
			// TODO check alerts activated ...
			lastAlert, err := a.DB.LoadLastAlert(stats.Signifier)

			if lastAlert.Timestamp+cfg.AlertDeadtime > time.Now().Unix() {
				continue
			}

			cnt, err := a.DB.GetCountThresholdExceeded(stats.Signifier, cfg.AlertDuration, cfg.AlertThreshold)
			if err != nil {
				log.Printf("E: could not retrieve count for: %s (%#v): %v", stats.Signifier, a, err)
				continue
			}

			log.Printf("D: %d threshold violations in %d for %s", cnt, cfg.AlertDuration, stats.Signifier)

			if cnt >= cfg.AlertCount {
				log.Printf("D: %d threshold violations, sending SMS to %s", cnt, cfg.AlertPhone)
				msg := fmt.Sprintf("Lautstaerkeueberschreitung an Strassenmusik-Messgeraet %s", cfg.Description)

				// no need to handle error, sendAlert either takes down system
				// or logs failure. There's nothing more we can do at the moment.
				alert, err := a.Notifier.SendAlert(msg, stats.Signifier, cfg.AlertPhone)

				if _, err = a.DB.SaveAlert(alert); err != nil {
					log.Printf("E: could no save alert: %#v (%v)", alert, err)
				}
			}
		}
		a.Done <- true
	}()
}
