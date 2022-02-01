package mqttGather

import (
	"math"
	"testing"
)

func RandomDBAStats() DBAStats {
	return DBAStats{
		"c4:dd:57:66:95:60",
		52.257,
		58.079,
		54.821,
		0.371,
		55.048,
		86,
	}
}

func TestDBAStatsFromString(t *testing.T) {
	str := "52.683,57.619,55.152,0.595,55.272,86"
	stats, err := DBAStatsFromString(str, "abc")
	if err != nil {
		t.Fatal(err)
	}
	if stats.Num != 86 {
		t.Fatal("num != 86")
	}

	if math.Abs(stats.Average-55.152000) > 0.00001 {
		t.Fatalf("average != %f %f", 55.152000-stats.Average, stats.Average)
	}
}
