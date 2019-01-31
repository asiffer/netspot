// ip_bytes.go

package counters

import (
	"sync/atomic"

	"github.com/google/gopacket/layers"
)

func init() {
	Register("IP_BYTES", func() BaseCtrInterface {
		return &IPBytes{
			IPCtr:   NewIPCtr(),
			Counter: 0}
	})
}

// IPBytes stores the size of IP payloads (in bytes)
type IPBytes struct {
	IPCtr
	Counter uint64
}

// Name returns the name of the counter (method of BaseCtrInterface)
func (ip_bytes *IPBytes) Name() string {
	return "IP_BYTES"
}

// Value returns the current value of the counter (method of BaseCtrInterface)
func (ip_bytes *IPBytes) Value() uint64 {
	return atomic.LoadUint64(&ip_bytes.Counter)
}

// Reset resets the counter
func (ip_bytes *IPBytes) Reset() {
	atomic.StoreUint64(&ip_bytes.Counter, 0)
}

// Process update the counter according to data it receives
func (ip_bytes *IPBytes) Process(ip *layers.IPv4) {
	atomic.AddUint64(&ip_bytes.Counter, uint64(ip.Length))
}

// END OF IPCtr
