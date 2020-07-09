// pkt_time.go

package counters

import (
	"sync/atomic"

	"github.com/google/gopacket"
)

func init() {
	Register(&SOURCE_TIME{Counter: 0})
}

// SOURCE_TIME stores the timestamp of the packets
type SOURCE_TIME struct {
	BaseCtr
	Counter uint64
}

// Name returns the name of the counter (method of BaseCtrInterface)
func (*SOURCE_TIME) Name() string {
	return "SOURCE_TIME"
}

// Value returns the current value of the counter (method of BaseCtrInterface)
func (tim *SOURCE_TIME) Value() uint64 {
	return atomic.LoadUint64(&tim.Counter)
}

// Reset resets the counter
func (tim *SOURCE_TIME) Reset() {
	// do nothing
}

// Process update the counter according to data it receives
func (tim *SOURCE_TIME) Process(pkt gopacket.Packet) {
	nano := pkt.Metadata().Timestamp.UnixNano()
	atomic.StoreUint64(&tim.Counter, uint64(nano))
}
