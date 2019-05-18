// tcp_nb_uniq_dst_port.go

package counters

import (
	"sync"

	"github.com/google/gopacket/layers"
)

func init() {
	Register("NB_UNIQ_DST_PORT", func() BaseCtrInterface {
		return &NbUniqDstPort{
			TCPCtr: NewTCPCtr(),
			Port:   make(map[uint16]bool)}
	})
}

// NbUniqDstPort gives the number of unique source addresses
type NbUniqDstPort struct {
	TCPCtr
	Port map[uint16]bool
	mux  sync.RWMutex
}

// Name returns the name of the counter (method of BaseCtrInterface)
func (*NbUniqDstPort) Name() string {
	return "NB_UNIQ_DST_PORT"
}

// Value returns the current value of the counter (method of BaseCtrInterface)
func (nudp *NbUniqDstPort) Value() uint64 {
	return uint64(len(nudp.Port))
}

// Reset resets the counter
func (nudp *NbUniqDstPort) Reset() {
	nudp.mux.Lock()
	for k := range nudp.Port {
		delete(nudp.Port, k)
	}
	nudp.mux.Unlock()
}

// Process update the counter according to data it receives
func (nudp *NbUniqDstPort) Process(tcp *layers.TCP) {
	nudp.mux.Lock()
	nudp.Port[uint16(tcp.DstPort)] = true
	nudp.mux.Unlock()
}

// END OF NbUniqDstPort
