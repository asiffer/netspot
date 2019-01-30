package counters

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
)

// CounterConstructor is a generic counter constructor
type CounterConstructor func() BaseCtrInterface

// AVAILABLE_COUNTERS maps counter names to their constructor.
// The constructor must be registered in each counter init() function.
var AVAILABLE_COUNTERS map[string]CounterConstructor = make(map[string]CounterConstructor)

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
	for k, _ := range AVAILABLE_COUNTERS {
		list = append(list, k)
	}
	return list
}

// Register aims to add implemented counters to the slice AVAILABLE_COUNTERS
func Register(name string, sc CounterConstructor) error {
	_, exists := AVAILABLE_COUNTERS[name]
	if exists {
		msg := fmt.Sprintf("The counter %s is already available", name)
		log.Error().Msg(msg)
		return errors.New(msg)
	} else {
		AVAILABLE_COUNTERS[name] = sc
		// log.Debug().Msgf("The counter %s is now available", name)
		return nil
	}
}

type BaseCtr struct {
	Running bool        // running state
	Sig     chan uint8  // receive signals
	Val     chan uint64 // send values
}

func NewBaseCtr() BaseCtr {
	return BaseCtr{
		Running: false,
		Sig:     make(chan uint8),
		Val:     make(chan uint64)}
}

func (ctr *BaseCtr) ValPipe() chan uint64 {
	return ctr.Val
}

func (ctr *BaseCtr) SigPipe() chan uint8 {
	return ctr.Sig
}

func (ctr *BaseCtr) IsRunning() bool {
	return ctr.Running
}

func (ctr *BaseCtr) SwitchRunningOn() {
	ctr.Running = true
}

func (ctr *BaseCtr) SwitchRunningOff() {
	ctr.Running = false
}

func Run(ctr BaseCtrInterface) {
	switch ctr.(type) {
	case IpCtrInterface:
		ipctr, ok := ctr.(IpCtrInterface)
		if ok {
			RunIpCtr(ipctr)
		}
	case TcpCtrInterface:
		tcpctr, ok := ctr.(TcpCtrInterface)
		if ok {
			RunTcpCtr(tcpctr)
		}
	case IcmpCtrInterface:
		icmpctr, ok := ctr.(IcmpCtrInterface)
		if ok {
			RunIcmpCtr(icmpctr)
		}
	default:
		fmt.Println("Unknown interface")
	}
}
