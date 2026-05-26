package analyzer

import "github.com/asiffer/netspot/collector"

// PPS is the Packets Per Second: #PKT / TIME
type PPS struct {
	initialized bool // To handle the first update where we don't have previous values

	prevPKT  uint64
	prevTIME uint64
}

func (s *PPS) Name() string {
	return "PPS"
}

func (s *PPS) Update(data *collector.Data) float64 {
	if !s.initialized {
		s.initialized = true
		s.prevTIME = data.TIME
		s.prevPKT = data.PKT
		return NaN
	}
	value := float64(data.PKT-s.prevPKT) / (float64(data.TIME-s.prevTIME) * 1e-9)
	s.prevPKT = data.PKT
	s.prevTIME = data.TIME
	return value
}
