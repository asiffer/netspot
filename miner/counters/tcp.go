// tcp.go

package counters

import (
	"github.com/google/gopacket/layers"
)

// TCPCtrInterface is the interface defining a TCP counter
// The paramount method is obviously 'process'
type TCPCtrInterface interface {
	BaseCtrInterface
	Process(*layers.TCP)       // method to process a packet
	LayPipe() chan *layers.TCP // receive packets
}

// TCPCtr is the generic TCP counter (inherits from BaseCtr)
type TCPCtr struct {
	BaseCtr
	Lay chan *layers.TCP
}

// NewTCPCtr is the generic constructor of an TCP counter
func NewTCPCtr() TCPCtr {
	return TCPCtr{
		BaseCtr: NewBaseCtr(),
		Lay:     make(chan *layers.TCP)}
}

// LayPipe returns the TCP layer channel of the TCP counter
func (ctr *TCPCtr) LayPipe() chan *layers.TCP {
	return ctr.Lay
}

// RunTCPCtr starts an TCP counter
func RunTCPCtr(ctr TCPCtrInterface) {
	ctr.SwitchRunningOn()
	for {
		select {
		case sig := <-ctr.SigPipe():
			switch sig {
			case 0: // stop the counter
				ctr.SwitchRunningOff()
				return
			case 1: // return the value
				ctr.ValPipe() <- ctr.Value()
			case 2:
				ctr.Reset()
			}
		case tcp := <-ctr.LayPipe(): // process the packet
			ctr.Process(tcp)
			// default:
			// nothing (non blocking)
		}
	}
}
