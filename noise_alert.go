package opennoise_daemon

import(
	"time"
	"log"
	"fmt"
)

type NoiseAlert struct {
  DB DB;
  Notifier Notifier;
}

func NewNoiseAlert(db DB, notifier Notifier) (*NoiseAlert, error) {
  na := NoiseAlert {
    DB: db,
    Notifier: notifier,
  }
  return &na, nil
}

func (n *NoiseAlert) SensorTelemetry(t *Telemetry) {
	//nothing to do here
}

func (n *NoiseAlert) SensorDBAStats(dba *DBAStats) {

	now := time.Now().Unix()

	devInfo, err := n.DB.LoadDeviceInfo(dba.DeviceSignifier)
	if err != nil {
		log.Printf("E: could not load configuration for device: %s (%v)", dba.DeviceSignifier, err)
		return
	}

	if devInfo.AlertActive == false {	//device is temporarily or permanently off
		if devInfo.TurnOnTime <= 0 {	//permanently off
			return
		}
		if devInfo.TurnOnTime > now {	//temporarily off, will turn on in the future
			return
		}
		//TODO: we might set AlertActive to true here
	}

	if dba.Max < devInfo.AlertThreshold {
		return	//threshold not exceeded, impossible to trigger alert
	}

	lastAlert, err := n.DB.LoadLastAlert(dba.DeviceSignifier)
	if err != nil {
		log.Printf("E: could not load configuration for device: %s (%v)", dba.DeviceSignifier, err)
		return		
	}

	if lastAlert.Timestamp+devInfo.AlertDeadtime > time.Now().Unix() {
		return //still in deadtime after last alert
	}

	cnt, err := n.DB.GetCountThresholdExceeded(dba.DeviceSignifier, devInfo.AlertDuration, devInfo.AlertThreshold)
	if err != nil {
		log.Printf("E: could not get threshold exceed count: %s (%v)", dba.DeviceSignifier, err)
		return
	}

	if cnt < devInfo.AlertCount {
		return	//not enough threshold exceedings
	}
	
	log.Printf("D: %d threshold violations, sending SMS to %s", cnt, devInfo.AlertPhone)
	a := Alert {
		DeviceSignifier:dba.DeviceSignifier,
		Timestamp:now,
		AlertPhone:devInfo.AlertPhone,
		Message:fmt.Sprintf("Lautstaerkeueberschreitung an Strassenmusik-Messgeraet %s", devInfo.Description),
		Status:"not sent yet",
	}
	go n.SendMessage(a)
}


//goroutine to send SMS messages without blocking. Alerts will be logged after completion.
func (n *NoiseAlert) SendMessage(alert Alert) {
	err := n.Notifier.SendAlert(alert.Message, alert.AlertPhone)
	if err != nil {
		log.Printf("E: Send alert failed: %v", err)
		alert.Status = err.Error()
	} else {
		alert.Status = "OK"
	}
	_ , sendErr := n.DB.SaveAlert(&alert) 
	if sendErr != nil {
		log.Printf("E: could not write to alert log: %v", sendErr)
	}
}
