// tcp_syn.go
package counters

import (
	"sync/atomic"

	"github.com/google/gopacket/layers"
)

func init() {
	Register("SYN", func() BaseCtrInterface {
		return &SYN{
			TcpCtr:  NewTcpCtr(),
			Counter: 0}
	})
}

// SYN
// Store the number of SYN packets (TCP)
type SYN struct {
	TcpCtr
	Counter uint64
}

// Generic function (BaseCtrInterface)
func (*SYN) Name() string {
	return "SYN"
}

func (syn *SYN) Value() uint64 {
	return atomic.LoadUint64(&syn.Counter)
}

func (syn *SYN) Reset() {
	atomic.StoreUint64(&syn.Counter, 0)
}

// Specific function (TcpCtr)
func (syn *SYN) Process(tcp *layers.TCP) {
	if tcp.SYN {
		atomic.AddUint64(&syn.Counter, 1)
	}
}

// END OF SYN
