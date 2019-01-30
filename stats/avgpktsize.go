// avgpktsize.go
// @name AVG_PKT_SIZE
// @brief The average size of IP packets
package stats

import "math"

func init() {
	Register("AVG_PKT_SIZE",
		func(bs BaseStat) StatInterface { return &AvgPktSize{bs} })
}

type AvgPktSize struct {
	BaseStat
}

func (stat *AvgPktSize) Name() string {
	return "AVG_PKT_SIZE"
}

func (stat *AvgPktSize) Requirement() []string {
	return []string{"IP_BYTES", "IP"}
}

func (stat *AvgPktSize) Compute(ctrvalues []uint64) float64 {
	//ctrvalues[0] -> ip_bytes
	//ctrvalues[1] -> ip
	if ctrvalues[0] == 0 {
		return 0.
	} else if ctrvalues[1] == 0 {
		return math.NaN()
	} else {
		return float64(ctrvalues[0]) / float64(ctrvalues[1])
	}
}
