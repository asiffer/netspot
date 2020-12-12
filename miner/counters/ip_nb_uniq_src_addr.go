// ip_nb_uniq_src_addr.go

package counters

import (
	"sync"

	"github.com/google/gopacket/layers"
)

func init() {
	Register(&NbUniqSrcAddr{Addr: make(map[string]bool)})
}

// NbUniqSrcAddr gives the number of unique source addresses
type NbUniqSrcAddr struct {
	BaseCtr
	Addr map[string]bool
	mux  sync.Mutex
}

// Name returns the name of the counter (method of BaseCtrInterface)
func (*NbUniqSrcAddr) Name() string {
	return "NB_UNIQ_SRC_ADDR"
}

// Value returns the current value of the counter (method of BaseCtrInterface)
func (nusa *NbUniqSrcAddr) Value() uint64 {
	return uint64(len(nusa.Addr))
}

// Reset resets the counter
func (nusa *NbUniqSrcAddr) Reset() {
	nusa.mux.Lock()
	defer nusa.mux.Unlock()
	nusa.Addr = make(map[string]bool)
}

// Process update the counter according to data it receives
func (nusa *NbUniqSrcAddr) Process(ip *layers.IPv4) {
	nusa.mux.Lock()
	defer nusa.mux.Unlock()
	nusa.Addr[ip.SrcIP.String()] = true
}

// END OF NbUniqSrcAddr
