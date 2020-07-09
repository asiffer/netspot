// counters.go

// Package counters implements a generic counter interface and some specific
// counter objects.
package counters

import (
	"fmt"
)

const (
	// STOP stops a counter
	STOP uint64 = 0
	// GET returns the value of the counter
	GET uint64 = 1
	// RESET resets the counter
	RESET uint64 = 2
	// FLUSH returns the value of the counter and reset it
	FLUSH uint64 = 3
	// TERMINATE stop the counter but ensure it has
	// no packet to parse anymore
	TERMINATE uint64 = 4
)

// AvailableCounters maps counter names to their constructor.
// The constructor must be registered in each counter init() function.
var AvailableCounters = make(map[string]BaseCtrInterface)

// BaseCtrInterface represents the method a counter must implement
type BaseCtrInterface interface {
	Name() string  // the name of the counter
	Value() uint64 // to send a value to the right pipe
	Reset()        // method to reset the counter
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
func Register(ctr BaseCtrInterface) error {
	if _, exists := AvailableCounters[ctr.Name()]; exists {
		return fmt.Errorf("The counter %s is already available", ctr.Name())
	}
	AvailableCounters[ctr.Name()] = ctr
	return nil
}

// BaseCtr is the basic counter object
type BaseCtr int
