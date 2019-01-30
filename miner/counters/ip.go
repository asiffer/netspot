// ip.go
package counters

import (
	"github.com/google/gopacket/layers"
)

// COMMON FEATURES

type IpCtrInterface interface {
	BaseCtrInterface
	Process(*layers.IPv4) // method to process a packet
	LayPipe() chan *layers.IPv4
}

type IpCtr struct {
	BaseCtr
	Lay chan *layers.IPv4
}

func NewIpCtr() IpCtr {
	return IpCtr{
		BaseCtr: NewBaseCtr(),
		Lay:     make(chan *layers.IPv4)}
}

func (ctr *IpCtr) LayPipe() chan *layers.IPv4 {
	return ctr.Lay
}

func RunIpCtr(ctr IpCtrInterface) {
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
		case ip := <-ctr.LayPipe(): // process the packet
			ctr.Process(ip)
			//default:
			// nothing (non blocking)
		}
	}
}
