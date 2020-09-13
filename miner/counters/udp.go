// udp.go

package counters

import (
	"github.com/google/gopacket/layers"
)

// UDPCtrInterface is the interface defining a UDP counter
// The paramount method is obviously 'process'
type UDPCtrInterface interface {
	BaseCtrInterface
	Process(*layers.UDP) // method to process a packet
}
