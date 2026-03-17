package analyzer

import "github.com/asiffer/netspot/collector"

// SYNFIN is the ratio : #SYN / #PFIN
type SYNFIN struct {
	initialized bool // To handle the first update where we don't have previous values

	prevSYN uint64
	prevFIN uint64
}

func (s *SYNFIN) Name() string {
	return "SYNFIN"
}

func (s *SYNFIN) Update(data *collector.Data) float64 {
	if !s.initialized {
		s.initialized = true
		s.prevSYN = data.SYN
		s.prevFIN = data.FIN
		return NaN
	}
	value := float64(data.SYN-s.prevSYN) / float64(data.FIN-s.prevFIN)
	s.prevFIN = data.FIN
	s.prevSYN = data.SYN
	return value
}
