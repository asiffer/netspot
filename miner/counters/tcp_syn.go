// tcp_syn.go

package counters

import (
	"sync/atomic"

	"github.com/google/gopacket/layers"
)

func init() {
	Register(&SYN{counter: 0})
}

// SYN stores the number of SYN packets (TCP)
type SYN struct {
	BaseCtr
	counter uint64
}

// Name returns the name of the counter (method of BaseCtrInterface)
func (*SYN) Name() string {
	return "SYN"
}

// Value returns the current value of the counter (method of BaseCtrInterface)
func (syn *SYN) Value() uint64 {
	return atomic.LoadUint64(&syn.counter)
}

// Reset resets the counter
func (syn *SYN) Reset() {
	atomic.StoreUint64(&syn.counter, 0)
}

// Process update the counter according to data it receives
func (syn *SYN) Process(tcp *layers.TCP) {
	if tcp.SYN {
		atomic.AddUint64(&syn.counter, 1)
	}
}

// END OF SYN
