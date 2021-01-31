package miner

import "sync"

// SniffingStatus is a basic object which
// stores the miner status
type SniffingStatus struct {
	mutex sync.Mutex
	value bool
}

var sniffing = NewSniffingStatus() // tells if the package is currently sniffing

// NewSniffingStatus init the internal sniffing status
func NewSniffingStatus() *SniffingStatus {
	return &SniffingStatus{mutex: sync.Mutex{}, value: false}
}

// Status return true if the miner is sniffing
func (ss *SniffingStatus) Status() bool {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()
	return ss.value
}

// Begin sets tue status to "sniffing"
func (ss *SniffingStatus) Begin() {
	ss.mutex.Lock()
	ss.value = true
	ss.mutex.Unlock()
}

// End sets tue status to "not sniffing"
func (ss *SniffingStatus) End() {
	ss.mutex.Lock()
	ss.value = false
	ss.mutex.Unlock()
}
