// arp.go

package counters

import (
	"sync"

	"github.com/google/gopacket/layers"
)

// ARPCtrInterface is the interface defining an ARP counter
// The paramount method is obviously 'process'
type ARPCtrInterface interface {
	BaseCtrInterface
	Process(*layers.ARP) // method to process a packet
}

// RunARPCtr starts an ARP counter
func RunARPCtr(ctr ARPCtrInterface, com chan uint64, input <-chan *layers.ARP, wg *sync.WaitGroup) {
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
				for arp := range input {
					ctr.Process(arp)
				}
				defer wg.Done()
				return
			}
		case arp := <-input: // process the packet
			ctr.Process(arp)
		}
	}
}
