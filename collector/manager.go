package collector

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type collectorState struct {
	loaded  atomic.Bool
	running atomic.Bool
}

type CollectorState struct {
	Loaded  bool `json:"loaded"`
	Running bool `json:"running"`
}

func (state *collectorState) State() CollectorState {
	return CollectorState{
		Loaded:  state.loaded.Load(),
		Running: state.running.Load(),
	}
}

// CollectorManager provides thread-safe state management
// above Collector
type CollectorManager struct {
	// managed collector
	Collector
	// inner state
	collectorState
	// stop utilities
	stop   chan bool
	waiter sync.WaitGroup
}

// Manage returns a managed collector
func Manage(c Collector) *CollectorManager {
	return &CollectorManager{
		Collector: c,
		stop:      make(chan bool),
	}
}

// Load triggers Collector.Load and store the state
// if success
func (m *CollectorManager) Load() error {
	if m.loaded.Load() {
		return fmt.Errorf("the collector is already loaded")
	}
	if err := m.Collector.Load(); err != nil {
		return err
	}
	m.loaded.Store(true)
	return nil
}

// Unload triggers Collector.Unload and store the state
// if success
func (m *CollectorManager) Unload() error {
	if m.running.Load() {
		return fmt.Errorf("could not unload running collector (stop first)")
	}
	if !m.loaded.Load() {
		return fmt.Errorf("the collector is not loaded")
	}
	if err := m.Collector.Unload(); err != nil {
		return err
	}
	m.loaded.Store(false)
	return nil
}

// Start triggers Collector.Start and store the state
// if success
func (m *CollectorManager) Start() error {
	if !m.loaded.Load() {
		return fmt.Errorf("the collector must be loaded first")
	}
	if m.running.Load() {
		return fmt.Errorf("the collector is already running")
	}

	m.running.Store(true)
	m.waiter.Add(1)
	go m.Collector.Start(m.stop)
	return nil
}

// Start sends a stop signal to the inner collector
func (m *CollectorManager) Stop() error {
	if !m.running.Load() {
		return fmt.Errorf("the collector is not running")
	}
	// send stop signal
	m.stop <- true
	// wait the end of the goroutine
	m.waiter.Wait()
	// update status
	m.running.Store(false)
	// no need to close channels
	return nil
}
