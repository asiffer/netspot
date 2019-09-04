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
	LayPipe() chan *layers.ARP
}

// ARPCtr is the generic IP counter (inherits from BaseCtr)
type ARPCtr struct {
	BaseCtr
	Lay chan *layers.ARP
}

// NewARPCtr is the generic constructor of an IP counter
func NewARPCtr() ARPCtr {
	return ARPCtr{
		BaseCtr: NewBaseCtr(),
		Lay:     make(chan *layers.ARP)}
}

// LayPipe returns the ARP layer channel of the ARP counter
func (ctr *ARPCtr) LayPipe() chan *layers.ARP {
	return ctr.Lay
}

// RunARPCtr starts an ARP counter
func RunARPCtr(ctr ARPCtrInterface) {
	ctr.SwitchRunningOn()
	for {
		select {
		case sig := <-ctr.SigPipe():
			switch sig {
			case STOP: // stop the counter
				ctr.SwitchRunningOff()
				return
			case GET: // return the value
				ctr.ValPipe() <- ctr.Value()
			case RESET: // reset
				ctr.Reset()
			case FLUSH: // return the value and reset
				ctr.ValPipe() <- ctr.Value()
				ctr.Reset()
			}
		case arp := <-ctr.LayPipe(): // process the packet
			ctr.Process(arp)

		}
	}
}
