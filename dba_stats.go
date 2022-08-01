package opennoise_daemon

import (
	"fmt"
	"strconv"
	"strings"
)

type DBAStats struct {
	DeviceSignifier string
	Min             float64
	Max             float64
	Average         float64
	AverageVar      float64
	Mean            float64
	Num             int
}

func DBAStatsFromString(csv string, signifier string) (*DBAStats, error) {
	vals := strings.Split(csv, ",")
	if len(vals) != 6 {
		return nil, fmt.Errorf("invalid DBAStats string: %s", csv)
	}

	stats := DBAStats{}
	if f, err := strconv.ParseFloat(vals[0], 32); err != nil {
		return nil, err
	} else {
		stats.Min = float64(f)
	}
	if f, err := strconv.ParseFloat(vals[1], 32); err != nil {
		return nil, err
	} else {
		stats.Max = float64(f)
	}
	if f, err := strconv.ParseFloat(vals[2], 32); err != nil {
		return nil, err
	} else {
		stats.Average = float64(f)
	}
	if f, err := strconv.ParseFloat(vals[3], 32); err != nil {
		return nil, err
	} else {
		stats.AverageVar = float64(f)
	}
	if f, err := strconv.ParseFloat(vals[4], 32); err != nil {
		return nil, err
	} else {
		stats.Mean = float64(f)
	}
	if n, err := strconv.Atoi(vals[5]); err != nil {
		return nil, err
	} else {
		stats.Num = n
	}

	stats.DeviceSignifier = signifier
	return &stats, nil
}

func (s *DBAStats) String() string {
	return fmt.Sprintf(`DBAStats:
Min:  %f
Max:  %f
Avg:  %f
Var:  %f
Mean: %f
Num:  %d`, s.Min, s.Max, s.Average, s.AverageVar, s.Mean, s.Num)
}
