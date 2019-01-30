// tcp.go
package counters

import (
	"github.com/google/gopacket/layers"
)

// Interface to define a Counter
// The paramount method is obviously 'process'
type TcpCtrInterface interface {
	BaseCtrInterface
	Process(*layers.TCP)       // method to process a packet
	LayPipe() chan *layers.TCP // receive packets
}

type TcpCtr struct {
	BaseCtr
	Lay chan *layers.TCP
}

func NewTcpCtr() TcpCtr {
	return TcpCtr{
		BaseCtr: NewBaseCtr(),
		Lay:     make(chan *layers.TCP)}
}

func (ctr *TcpCtr) LayPipe() chan *layers.TCP {
	return ctr.Lay
}

func RunTcpCtr(ctr TcpCtrInterface) {
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
