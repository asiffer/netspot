// tcp_ack.go

package counters

import (
	"sync"
	"sync/atomic"

	"github.com/google/gopacket/layers"
)

func init() {
	Register("ACK", func() BaseCtrInterface {
		return &ACK{
			TCPCtr:  NewTCPCtr(),
			Counter: 0}
	})
}

var ackmux sync.Mutex

// ACK stores the number of ACK packets (TCP)
type ACK struct {
	TCPCtr
	Counter uint64
}

// Name returns the name of the counter (method of BaseCtrInterface)
func (*ACK) Name() string {
	return "ACK"
}

// Value returns the current value of the counter (method of BaseCtrInterface)
func (ack *ACK) Value() uint64 {
	// ackmux.Lock()
	// defer ackmux.Unlock()
	// return ack.Counter
	return atomic.LoadUint64(&ack.Counter)
}

// Reset resets the counter
func (ack *ACK) Reset() {
	// ackmux.Lock()
	// ack.Counter = 0
	atomic.StoreUint64(&ack.Counter, 0)
	// ackmux.Unlock()
}

// Process update the counter according to data it receives
func (ack *ACK) Process(tcp *layers.TCP) {
	if tcp.ACK {
		// ackmux.Lock()
		// ack.Counter++
		atomic.AddUint64(&ack.Counter, 1)
		// ackmux.Unlock()
	}
}
