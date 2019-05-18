// ricmp.go
// R_ICMP: The ratio of ICMP packets

package stats

import "math"

func init() {
	Register("R_ICMP",
		func(bs BaseStat) StatInterface { return &RIcmp{bs} })
}

// RIcmp computes the ratio of ICMP packets
type RIcmp struct {
	BaseStat
}

// Name returns the unique name of the stat
func (stat *RIcmp) Name() string {
	return "R_ICMP"
}

// Requirement returns the requested counters to compute the stat
func (stat *RIcmp) Requirement() []string {
	return []string{"ICMP", "IP"}
}

// Compute implements the way to compute the stat from the counters
func (stat *RIcmp) Compute(ctrvalues []uint64) float64 {
	//ctrvalues[0] -> icmp
	//ctrvalues[1] -> ip
	if ctrvalues[0] == 0 {
		return 0.
	} else if ctrvalues[1] == 0 {
		return math.NaN()
	} else {
		return float64(ctrvalues[0]) / float64(ctrvalues[1])
	}
}
