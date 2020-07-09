// icmp_icmp.go

package counters

import (
	"sync/atomic"

	"github.com/google/gopacket/layers"
)

func init() {
	Register(&ICMP{Counter: 0})
}

// ICMP stores the number of ICMP packets
type ICMP struct {
	BaseCtr
	Counter uint64
}

// Name returns the name of the counter (method of BaseCtrInterface)
func (*ICMP) Name() string {
	return "ICMP"
}

// Value returns the current value of the counter (method of BaseCtrInterface)
func (icmp *ICMP) Value() uint64 {
	return atomic.LoadUint64(&icmp.Counter)
}

// Reset resets the counter
func (icmp *ICMP) Reset() {
	atomic.StoreUint64(&icmp.Counter, 0)
}

// Process update the counter according to data it receives
func (icmp *ICMP) Process(*layers.ICMPv4) {
	atomic.AddUint64(&icmp.Counter, 1)
}
