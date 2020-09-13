// tcp.go

package counters

import (
	"github.com/google/gopacket/layers"
)

// TCPCtrInterface is the interface defining a TCP counter
// The paramount method is obviously 'process'
type TCPCtrInterface interface {
	BaseCtrInterface
	Process(*layers.TCP) // method to process a packet
}
