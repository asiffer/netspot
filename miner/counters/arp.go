// arp.go

package counters

import (
	"github.com/google/gopacket/layers"
)

// ARPCtrInterface is the interface defining an ARP counter
// The paramount method is obviously 'process'
type ARPCtrInterface interface {
	BaseCtrInterface
	Process(*layers.ARP) // method to process a packet
}
