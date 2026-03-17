package analyzer

import "github.com/asiffer/netspot/collector"

// RACK is the ratio of ACK packets: #ACK / #PKT
type RACK struct {
	initialized bool // To handle the first update where we don't have previous values

	prevACK uint64
	prevPKT uint64
}

func (s *RACK) Name() string {
	return "RACK"
}

func (s *RACK) Update(data *collector.Data) float64 {
	if !s.initialized {
		s.initialized = true
		s.prevPKT = data.PKT
		s.prevACK = data.ACK
		return NaN
	}
	value := float64(data.ACK-s.prevACK) / float64(data.PKT-s.prevPKT)
	s.prevPKT = data.PKT
	s.prevACK = data.ACK
	return value
}
