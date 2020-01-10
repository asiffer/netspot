// exporter.go
//

// Package exporter aims to provide a general framework to send
// netspot data to different endpoints (console, files, db, sockets etc.)
package exporter

import (
	"fmt"
	"time"
)

var (
	// loaded stores all the loaded Hauliers
	loaded = make([]Haulier, 0)
	// available
	available = map[string]Haulier{
		"console": &Console{},
		"file":    &File{},
	}
)

func isAvailable(s string) bool {
	for name := range available {
		if name == s {
			return true
		}
	}
	return false
}

func isLoaded(s string) bool {
	for _, e := range loaded {
		if e.Name() == s {
			return true
		}
	}
	return false
}

// SpotAlert is a simple structure to log alerts sent
// by spot instances
type SpotAlert struct {
	Status      string
	Stat        string
	Value       float64
	Code        int
	Probability float64
}

// Haulier is the general interface which denotes
// a module which sends data from netspot
type Haulier interface {
	Name() string
	Init(...interface{}) error
	Write(time.Time, map[string]float64) error
	Warn(time.Time, *SpotAlert) error
	Close() error
}

// Load init loads a new Haulier
func Load(name string, options ...interface{}) error {
	if isLoaded(name) {
		return fmt.Errorf("The '%s' Haulier is already loaded", name)
	}
	exp := available[name]
	if err := exp.Init(options); err != nil {
		return fmt.Errorf("Error while loading %s (%w)", name, err)
	}
	loaded = append(loaded, exp)
	return nil
}

// Clear removes all the loaded Haulier
func Clear() {
	loaded = make([]Haulier, 0)
}

// Write sends data to all the Haulier
func Write(t time.Time, data map[string]float64) error {
	for _, h := range loaded {
		err := h.Write(t, data)
		if err != nil {
			return fmt.Errorf("Error from %s: %w", h.Name(), err)
		}
	}
	return nil
}

// Warn sends alarm to the Hauliers
func Warn(t time.Time, s *SpotAlert) error {
	for _, h := range loaded {
		err := h.Warn(t, s)
		if err != nil {
			return fmt.Errorf("Error from %s: %w", h.Name(), err)
		}
	}
	return nil
}
