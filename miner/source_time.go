package miner

import (
	"sync"
	"time"
)

// SourceTime is a basic object which
// stores the time
type SourceTime struct {
	mutex sync.Mutex
	value time.Time
}

// Time
var (
	// sourceTime is the clock given by the packet capture
	sourceTime = NewSourceTime()
)

// NewSourceTime inis the internal source of time
func NewSourceTime() *SourceTime {
	return &SourceTime{mutex: sync.Mutex{}, value: time.Now()}
}

// Get returns the current source time
func (st *SourceTime) Get() time.Time {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	return st.value
}

// GetNano returns the current source time as
// nanoseconds
func (st *SourceTime) GetNano() int64 {
	st.mutex.Lock()
	defer st.mutex.Unlock()
	return st.value.UnixNano()
}

// Set defines the current source time
func (st *SourceTime) Set(t time.Time) {
	st.mutex.Lock()
	st.value = t
	st.mutex.Unlock()
}
