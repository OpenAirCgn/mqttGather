package opennoise_daemon

import (
	"log"
)

type NoiseGather struct {
	DB DB;
}

func NewNoiseGather(db DB) (*NoiseGather, error) {
	ng := NoiseGather {
		DB: db,
	}
	return &ng, nil
}

func (n *NoiseGather) SensorDBAStats(dba *DBAStats) {
	_, err := n.DB.SaveNow(dba);
	if err != nil {
		log.Printf("E: could not save %s : %v", dba, err)
		err := n.DB.Reconnect();
		if err != nil {
			log.Printf("E: could not reconnect to db (%v)", err)
		}
	}
}

func (n *NoiseGather) SensorTelemetry(t *Telemetry) {
	_, err := n.DB.SaveTelemetryNow(t);
	if err != nil {
		log.Printf("E: could not save %s : %v", t, err)
		err := n.DB.Reconnect();
		if err != nil {
			log.Printf("E: could not reconnect to db (%v)", err)
		}
	}
}

