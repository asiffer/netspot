// icmp.go

package counters

import (
	"github.com/google/gopacket/layers"
)

// ICMPv4CtrInterface si the interface defining an ICMP Counter
// The paramount method is obviously 'process'
type ICMPv4CtrInterface interface {
	BaseCtrInterface
	Process(*layers.ICMPv4) // method to process a packet
}
