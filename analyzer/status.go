package analyzer

import "sync"

// Status is a basic object which
// stores the analyzer status
type Status struct {
	mutex sync.Mutex
	value bool
}

var running = NewStatus()

// NewStatus init the internal status
func NewStatus() *Status {
	return &Status{mutex: sync.Mutex{}, value: false}
}

// Status return true if the analyzer is running
func (es *Status) Status() bool {
	es.mutex.Lock()
	defer es.mutex.Unlock()
	return es.value
}

// Begin sets the status to "running"
func (es *Status) Begin() {
	es.mutex.Lock()
	es.value = true
	es.mutex.Unlock()
}

// End sets tue status to "not running"
func (es *Status) End() {
	es.mutex.Lock()
	es.value = false
	es.mutex.Unlock()
}
