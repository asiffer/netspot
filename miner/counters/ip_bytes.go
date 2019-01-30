// ip_bytes.go
package counters

import (
	"sync/atomic"

	"github.com/google/gopacket/layers"
)

func init() {
	Register("IP_BYTES", func() BaseCtrInterface {
		return &IP_BYTES{
			IpCtr:   NewIpCtr(),
			Counter: 0}
	})
}

// IP_BYTES
// Store the size of IP payloads (in bytes)
type IP_BYTES struct {
	IpCtr
	Counter uint64
}

// Generic function (BaseCtrInterface)
func (ip_bytes *IP_BYTES) Name() string {
	return "IP_BYTES"
}

func (ip_bytes *IP_BYTES) Value() uint64 {
	return atomic.LoadUint64(&ip_bytes.Counter)
}

func (ip_bytes *IP_BYTES) Reset() {
	atomic.StoreUint64(&ip_bytes.Counter, 0)
}

// Specific function (IpCtr)
func (ip_bytes *IP_BYTES) Process(ip *layers.IPv4) {
	atomic.AddUint64(&ip_bytes.Counter, uint64(ip.Length))
}

// END OF IP_BYTES
