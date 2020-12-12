// ip.go

package counters

import (
	"github.com/google/gopacket/layers"
)

// IPv4CtrInterface is the interface defining an IP counter
// The paramount method is obviously 'process'
type IPv4CtrInterface interface {
	BaseCtrInterface
	Process(*layers.IPv4) // method to process a packet
}
