// ip_nb_uniq_dst_addr.go

package counters

import (
	"sync"

	"github.com/google/gopacket/layers"
)

func init() {
	Register("NB_UNIQ_DST_ADDR", func() BaseCtrInterface {
		return &NbUniqDstAddr{
			IPCtr: NewIPCtr(),
			Addr:  make(map[string]bool)}
	})
}

// NbUniqDstAddr gives the number of unique source addresses
type NbUniqDstAddr struct {
	IPCtr
	Addr map[string]bool
	mux  sync.RWMutex
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
	for k := range nuda.Addr {
		delete(nuda.Addr, k)
	}
	nuda.mux.Unlock()
}

// Process update the counter according to data it receives
func (nuda *NbUniqDstAddr) Process(ip *layers.IPv4) {
	nuda.mux.Lock()
	nuda.Addr[ip.DstIP.String()] = true
	nuda.mux.Unlock()
}

// END OF NbUniqDstAddr
