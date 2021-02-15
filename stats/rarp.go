// rarp.go
// R_ARP: The ratio of ARP packets

package stats

import "math"

func init() {
	Register(&RARP{BaseStat{
		name:        "R_ARP",
		description: "Ratio of ARP packets (ARP/ALL)"}})
}

// RARP computes the ratio of ICMP packets
type RARP struct {
	BaseStat
}

// Requirement returns the requested counters to compute the stat
func (stat *RARP) Requirement() []string {
	return []string{"ARP", "PKTS"}
}

// Compute implements the way to compute the stat from the counters
func (stat *RARP) Compute(ctrvalues []uint64) float64 {
	//ctrvalues[0] -> arp
	//ctrvalues[1] -> pkts
	if ctrvalues[0] == 0 {
		return 0.
	} else if ctrvalues[1] == 0 {
		return math.NaN()
	} else {
		return float64(ctrvalues[0]) / float64(ctrvalues[1])
	}
}
