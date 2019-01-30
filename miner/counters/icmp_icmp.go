// icmp_icmp.go
package counters

import (
	"sync/atomic"

	"github.com/google/gopacket/layers"
)

func init() {
	Register("ICMP", func() BaseCtrInterface {
		return &ICMP{
			IcmpCtr: NewIcmpCtr(),
			Counter: 0}
	})
}

// ICMP
// Store the number of ICMP packets
type ICMP struct {
	IcmpCtr
	Counter uint64
}

// Generic function (BaseCtrInterface)
func (*ICMP) Name() string {
	return "ICMP"
}

func (icmp *ICMP) Value() uint64 {
	return atomic.LoadUint64(&icmp.Counter)
}

func (icmp *ICMP) Reset() {
	atomic.StoreUint64(&icmp.Counter, 0)
}

// Specific function (IcmpCtr)
func (icmp *ICMP) Process(*layers.ICMPv4) {
	atomic.AddUint64(&icmp.Counter, 1)
}
