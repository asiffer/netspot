// ricmp.go

// R_ICMP: The ratio of ICMP packets
package stats

import "math"

func init() {
	Register("R_ICMP",
		func(bs BaseStat) StatInterface { return &RIcmp{bs} })
}

type RIcmp struct {
	BaseStat
}

func (stat *RIcmp) Name() string {
	return "R_ICMP"
}

func (stat *RIcmp) Requirement() []string {
	return []string{"ICMP", "IP"}
}

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
