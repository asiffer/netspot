package analyzer

import (
	"github.com/asiffer/netspot/collector"
)

// BPS is the rate of bytes/sec traffic
type BPS struct {
	initialized bool // To handle the first update where we don't have previous values

	prevBytes uint64
	prevTime  uint64
}

func (s *BPS) Name() string {
	return "BPS"
}

func (s *BPS) Update(data *collector.Data) float64 {
	if !s.initialized {
		s.initialized = true
		s.prevTime = data.TIME
		s.prevBytes = data.BYTES
		return NaN
	}
	duration := float64(data.TIME-s.prevTime) / 1e9 // in seconds
	if duration <= 0 {
		return NaN
	}
	value := float64(data.BYTES-s.prevBytes) / duration
	s.prevTime = data.TIME
	s.prevBytes = data.BYTES
	return value
}
