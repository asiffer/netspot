// pkt.go

package counters

import (
	"github.com/google/gopacket"
)

// PktCtrInterface is the interface defining an packet counter.
// This counter processes all the packets.
// The paramount method is obviously 'process'
type PktCtrInterface interface {
	BaseCtrInterface
	Process(gopacket.Packet) // method to process a packet
	LayPipe() chan gopacket.Packet
}

// PktCtr is the generic Pkt counter (inherits from BaseCtr)
type PktCtr struct {
	BaseCtr
	Lay chan gopacket.Packet
}

// NewPktCtr is the generic constructor of an Pkt counter
func NewPktCtr() PktCtr {
	return PktCtr{
		BaseCtr: NewBaseCtr(),
		Lay:     make(chan gopacket.Packet)}
}

// LayPipe returns the layer channel of the counter
func (ctr *PktCtr) LayPipe() chan gopacket.Packet {
	return ctr.Lay
}

// RunPktCtr starts a PKT counter
func RunPktCtr(ctr PktCtrInterface) {
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
		case pkt := <-ctr.LayPipe(): // process the packet
			ctr.Process(pkt)

		}
	}
}
