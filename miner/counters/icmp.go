// icmp.go
package counters

import (
	"github.com/google/gopacket/layers"
)

// Interface to define a Counter
// The paramount method is obviously 'process'
type IcmpCtrInterface interface {
	BaseCtrInterface
	Process(*layers.ICMPv4)       // method to process a packet
	LayPipe() chan *layers.ICMPv4 // receive packets
}

type IcmpCtr struct {
	BaseCtr
	Lay chan *layers.ICMPv4
}

func NewIcmpCtr() IcmpCtr {
	return IcmpCtr{
		BaseCtr: NewBaseCtr(),
		Lay:     make(chan *layers.ICMPv4)}
}

func (ctr *IcmpCtr) LayPipe() chan *layers.ICMPv4 {
	return ctr.Lay
}

func RunIcmpCtr(ctr IcmpCtrInterface) {
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
		case icmp := <-ctr.LayPipe(): // process the packet
			ctr.Process(icmp)
			// default:
			// nothing (non blocking)
		}
	}
}
