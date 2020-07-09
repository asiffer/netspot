// udp.go

package counters

import (
	"sync"

	"github.com/google/gopacket/layers"
)

// UDPCtrInterface is the interface defining a UDP counter
// The paramount method is obviously 'process'
type UDPCtrInterface interface {
	BaseCtrInterface
	Process(*layers.UDP) // method to process a packet
}

// RunUDPCtr starts an UDP counter
func RunUDPCtr(ctr UDPCtrInterface, com chan uint64, input <-chan *layers.UDP, wg *sync.WaitGroup) {
	// reset
	ctr.Reset()
	// loop
	for {
		select {
		case signal := <-com:
			switch signal {
			case STOP: // stop the counter
				defer wg.Done()
				return
			case GET: // return the value
				com <- ctr.Value()
			case RESET: // reset
				ctr.Reset()
			case FLUSH: // return the value and reset
				com <- ctr.Value()
				ctr.Reset()
			case TERMINATE:
				// process the remaining packet
				for udp := range input {
					ctr.Process(udp)
				}
				defer wg.Done()
				return
			}
		case udp := <-input: // process the packet
			ctr.Process(udp)
		}
	}
}
