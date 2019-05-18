// udp_udp.go

package counters

import (
	"sync/atomic"

	"github.com/google/gopacket/layers"
)

func init() {
	Register("UDP", func() BaseCtrInterface {
		return &UDP{
			UDPCtr:  NewUDPCtr(),
			Counter: 0}
	})
}

// UDP is an UDP counter counting the number of UDP packets
type UDP struct {
	UDPCtr
	Counter uint64
}

// Name returns the name of the counter (method of BaseCtrInterface)
func (udp *UDP) Name() string {
	return "UDP"
}

// Value returns the current value of the counter (method of BaseCtrInterface)
func (udp *UDP) Value() uint64 {
	return atomic.LoadUint64(&udp.Counter)
}

// Reset resets the counter
func (udp *UDP) Reset() {
	atomic.StoreUint64(&udp.Counter, 0)
}

// Process update the counter according to data it receives
func (udp *UDP) Process(*layers.UDP) {
	atomic.AddUint64(&udp.Counter, 1)
}
