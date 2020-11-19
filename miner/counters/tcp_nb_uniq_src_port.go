// tcp_nb_uniq_src_port.go

package counters

import (
	"sync"

	"github.com/google/gopacket/layers"
)

func init() {
	Register(&NbUniqSrcPort{port: make(map[uint16]bool)})
}

// NbUniqSrcPort gives the number of unique source addresses
type NbUniqSrcPort struct {
	BaseCtr
	mux  sync.Mutex
	port map[uint16]bool
}

// Name returns the name of the counter (method of BaseCtrInterface)
func (*NbUniqSrcPort) Name() string {
	return "NB_UNIQ_SRC_PORT"
}

// Value returns the current value of the counter (method of BaseCtrInterface)
func (nusp *NbUniqSrcPort) Value() uint64 {
	return uint64(len(nusp.port))
}

// Reset resets the counter
func (nusp *NbUniqSrcPort) Reset() {
	nusp.mux.Lock()
	defer nusp.mux.Unlock()
	nusp.port = make(map[uint16]bool)
}

// Process update the counter according to data it receives
func (nusp *NbUniqSrcPort) Process(tcp *layers.TCP) {
	nusp.mux.Lock()
	defer nusp.mux.Unlock()
	nusp.port[uint16(tcp.SrcPort)] = true
}

// END OF NbUniqSrcPort
