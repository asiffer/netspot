// udp.go

package counters

import (
	"github.com/google/gopacket/layers"
)

// UDPCtrInterface is the interface defining a UDP counter
// The paramount method is obviously 'process'
type UDPCtrInterface interface {
	BaseCtrInterface
	Process(*layers.UDP)       // method to process a packet
	LayPipe() chan *layers.UDP // receive packets
}

// UDPCtr is the generic UDP counter (inherits from BaseCtr)
type UDPCtr struct {
	BaseCtr
	Lay chan *layers.UDP
}

// NewUDPCtr is the generic constructor of an UDP counter
func NewUDPCtr() UDPCtr {
	return UDPCtr{
		BaseCtr: NewBaseCtr(),
		Lay:     make(chan *layers.UDP)}
}

// LayPipe returns the UDP layer channel of the UDP counter
func (ctr *UDPCtr) LayPipe() chan *layers.UDP {
	return ctr.Lay
}

// RunUDPCtr starts an UDP counter
func RunUDPCtr(ctr UDPCtrInterface) {
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
		case udp := <-ctr.LayPipe(): // process the packet
			ctr.Process(udp)
		}
	}
}
