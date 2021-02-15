// ricmp.go
// R_ICMP: The ratio of ICMP packets

package stats

import "math"

func init() {
	Register(&RIcmp{BaseStat{
		name:        "R_ICMP",
		description: "Ratio of ICMP packets (ICMP/IP)"}})
}

// RIcmp computes the ratio of ICMP packets
type RIcmp struct {
	BaseStat
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
