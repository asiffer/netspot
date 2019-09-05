// rarp.go
// R_ARP: The ratio of ARP packets

package stats

import "math"

func init() {
	Register("R_ARP",
		func(bs BaseStat) StatInterface { return &RARP{bs} })
}

// RARP computes the ratio of ICMP packets
type RARP struct {
	BaseStat
}

// Name returns the unique name of the stat
func (stat *RARP) Name() string {
	return "R_ARP"
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
