// rdstsrc.go
// @name R_DST_SRC
// @brief The ratio 'number of unique destination addresses' /
// 'number of unique source addresses'
package stats

import "math"

func init() {
	Register("R_DST_SRC",
		func(bs BaseStat) StatInterface { return &RDstSrc{bs} })
}

type RDstSrc struct {
	BaseStat
}

func (stat *RDstSrc) Name() string {
	return "R_DST_SRC"
}

func (stat *RDstSrc) Requirement() []string {
	return []string{"NB_UNIQ_DST_ADDR", "NB_UNIQ_SRC_ADDR"}
}

func (stat *RDstSrc) Compute(ctrvalues []uint64) float64 {
	//ctrvalues[0] -> NB_UNIQ_DST_ADDR
	//ctrvalues[1] -> NB_UNIQ_SRC_ADDR
	if ctrvalues[0] == 0 {
		return 0.
	} else if ctrvalues[1] == 0 {
		return math.NaN()
	} else {
		return float64(ctrvalues[0]) / float64(ctrvalues[1])
	}
}
