// rdstsrcport.go
// R_DST_SRC_PORT: The ratio 'number of unique destination ports' /
// 'number of unique source ports'

package stats

import "math"

func init() {
	Register("R_DST_SRC_PORT",
		func(bs BaseStat) StatInterface { return &RDstSrcPort{bs} })
}

// RDstSrcPort computes the ratio 'number of unique destination ports'
// / 'number of unique source portd'
//
type RDstSrcPort struct {
	BaseStat
}

// Name returns the unique name of the stat
func (stat *RDstSrcPort) Name() string {
	return "R_DST_SRC_PORT"
}

// Requirement returns teh requested counters to compute the stat
func (stat *RDstSrcPort) Requirement() []string {
	return []string{"NB_UNIQ_DST_PORT", "NB_UNIQ_SRC_PORT"}
}

// Compute implements the way to compute the stat from the counters
func (stat *RDstSrcPort) Compute(ctrvalues []uint64) float64 {
	//ctrvalues[0] -> NB_UNIQ_DST_PORT
	//ctrvalues[1] -> NB_UNIQ_SRC_PORT
	if ctrvalues[0] == 0 {
		return 0.
	} else if ctrvalues[1] == 0 {
		return math.NaN()
	} else {
		return float64(ctrvalues[0]) / float64(ctrvalues[1])
	}
}
