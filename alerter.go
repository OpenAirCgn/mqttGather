package mqttGather

import (
	"fmt"
	"log"
	//	"time"
)

type Alerter struct {
	DB           *SqliteDB
	Notifier     Notifier
	Threshold    float64
	DelayMS      int64
	Count        int64
	StatsChannel <-chan DBAStats
	Done         chan<- bool
}

func (a *Alerter) Start() {
	go func() {
		for stats := range a.StatsChannel {
			if stats.Max < a.Threshold {
				continue
			}

			cnt, err := a.DB.GetCountThresholdExceeded(stats.Signifier, a.DelayMS/1000, a.Threshold)
			if err != nil {
				log.Printf("could not retrieve count for: %s (%#v): %v", stats.Signifier, a, err)
			}

			if cnt >= a.Count {
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
