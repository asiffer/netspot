// ip_nb_uniq_dst_addr.go
package counters

import (
	"sync"

	"github.com/google/gopacket/layers"
)

func init() {
	Register("NB_UNIQ_DST_ADDR", func() BaseCtrInterface {
		return &NB_UNIQ_DST_ADDR{
			IpCtr: NewIpCtr(),
			Addr:  make(map[string]bool)}
	})
}

// NB_UNIQ_DST_ADDR
// The number of unique source addresses
type NB_UNIQ_DST_ADDR struct {
	IpCtr
	Addr map[string]bool
	mux  sync.RWMutex
}

// Generic function (BaseCtrInterface)
func (*NB_UNIQ_DST_ADDR) Name() string {
	return "NB_UNIQ_DST_ADDR"
}

func (nuda *NB_UNIQ_DST_ADDR) Value() uint64 {
	return uint64(len(nuda.Addr))
}

func (nuda *NB_UNIQ_DST_ADDR) Reset() {
	nuda.mux.Lock()
	for k, _ := range nuda.Addr {
		delete(nuda.Addr, k)
	}
	nuda.mux.Unlock()
}

// Specific function (IpCtr)

func (nuda *NB_UNIQ_DST_ADDR) Process(ip *layers.IPv4) {
	nuda.mux.Lock()
	nuda.Addr[ip.SrcIP.String()] = true
	nuda.mux.Unlock()
}

// END OF NB_UNIQ_DST_ADDR
