// rip.go
// R_IP: The ratio of IP packets

package stats

import "math"

func init() {
	Register("R_IP",
		func(bs BaseStat) StatInterface { return &RIP{bs} })
}

// RIP computes the ratio of ICMP packets
type RIP struct {
	BaseStat
}

// Name returns the unique name of the stat
func (stat *RIP) Name() string {
	return "R_IP"
}

// Requirement returns the requested counters to compute the stat
func (stat *RIP) Requirement() []string {
	return []string{"IP", "PKTS"}
}

// Compute implements the way to compute the stat from the counters
func (stat *RIP) Compute(ctrvalues []uint64) float64 {
	//ctrvalues[0] -> ip
	//ctrvalues[1] -> pkts
	if ctrvalues[0] == 0 {
		return 0.
	} else if ctrvalues[1] == 0 {
		return math.NaN()
	} else {
		return float64(ctrvalues[0]) / float64(ctrvalues[1])
	}
}
