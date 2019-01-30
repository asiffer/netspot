// rack.go
// @name R_ACK
// @brief The ratio of packets with TCP + ACK flag
package stats

import "math"

func init() {
	Register("R_ACK",
		func(bs BaseStat) StatInterface { return &RAck{bs} })
}

type RAck struct {
	BaseStat
}

func (stat *RAck) Name() string {
	return "R_ACK"
}

func (stat *RAck) Requirement() []string {
	return []string{"ACK", "IP"}
}

func (stat *RAck) Compute(ctrvalues []uint64) float64 {
	//ctrvalues[0] -> ack
	//ctrvalues[1] -> ip
	if ctrvalues[0] == 0 {
		return 0.
	} else if ctrvalues[1] == 0 {
		return math.NaN()
	} else {
		return float64(ctrvalues[0]) / float64(ctrvalues[1])
	}
}
