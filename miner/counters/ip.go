// ip.go

package counters

import (
	"github.com/google/gopacket/layers"
)

// IPCtrInterface is the interface defining an IP counter
// The paramount method is obviously 'process'
type IPCtrInterface interface {
	BaseCtrInterface
	Process(*layers.IPv4) // method to process a packet
	LayPipe() chan *layers.IPv4
}

// IPCtr is the generic IP counter (inherits from BaseCtr)
type IPCtr struct {
	BaseCtr
	Lay chan *layers.IPv4
}

// NewIPCtr is the generic constructor of an IP counter
func NewIPCtr() IPCtr {
	return IPCtr{
		BaseCtr: NewBaseCtr(),
		Lay:     make(chan *layers.IPv4)}
}

// LayPipe returns the IPv4 layer channel of the IP counter
func (ctr *IPCtr) LayPipe() chan *layers.IPv4 {
	return ctr.Lay
}

// RunIPCtr starts an IP counter
func RunIPCtr(ctr IPCtrInterface) {
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
		case ip := <-ctr.LayPipe(): // process the packet
			ctr.Process(ip)

		}
	}
}
