// tcp_rst.go

package counters

import (
	"sync/atomic"

	"github.com/google/gopacket/layers"
)

func init() {
	Register("RST", func() BaseCtrInterface {
		return &RST{
			TCPCtr:  NewTCPCtr(),
			Counter: 0}
	})
}

// RST stores the number of RST packets (TCP)
type RST struct {
	TCPCtr
	Counter uint64
}

// Name returns the name of the counter (method of BaseCtrInterface)
func (*RST) Name() string {
	return "RST"
}

// Value returns the current value of the counter (method of BaseCtrInterface)
func (rst *RST) Value() uint64 {
	return atomic.LoadUint64(&rst.Counter)
}

// Reset resets the counter
func (rst *RST) Reset() {
	atomic.StoreUint64(&rst.Counter, 0)
}

// Process update the counter according to data it receives
func (rst *RST) Process(tcp *layers.TCP) {
	if tcp.RST {
		atomic.AddUint64(&rst.Counter, 1)
	}
}
