// arp_arp.go

package counters

import (
	"sync/atomic"

	"github.com/google/gopacket/layers"
)

func init() {
	Register("ARP", func() BaseCtrInterface {
		return &ARP{
			ARPCtr:  NewARPCtr(),
			Counter: 0}
	})
}

// ARP is an ARP counter counting the number of ARP packets
type ARP struct {
	ARPCtr
	Counter uint64
}

// Name returns the name of the counter (method of BaseCtrInterface)
func (arp *ARP) Name() string {
	return "ARP"
}

// Value returns the current value of the counter (method of BaseCtrInterface)
func (arp *ARP) Value() uint64 {
	return atomic.LoadUint64(&arp.Counter)
}

// Reset resets the counter
func (arp *ARP) Reset() {
	atomic.StoreUint64(&arp.Counter, 0)
}

// Process update the counter according to data it receives
func (arp *ARP) Process(*layers.ARP) {
	atomic.AddUint64(&arp.Counter, 1)
}
