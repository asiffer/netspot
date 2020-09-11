// avgpktsize.go
// AVG_PKT_SIZE: The average size of IP packets

package stats

import "math"

func init() {
	Register(&AvgPktSize{BaseStat{name: "AVG_PKT_SIZE"}})
}

// AvgPktSize computes the average size of an IP packet
type AvgPktSize struct {
	BaseStat
}

// Name returns the unique name of the stat
func (stat *AvgPktSize) Name() string {
	return "AVG_PKT_SIZE"
}

// Requirement returns teh requested counters to compute the stat
func (stat *AvgPktSize) Requirement() []string {
	return []string{"IP_BYTES", "IP"}
}

// Compute implements the way to compute the stat from the counters
func (stat *AvgPktSize) Compute(ctrvalues []uint64) float64 {
	//ctrvalues[0] -> ip_bytes
	//ctrvalues[1] -> ip
	if ctrvalues[0] == 0 {
		return 0.
	} else if ctrvalues[1] == 0 {
		return math.NaN()
	} else {
		return float64(ctrvalues[0]) / float64(ctrvalues[1])
	}
}
