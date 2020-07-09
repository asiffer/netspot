// pkt_real_time.go

package counters

import (
	"sync/atomic"
	"time"

	"github.com/google/gopacket"
)

func init() {
	Register(&REAL_TIME{Counter: 0})
}

// REAL_TIME stores the timestamp of the packets
type REAL_TIME struct {
	BaseCtr
	Counter uint64
}

// Name returns the name of the counter (method of BaseCtrInterface)
func (*REAL_TIME) Name() string {
	return "REAL_TIME"
}

// Value returns the current value of the counter (method of BaseCtrInterface)
func (tim *REAL_TIME) Value() uint64 {
	return atomic.LoadUint64(&tim.Counter)
}

// Reset resets the counter
func (tim *REAL_TIME) Reset() {
	// do nothing
}

// Process update the counter according to data it receives
func (tim *REAL_TIME) Process(pkt gopacket.Packet) {
	nano := time.Now().UnixNano()
	atomic.StoreUint64(&tim.Counter, uint64(nano))

}
