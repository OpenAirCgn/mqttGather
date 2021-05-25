package mqttGather

import "time"

type DB interface {
	Save(*DBAStats, time.Time) (int64, error)
	SaveNow(*DBAStats) (int64, error)
	Close()
}
