// traffic.go
// TRAFFIC: The ratio #IP packets / window size

package stats

import "math"

func init() {
	Register(&Traffic{
		BaseStat: BaseStat{name: "TRAFFIC"},
		init:     false,
		lastTime: 0},
	)
}

// Traffic computes the ratio number of IP packets / time window size
type Traffic struct {
	BaseStat
	init     bool
	lastTime uint64
}

// Name returns the unique name of the stat
func (stat *Traffic) Name() string {
	return "TRAFFIC"
}

// Requirement returns teh requested counters to compute the stat
func (stat *Traffic) Requirement() []string {
	return []string{"IP", "SOURCE_TIME"}
}

// Compute implements the way to compute the stat from the counters
func (stat *Traffic) Compute(ctrvalues []uint64) float64 {
	//ctrvalues[0] -> ip
	//ctrvalues[1] -> source time

	// if the stat has not stored a starting point yet
	if !stat.init {
		// stat.lastPackets = ctrvalues[0]
		stat.lastTime = ctrvalues[1]
		stat.init = true
		return math.NaN()
	}

	output := 0.
	// packets is flushed
	nbPackets := ctrvalues[0]
	deltaTime := ctrvalues[1] - stat.lastTime

	if deltaTime == 0 || nbPackets == 0 {
		output = 0.
	} else {
		// try to compute pkt/ms
		output = 1e6 * float64(nbPackets) / float64(deltaTime)
		// output = float64(nbPackets) / (1e-9 * float64(deltaTime))
	}

	stat.lastTime = ctrvalues[1]
	return output
}
