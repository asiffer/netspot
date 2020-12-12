// tcp_nb_uniq_dst_port.go

package counters

import (
	"sync"

	"github.com/google/gopacket/layers"
)

func init() {
	Register(&NbUniqDstPort{port: make(map[uint16]bool)})
}

// NbUniqDstPort gives the number of unique source addresses
type NbUniqDstPort struct {
	BaseCtr
	mux  sync.Mutex
	port map[uint16]bool
}

// Name returns the name of the counter (method of BaseCtrInterface)
func (*NbUniqDstPort) Name() string {
	return "NB_UNIQ_DST_PORT"
}

// Value returns the current value of the counter (method of BaseCtrInterface)
func (nudp *NbUniqDstPort) Value() uint64 {
	return uint64(len(nudp.port))
}

// Reset resets the counter
func (nudp *NbUniqDstPort) Reset() {
	nudp.mux.Lock()
	defer nudp.mux.Unlock()
	nudp.port = make(map[uint16]bool)
}

// Process update the counter according to data it receives
func (nudp *NbUniqDstPort) Process(tcp *layers.TCP) {
	nudp.mux.Lock()
	defer nudp.mux.Unlock()
	nudp.port[uint16(tcp.DstPort)] = true
}

// END OF NbUniqDstPort
