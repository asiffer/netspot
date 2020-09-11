// icmp.go

package counters

import (
	"sync"

	"github.com/google/gopacket/layers"
)

// ICMPCtrInterface si the interface defining an ICMP Counter
// The paramount method is obviously 'process'
type ICMPv4CtrInterface interface {
	BaseCtrInterface
	Process(*layers.ICMPv4) // method to process a packet
}

// RunICMPv4Ctr starts the ICMP counter
func RunICMPv4Ctr(ctr ICMPv4CtrInterface, com chan uint64, input <-chan *layers.ICMPv4, wg *sync.WaitGroup) {
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
				for icmp := range input {
					ctr.Process(icmp)
				}
				return
			}
		case icmp := <-input: // process the packet
			ctr.Process(icmp)
		}
	}
}
