// ip_nb_uniq_src_addr.go

package counters

import (
	"sync"

	"github.com/google/gopacket/layers"
)

func init() {
	Register("NB_UNIQ_SRC_ADDR", func() BaseCtrInterface {
		return &NbUniqSrcAddr{
			IPCtr: NewIPCtr(),
			Addr:  make(map[string]bool)}
	})
}

// NbUniqSrcAddr gives the number of unique source addresses
type NbUniqSrcAddr struct {
	IPCtr
	Addr map[string]bool
	mux  sync.RWMutex
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
	for k := range nusa.Addr {
		delete(nusa.Addr, k)
	}
	nusa.mux.Unlock()
}

// Process update the counter according to data it receives
func (nusa *NbUniqSrcAddr) Process(ip *layers.IPv4) {
	nusa.mux.Lock()
	nusa.Addr[ip.SrcIP.String()] = true
	nusa.mux.Unlock()
}

// END OF NbUniqSrcAddr
