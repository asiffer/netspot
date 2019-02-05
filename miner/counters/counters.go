package counters

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
)

// CounterConstructor is a generic counter constructor
type CounterConstructor func() BaseCtrInterface

// AvailableCounters maps counter names to their constructor.
// The constructor must be registered in each counter init() function.
var AvailableCounters = make(map[string]CounterConstructor)

// BaseCtrInterface represents the method a counter must implement
type BaseCtrInterface interface {
	Name() string         // the name of the counter
	Value() uint64        // to send a value to the right pipe
	Reset()               // method to reset the counter
	ValPipe() chan uint64 // to access the value pipe
	SigPipe() chan uint8  // to access the signal pipe
	IsRunning() bool      // to check whether the counter is running or not
	SwitchRunningOn()     // to activate the running state
	SwitchRunningOff()    // to desactivate the running state
}

// GetAvailableCounters return the list of the registered counters
func GetAvailableCounters() []string {
	list := make([]string, 0)
	for k := range AvailableCounters {
		list = append(list, k)
	}
	return list
}

// Register aims to add implemented counters to the slice AvailableCounters
func Register(name string, sc CounterConstructor) error {
	_, exists := AvailableCounters[name]
	if exists {
		msg := fmt.Sprintf("The counter %s is already available", name)
		log.Error().Msg(msg)
		return errors.New(msg)
	}
	AvailableCounters[name] = sc
	return nil

}

// BaseCtr is the basic counter object
type BaseCtr struct {
	Running bool        // running state
	Sig     chan uint8  // receive signals
	Val     chan uint64 // send values
}

// NewBaseCtr is the basic counter contructor. It is called
// by specific counters implementation
func NewBaseCtr() BaseCtr {
	return BaseCtr{
		Running: false,
		Sig:     make(chan uint8),
		Val:     make(chan uint64)}
}

// ValPipe returns the Value channel of the counter
func (ctr *BaseCtr) ValPipe() chan uint64 {
	return ctr.Val
}

// SigPipe returns the Signal channel of the counter
func (ctr *BaseCtr) SigPipe() chan uint8 {
	return ctr.Sig
}

// IsRunning check if the counter is running (it is running
// when the function 'Run' has been called)
func (ctr *BaseCtr) IsRunning() bool {
	return ctr.Running
}

// SwitchRunningOn turns on the running state of the counter
func (ctr *BaseCtr) SwitchRunningOn() {
	ctr.Running = true
}

// SwitchRunningOff turns off the running state of the counter
func (ctr *BaseCtr) SwitchRunningOff() {
	ctr.Running = false
}

// Run starts a counter, making it waiting for new incoming packets
func Run(ctr BaseCtrInterface) {
	switch ctr.(type) {
	case IPCtrInterface:
		ipctr, ok := ctr.(IPCtrInterface)
		if ok {
			RunIPCtr(ipctr)
		}
	case TCPCtrInterface:
		tcpctr, ok := ctr.(TCPCtrInterface)
		if ok {
			RunTCPCtr(tcpctr)
		}
	case ICMPCtrInterface:
		icmpctr, ok := ctr.(ICMPCtrInterface)
		if ok {
			RunICMPCtr(icmpctr)
		}
	default:
		if ctr != nil {
			log.Error().Msgf("The type of the counter '%s' is unknown", ctr.Name())
		}
	}
}
