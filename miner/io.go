// io.go

package miner

import (
	"errors"
	"fmt"
	"netspot/miner/counters"

	"github.com/rs/zerolog/log"
)

//----------------------------------------------------------------------------//
//--------------------------- UNEXPORTED FUNCTIONS ---------------------------//
//----------------------------------------------------------------------------//

// counterFromName returns the BaseCtrInterface related to the
// given name. It returns an error when the desired counter does
// not exist.
func counterFromName(name string) counters.BaseCtrInterface {
	cc, exists := counters.AvailableCounters[name]
	if exists {
		return cc()
	}
	log.Error().Msg("Unknown counter")
	return nil
}

func load(ctr counters.BaseCtrInterface) (int, error) {
	if ctr != nil {
		if isAlreadyLoaded(ctr.Name()) {
			msg := fmt.Sprintf("Counter %s already loaded", ctr.Name())
			log.Debug().Msgf("Counter %s already loaded", ctr.Name())
			return -2, errors.New(msg)
		}
		counterID = counterID + 1
		counterMap[counterID] = ctr
		log.Debug().Msgf("Loading counter %s", ctr.Name())
		return counterID, nil

	}
	log.Error().Msg("Cannot load null counter")
	return -1, errors.New("Cannot load null counter")
}

//----------------------------------------------------------------------------//
//---------------------------- EXPORTED FUNCTIONS ----------------------------//
//----------------------------------------------------------------------------//

// GetNumberOfLoadedCounters returns the current number of
// counters that are loaded
func GetNumberOfLoadedCounters() int {
	return len(counterMap)
}

// Unload removes the counter identified by its id
func Unload(id int) {
	log.Debug().Msgf("Unloading counter %s", counterMap[id].Name())
	delete(counterMap, id)
}

// UnloadAll remove all the loaded counters
func UnloadAll() {
	for k := range counterMap {
		Unload(k)
	}
	// reset the counter
	counterID = 0
}

// UnloadFromName unloads a counter and return 0 if the operation
// has correctly been done. Returns -1 if the counter does not exist
func UnloadFromName(ctrname string) int {
	id := idFromName(ctrname)
	if id == -1 {
		return -1
	}
	Unload(id)
	log.Debug().Msgf("Unloading %s", ctrname)
	return 0

}

// LoadFromName loads a new counter and returns its id
func LoadFromName(ctrname string) int {
	ctr := counterFromName(ctrname)
	id, _ := load(ctr)
	if sniffing && id >= 0 {
		mux.Lock()
		startCounter(id)
		mux.Unlock()
	}
	return id
}

// GetLoadedCounters returns a slice of the names of
// the loaded counters
func GetLoadedCounters() []string {
	nbCounters := len(counterMap)
	names := make([]string, 0, nbCounters)
	for _, ctr := range counterMap {
		names = append(names, ctr.Name())
	}
	return names
}

// Reset sets the counter value to zero (the "zero" value of the counter)
func Reset(id int) int {
	ctr, exists := counterMap[id]
	if exists {
		if ctr.IsRunning() {
			ctr.SigPipe() <- 2
		} else {
			ctr.Reset()
		}
		return 0
	}
	return -1
}

// ResetAll sets all the counters to zero
func ResetAll() {
	mux.Lock()
	for _, ctr := range counterMap {
		if ctr.IsRunning() {
			ctr.SigPipe() <- 2
		} else {
			ctr.Reset()
		}

	}
	mux.Unlock()

}

//----------------------------------------------------------------------------//
//--------------------------------- PATTERNS ---------------------------------//
//----------------------------------------------------------------------------//

// LoadPattern tries to load a counter on a given pattern
// the pattern has the form 'ip/mask:port->ip/mask:port'
// The symbols '->' and ':' are mandatory. If mask is not given
// the maximum mask length is set (ex: 32 for IPv4). If a port or
// an address is not given, it is replaced by 0 (address 0.0.0.0
// or port 0)
func LoadPattern(p string, tag string) (int, error) {
	pattern, err := counters.ParsePattern(p)
	if err != nil {
		return -1, err
	}
	ctr := counters.NewPatternCtr(pattern, tag)
	id, err := load(ctr)
	if sniffing && id >= 0 {
		mux.Lock()
		startCounter(id)
		mux.Unlock()
	}
	return id, err
}
