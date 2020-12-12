// rack.go
// R_ACK: The ratio of packets with TCP + ACK flag

package stats

func init() {
	Register(&RACK{BaseStat: BaseStat{name: "R_ACK"}})
}

// RACK computes the ratio of packets with TCP + ACK flag
type RACK struct {
	BaseStat
}

// Name returns the unique name of the stat
func (stat *RACK) Name() string {
	return "R_ACK"
}

// Requirement returns teh requested counters to compute the stat
func (stat *RACK) Requirement() []string {
	return []string{"ACK", "IP"}
}

// Compute implements the way to compute the stat from the counters
func (stat *RACK) Compute(ctrvalues []uint64) float64 {
	//ctrvalues[0] -> ack
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
