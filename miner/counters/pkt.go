// pkt.go

package counters

import (
	"github.com/google/gopacket"
)

// PktCtrInterface is the interface defining an packet counter.
// This counter processes all the packets.
// The paramount method is obviously 'process'
type PktCtrInterface interface {
	BaseCtrInterface
	Process(gopacket.Packet) // method to process a packet
}
