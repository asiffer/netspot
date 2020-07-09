// tcp_urg.go

package counters

import (
	"sync/atomic"

	"github.com/google/gopacket/layers"
)

func init() {
	Register(&URG{counter: 0})
}

// URG stores the number of URG packets (TCP)
type URG struct {
	BaseCtr
	counter uint64
}

// Name returns the name of the counter (method of BaseCtrInterface)
func (*URG) Name() string {
	return "URG"
}

// Value returns the current value of the counter (method of BaseCtrInterface)
func (urg *URG) Value() uint64 {
	return atomic.LoadUint64(&urg.counter)
}

// Reset resets the counter
func (urg *URG) Reset() {
	atomic.StoreUint64(&urg.counter, 0)
}

// Process update the counter according to data it receives
func (urg *URG) Process(tcp *layers.TCP) {
	if tcp.URG {
		atomic.AddUint64(&urg.counter, 1)
	}
}
