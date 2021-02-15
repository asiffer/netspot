// perf.go
// PERF: the packet parsing rate

package stats

import (
	"math"
)

func init() {
	Register(&Perf{
		BaseStat: BaseStat{
			name:        "PERF",
			description: "Packet processing rate (pkts/second)"},
		init:     false,
		lastTime: 0})
}

// Perf computes the ratio of packets with TCP + SYN flag
type Perf struct {
	BaseStat
	init     bool
	lastTime uint64
	// lastPackets uint64
}

// Requirement returns the requested counters to compute the stat
func (stat *Perf) Requirement() []string {
	return []string{"PKTS", "REAL_TIME"}
}

// Update normally feeds the DSpot instance embedded
// in the BaseStat. But this stat is not expected
// to be monitored.
func (stat *Perf) Update(val float64) int {
	// Returns 0 to avoid the monitoring
	return 0
}

// Compute implements the way to compute the stat from the counters
func (stat *Perf) Compute(ctrvalues []uint64) float64 {
	// if the stat has not stored a starting point yet
	if !stat.init {
		// stat.lastPackets = ctrvalues[0]
		stat.lastTime = ctrvalues[1]
		stat.init = true
		return math.NaN()
	}

	output := 0.
	// deltaPackets := ctrvalues[0] - stat.lastPackets
	// packets is flushed
	nbPackets := ctrvalues[0]
	deltaTime := ctrvalues[1] - stat.lastTime

	if deltaTime == 0 || nbPackets == 0 {
		output = 0.
	} else {
		output = float64(nbPackets) / (1e-9 * float64(deltaTime))
	}

	// fmt.Println(ctrvalues, deltaPackets, deltaTime)

	// store new values
	// stat.lastPackets = ctrvalues[0]
	stat.lastTime = ctrvalues[1]
	return output
}
