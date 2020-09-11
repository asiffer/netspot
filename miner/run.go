// run.go

package miner

import (
	"fmt"
	"time"
)

// API ====================================================================== //
// ========================================================================== //
// ========================================================================== //

// IsSniffing returns the sniffing status
func IsSniffing() bool {
	return sniffing
}

// GetSourceTime returns the time given by the current packet source
func GetSourceTime() int64 {
	return SourceTime.UnixNano()
}

// Start starts to sniff the current device. It does nothing
// if the sniffing is in progress. This is a goroutine, so it returns
// once the sniffing has started.
func Start() (DataChannel, error) {
	if sniffing {
		return nil, fmt.Errorf("Already sniffing")
	}
	if len(dispatcher.loadedCounters()) == 0 {
		return nil, fmt.Errorf("No counters loaded")
	}
	data := make(DataChannel)
	minerLogger.Info().Msgf("Start sniffing %s", device)
	minerLogger.Debug().Msgf("Loaded counters: %v", dispatcher.loadedCounters())
	go sniff(internalEventChannel, data)
	for !sniffing {
		// wait for sniffing
	}
	return data, nil
}

// StartAndYield starts the miner and demands it to send
// counter values at given period
func StartAndYield(period time.Duration) (DataChannel, error) {
	if sniffing {
		return nil, fmt.Errorf("Already sniffing")
	}
	if len(dispatcher.loadedCounters()) == 0 {
		return nil, fmt.Errorf("No counters loaded")
	}
	data := make(DataChannel)
	minerLogger.Info().Msgf("Start sniffing %s", device)
	minerLogger.Debug().Msgf("Loaded counters: %v", dispatcher.loadedCounters())
	go sniffAndYield(period, internalEventChannel, data)
	for !sniffing {
		// wait for sniffing
	}
	return data, nil
}

// Stop stops to sniff the device
func Stop() error {
	if !sniffing {
		return fmt.Errorf("The miner is already stopped")
	}
	minerLogger.Info().Msgf("Stopping counter")
	internalEventChannel <- STOP
	for sniffing {
		// wait for stop sniffing
	}
	return nil
}
