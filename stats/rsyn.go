// rsyn.go
// @name R_SYN
// @brief The ratio of packets with TCP + SYN flag
package stats

import "math"

func init() {
	Register("R_SYN",
		func(bs BaseStat) StatInterface { return &RSyn{bs} })
}

type RSyn struct {
	BaseStat
}

func (stat *RSyn) Name() string {
	return "R_SYN"
}

func (stat *RSyn) Requirement() []string {
	return []string{"SYN", "IP"}
}

func (stat *RSyn) Compute(ctrvalues []uint64) float64 {
	//ctrvalues[0] -> syn
	//ctrvalues[1] -> ip
	if ctrvalues[0] == 0 {
		return 0.
	} else if ctrvalues[1] == 0 {
		return math.NaN()
	} else {
		return float64(ctrvalues[0]) / float64(ctrvalues[1])
	}
}
