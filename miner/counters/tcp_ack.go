// tcp_ack.go

package counters

import (
	"sync/atomic"

	"github.com/google/gopacket/layers"
)

func init() {
	Register(&ACK{Counter: 0})
}

// ACK stores the number of ACK packets (TCP)
type ACK struct {
	BaseCtr
	Counter uint64
}

// Name returns the name of the counter (method of BaseCtrInterface)
func (*ACK) Name() string {
	return "ACK"
}

// Value returns the current value of the counter (method of BaseCtrInterface)
func (ack *ACK) Value() uint64 {
	return atomic.LoadUint64(&ack.Counter)
}

// Reset resets the counter
func (ack *ACK) Reset() {
	atomic.StoreUint64(&ack.Counter, 0)
}

// Process update the counter according to data it receives
func (ack *ACK) Process(tcp *layers.TCP) {
	if tcp.ACK {
		atomic.AddUint64(&ack.Counter, 1)
	}
}
