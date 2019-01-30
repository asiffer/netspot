// ip_nb_uniq_src_addr.go
package counters

import (
	"sync"

	"github.com/google/gopacket/layers"
)

func init() {
	Register("NB_UNIQ_SRC_ADDR", func() BaseCtrInterface {
		return &NB_UNIQ_SRC_ADDR{
			IpCtr: NewIpCtr(),
			Addr:  make(map[string]bool)}
	})
}

// NB_UNIQ_SRC_ADDR
// The number of unique source addresses
type NB_UNIQ_SRC_ADDR struct {
	IpCtr
	Addr map[string]bool
	mux  sync.RWMutex
}

// Generic function (BaseCtrInterface)
func (*NB_UNIQ_SRC_ADDR) Name() string {
	return "NB_UNIQ_SRC_ADDR"
}

func (nusa *NB_UNIQ_SRC_ADDR) Value() uint64 {
	return uint64(len(nusa.Addr))
}

func (nusa *NB_UNIQ_SRC_ADDR) Reset() {
	nusa.mux.Lock()
	for k, _ := range nusa.Addr {
		delete(nusa.Addr, k)
	}
	nusa.mux.Unlock()
}

// Specific function (IpCtr)

func (nusa *NB_UNIQ_SRC_ADDR) Process(ip *layers.IPv4) {
	nusa.mux.Lock()
	nusa.Addr[ip.SrcIP.String()] = true
	nusa.mux.Unlock()
}

// END OF NB_UNIQ_SRC_ADDR
