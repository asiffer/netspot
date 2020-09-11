// rsyn.go
// R_SYN: The ratio of packets with TCP + SYN flag

package stats

func init() {
	Register(&RSYN{BaseStat{name: "R_SYN"}})
}

// RSYN computes the ratio of packets with TCP + SYN flag
type RSYN struct {
	BaseStat
}

// Name returns the unique name of the stat
func (stat *RSYN) Name() string {
	return "R_SYN"
}

// Requirement returns teh requested counters to compute the stat
func (stat *RSYN) Requirement() []string {
	return []string{"SYN", "IP"}
}

// Compute implements the way to compute the stat from the counters
func (stat *RSYN) Compute(ctrvalues []uint64) float64 {
	//ctrvalues[0] -> syn
	//ctrvalues[1] -> ip
	// if ctrvalues[0] == 0 {
	// 	return 0.
	// } else if ctrvalues[1] == 0 {
	// 	return math.NaN()
	// } else {
	// 	return float64(ctrvalues[0]) / float64(ctrvalues[1])
	// }
	if ctrvalues[0] == 0 || ctrvalues[1] == 0 {
		return 0.
	}
	return float64(ctrvalues[0]) / float64(ctrvalues[1])
}
