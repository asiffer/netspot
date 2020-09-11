// ip.go

package counters

import (
	"sync"

	"github.com/google/gopacket/layers"
)

// IPv4CtrInterface is the interface defining an IP counter
// The paramount method is obviously 'process'
type IPv4CtrInterface interface {
	BaseCtrInterface
	Process(*layers.IPv4) // method to process a packet
}

// RunIPv4Ctr starts an IP counter
func RunIPv4Ctr(ctr IPv4CtrInterface, com chan uint64, input <-chan *layers.IPv4, wg *sync.WaitGroup) {
	// reset
	ctr.Reset()
	// defer done task
	defer wg.Done()
	// loop
	for {
		select {
		case signal := <-com:
			switch signal {
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
				for ip := range input {
					ctr.Process(ip)
				}
				return
			}
		case ip := <-input: // process the packet
			ctr.Process(ip)
		}
	}
}
