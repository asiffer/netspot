// ip_ip.go
package counters

import (
	"sync/atomic"

	"github.com/google/gopacket/layers"
)

func init() {
	Register("IP", func() BaseCtrInterface {
		return &IP{
			IpCtr:   NewIpCtr(),
			Counter: 0}
	})
}

// IP
type IP struct {
	IpCtr
	Counter uint64
}

// Generic function (BaseCtrInterface)
func (ip *IP) Name() string {
	return "IP"
}

func (ip *IP) Value() uint64 {
	return atomic.LoadUint64(&ip.Counter)
}

func (ip *IP) Reset() {
	atomic.StoreUint64(&ip.Counter, 0)
}

// Specific function (IpCtr)
func (ip *IP) Process(*layers.IPv4) {
	atomic.AddUint64(&ip.Counter, 1)
}
