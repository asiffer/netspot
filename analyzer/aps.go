package analyzer

import (
	"github.com/asiffer/netspot/collector"
)

// APS is the Average Packet Size: #BYTES / #PKT
type APS struct {
	initialized bool // To handle the first update where we don't have previous values

	prevPKT   uint64
	prevBYTES uint64
}

func (s *APS) Name() string {
	return "APS"
}

func (s *APS) Update(data *collector.Data) float64 {
	if !s.initialized {
		s.initialized = true
		s.prevPKT = data.PKT
		s.prevBYTES = data.BYTES
		return NaN
	}
	value := float64(data.BYTES-s.prevBYTES) / float64(data.PKT-s.prevPKT)
	s.prevPKT = data.PKT
	s.prevBYTES = data.BYTES
	return value
}
