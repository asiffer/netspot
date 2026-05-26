package analyzer

import (
	"math"

	"github.com/asiffer/netspot/collector"
)

// DPE is the Destination Port Entropy: -SUM(p_i * log(p_i)) where p_i is the probability of a packet being sent to port i
// p_i = N_i / N where N_i is the number of packets sent to port i and N is the total number of packets
type DPE struct {
	initialized bool // To handle the first update where we don't have previous values

	prevPORTS [65536]uint64
}

func (s *DPE) Name() string {
	return "DPE"
}

func (s *DPE) Update(data *collector.Data) float64 {

	if !s.initialized {
		s.initialized = true
		s.prevPORTS = data.TCP_DST_PORTS
		return NaN
	}
	// compute entropy
	H := 0.0
	total := uint64(0)
	for i := range 65536 {
		Ni := data.TCP_DST_PORTS[i] - s.prevPORTS[i]
		if Ni > 0 {
			total += Ni
			H -= float64(Ni) * math.Log(float64(Ni))
		}
	}
	if total == 0 {
		return NaN
	}
	N := float64(total)
	value := math.Log(N) - H/N
	s.prevPORTS = data.TCP_DST_PORTS
	return value
}
