// rip.go
// R_IP: The ratio of IP packets

package stats

import "math"

func init() {
	Register(&RIP{BaseStat{name: "R_IP", description: "Ratio of IP packets (IP/ALL)"}})
}

// RIP computes the ratio of ICMP packets
type RIP struct {
	BaseStat
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
