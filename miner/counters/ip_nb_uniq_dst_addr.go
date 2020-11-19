// ip_nb_uniq_dst_addr.go

package counters

import (
	"sync"

	"github.com/google/gopacket/layers"
)

func init() {
	Register(&NbUniqDstAddr{Addr: make(map[string]bool)})
}

// NbUniqDstAddr gives the number of unique source addresses
type NbUniqDstAddr struct {
	BaseCtr
	Addr map[string]bool
	mux  sync.Mutex
}

// Name returns the name of the counter (method of BaseCtrInterface)
func (*NbUniqDstAddr) Name() string {
	return "NB_UNIQ_DST_ADDR"
}

// Value returns the current value of the counter (method of BaseCtrInterface)
func (nuda *NbUniqDstAddr) Value() uint64 {
	return uint64(len(nuda.Addr))
}

// Reset resets the counter
func (nuda *NbUniqDstAddr) Reset() {
	nuda.mux.Lock()
	defer nuda.mux.Unlock()
	nuda.Addr = make(map[string]bool)
}

// Process update the counter according to data it receives
func (nuda *NbUniqDstAddr) Process(ip *layers.IPv4) {
	nuda.mux.Lock()
	defer nuda.mux.Unlock()
	nuda.Addr[ip.DstIP.String()] = true
}

// END OF NbUniqDstAddr
