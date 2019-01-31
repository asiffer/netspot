// icmp.go

package counters

import (
	"github.com/google/gopacket/layers"
)

// ICMPCtrInterface si the interface defining an ICMP Counter
// The paramount method is obviously 'process'
type ICMPCtrInterface interface {
	BaseCtrInterface
	Process(*layers.ICMPv4)       // method to process a packet
	LayPipe() chan *layers.ICMPv4 // receive packets
}

// ICMPCtr is the generic object for ICMP counter
type ICMPCtr struct {
	BaseCtr
	Lay chan *layers.ICMPv4
}

// NewICMPCtr is the generic constructor of an ICMP counter
func NewICMPCtr() ICMPCtr {
	return ICMPCtr{
		BaseCtr: NewBaseCtr(),
		Lay:     make(chan *layers.ICMPv4)}
}

// LayPipe returns the ICMP layer channel of the ICMP counter
func (ctr *ICMPCtr) LayPipe() chan *layers.ICMPv4 {
	return ctr.Lay
}

// RunICMPCtr starts the ICMP counter
func RunICMPCtr(ctr ICMPCtrInterface) {
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
		}
	}
}
