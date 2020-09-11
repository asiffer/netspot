// tcp.go

package counters

import (
	"sync"

	"github.com/google/gopacket/layers"
)

// TCPCtrInterface is the interface defining a TCP counter
// The paramount method is obviously 'process'
type TCPCtrInterface interface {
	BaseCtrInterface
	Process(*layers.TCP) // method to process a packet
}

// RunTCPCtr starts an TCP counter
func RunTCPCtr(ctr TCPCtrInterface, com chan uint64, input <-chan *layers.TCP, wg *sync.WaitGroup) {
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
				for tcp := range input {
					ctr.Process(tcp)
				}
				return
			}
		case tcp := <-input: // process the packet
			ctr.Process(tcp)
		}
	}
}
