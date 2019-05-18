// tcp_nb_uniq_src_port.go

package counters

import (
	"sync"

	"github.com/google/gopacket/layers"
)

func init() {
	Register("NB_UNIQ_SRC_PORT", func() BaseCtrInterface {
		return &NbUniqSrcPort{
			TCPCtr: NewTCPCtr(),
			Port:   make(map[uint16]bool)}
	})
}

// NbUniqSrcPort gives the number of unique source addresses
type NbUniqSrcPort struct {
	TCPCtr
	Port map[uint16]bool
	mux  sync.RWMutex
}

// Name returns the name of the counter (method of BaseCtrInterface)
func (*NbUniqSrcPort) Name() string {
	return "NB_UNIQ_SRC_PORT"
}

// Value returns the current value of the counter (method of BaseCtrInterface)
func (nudp *NbUniqSrcPort) Value() uint64 {
	return uint64(len(nudp.Port))
}

// Reset resets the counter
func (nudp *NbUniqSrcPort) Reset() {
	nudp.mux.Lock()
	for k := range nudp.Port {
		delete(nudp.Port, k)
	}
	nudp.mux.Unlock()
}

// Process update the counter according to data it receives
func (nudp *NbUniqSrcPort) Process(tcp *layers.TCP) {
	nudp.mux.Lock()
	nudp.Port[uint16(tcp.SrcPort)] = true
	nudp.mux.Unlock()
}

// END OF NbUniqSrcPort
