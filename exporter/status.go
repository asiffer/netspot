package exporter

import "sync"

// Status is a basic object which
// stores the exporter status
type Status struct {
	mutex sync.Mutex
	value bool
}

var started = NewStatus()

// NewStatus init the internal  status
func NewStatus() *Status {
	return &Status{mutex: sync.Mutex{}, value: false}
}

// Status return true if the exporter has started
func (es *Status) Status() bool {
	es.mutex.Lock()
	defer es.mutex.Unlock()
	return es.value
}

// Begin sets the status to "started"
func (es *Status) Begin() {
	es.mutex.Lock()
	es.value = true
	es.mutex.Unlock()
}

// End sets the status to "not started"
func (es *Status) End() {
	es.mutex.Lock()
	es.value = false
	es.mutex.Unlock()
}
