// pkt_pkt.go

package counters

import (
	"sync/atomic"

	"github.com/google/gopacket"
)

func init() {
	Register("PKTS", func() BaseCtrInterface {
		return &PKTS{
			PktCtr:  NewPktCtr(),
			Counter: 0}
	})
}

// PKTS stores the pestamp of the packets
type PKTS struct {
	PktCtr
	Counter uint64
}

// Name returns the name of the counter (method of BaseCtrInterface)
func (*PKTS) Name() string {
	return "PKTS"
}

// Value returns the current value of the counter (method of BaseCtrInterface)
func (p *PKTS) Value() uint64 {
	return atomic.LoadUint64(&p.Counter)
}

// Reset resets the counter
func (p *PKTS) Reset() {
	atomic.StoreUint64(&p.Counter, 0)
}

// Process update the counter according to data it receives
func (p *PKTS) Process(pkt gopacket.Packet) {
	atomic.AddUint64(&p.Counter, 1)

}
