// pkt.go

package counters

import (
	"sync"

	"github.com/google/gopacket"
)

// PktCtrInterface is the interface defining an packet counter.
// This counter processes all the packets.
// The paramount method is obviously 'process'
type PktCtrInterface interface {
	BaseCtrInterface
	Process(gopacket.Packet) // method to process a packet
}

// RunPktCtr starts a PKT counter
func RunPktCtr(ctr PktCtrInterface, com chan uint64, input <-chan gopacket.Packet, wg *sync.WaitGroup) {
	// reset
	ctr.Reset()
	// defer done task
	defer wg.Done()
	// loop
	for {
		select {
		case sig := <-com:
			switch sig {
			case STOP: // stop the counter
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
				for pkt := range input {
					ctr.Process(pkt)
				}
				return
			}
		case pkt := <-input: // process the packet
			ctr.Process(pkt)
		}
	}
}
