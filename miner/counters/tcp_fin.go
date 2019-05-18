// tcp_fin.go

package counters

import (
	"sync/atomic"

	"github.com/google/gopacket/layers"
)

func init() {
	Register("FIN", func() BaseCtrInterface {
		return &FIN{
			TCPCtr:  NewTCPCtr(),
			Counter: 0}
	})
}

// FIN stores the number of FIN packets (TCP)
type FIN struct {
	TCPCtr
	Counter uint64
}

// Name returns the name of the counter (method of BaseCtrInterface)
func (*FIN) Name() string {
	return "FIN"
}

// Value returns the current value of the counter (method of BaseCtrInterface)
func (fin *FIN) Value() uint64 {
	return atomic.LoadUint64(&fin.Counter)
}

// Reset resets the counter
func (fin *FIN) Reset() {
	atomic.StoreUint64(&fin.Counter, 0)
}

// Process update the counter according to data it receives
func (fin *FIN) Process(tcp *layers.TCP) {
	if tcp.FIN {
		atomic.AddUint64(&fin.Counter, 1)
	}
}
