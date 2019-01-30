// tcp_ack.go
package counters

import (
	"sync/atomic"

	"github.com/google/gopacket/layers"
)

func init() {
	Register("ACK", func() BaseCtrInterface {
		return &ACK{
			TcpCtr:  NewTcpCtr(),
			Counter: 0}
	})
}

// ACK
// Store the number of ACK packets (TCP)
type ACK struct {
	TcpCtr
	Counter uint64
}

// Generic function (BaseCtrInterface)
func (*ACK) Name() string {
	return "ACK"
}

func (ack *ACK) Value() uint64 {
	return atomic.LoadUint64(&ack.Counter)
}

func (ack *ACK) Reset() {
	atomic.StoreUint64(&ack.Counter, 0)
}

// Specific function (TcpCtr)
func (ack *ACK) Process(tcp *layers.TCP) {
	if tcp.ACK {
		atomic.AddUint64(&ack.Counter, 1)
	}
}
