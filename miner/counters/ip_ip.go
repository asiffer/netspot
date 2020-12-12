// ip_ip.go

package counters

import (
	"sync/atomic"

	"github.com/google/gopacket/layers"
)

func init() {
	Register(&IP{Counter: 0})
}

// IP is an IP counter counting the number of IP packets
type IP struct {
	BaseCtr
	Counter uint64
}

// Name returns the name of the counter (method of BaseCtrInterface)
func (ip *IP) Name() string {
	return "IP"
}

// Value returns the current value of the counter (method of BaseCtrInterface)
func (ip *IP) Value() uint64 {
	return atomic.LoadUint64(&ip.Counter)
}

// Reset resets the counter
func (ip *IP) Reset() {
	atomic.StoreUint64(&ip.Counter, 0)
}

// Process update the counter according to data it receives
func (ip *IP) Process(*layers.IPv4) {
	atomic.AddUint64(&ip.Counter, 1)
}
