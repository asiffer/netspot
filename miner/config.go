// config.go

package miner

import (
	"fmt"
	"netspot/config"
	"netspot/miner/counters"
	"path/filepath"
	"time"
)

//----------------------------------------------------------------------------//
//---------------------------- EXPORTED FUNCTIONS ----------------------------//
//----------------------------------------------------------------------------//

// InitConfig initializes the miner package from the config module
func InitConfig() error {
	if err := Zero(); err != nil {
		return err
	}

	key := "miner.device"
	s, err := config.GetString(key)
	if err != nil {
		minerLogger.Fatal().Msgf("Error while retrieving key %s: %v", key, err)
		return err
	}
	if err := SetDevice(s); err != nil {
		return err
	}

	key = "miner.snapshot_len"
	l, err := config.GetInt(key)
	if err != nil {
		minerLogger.Fatal().Msgf("Error while retrieving key %s: %v", key, err)
		return err
	}
	if err := SetSnapshotLen(int32(l)); err != nil {
		return err
	}

	key = "miner.promiscuous"
	p, err := config.GetBool(key)
	if err != nil {
		minerLogger.Fatal().Msgf("Error while retrieving key %s: %v", key, err)
		return err
	}
	if err := SetPromiscuous(p); err != nil {
		return err
	}

	key = "miner.timeout"
	t, err := config.GetDuration(key)
	if err != nil {
		minerLogger.Fatal().Msgf("Error while retrieving key %s: %v", key, err)
		return err
	}
	if err := SetTimeout(t); err != nil {
		return err
	}

	// log
	minerLogger.Debug().Msg(fmt.Sprint("Available counters: ", counters.GetAvailableCounters()))
	minerLogger.Info().Msg("Miner package configured")
	return nil
}

// GetNumberOfDevices returns the number of available devices (interfaces)
func GetNumberOfDevices() int {
	return len(GetAvailableDevices())
}

// IsDeviceInterface check if the current device is an interface
func IsDeviceInterface() bool {
	return iface
}

// IsPromiscuous returns the current status of the interface
// (not relevant for pcap file)
func IsPromiscuous() bool {
	return promiscuous
}

// SetPromiscuous set the promiscuous mode. If true, it means that the interface
// will receives packets  that are not intended for it.
func SetPromiscuous(b bool) error {
	promiscuous = b
	minerLogger.Debug().Msgf("Promiscuous set to %v", b)
	return nil
}

// SetSnapshotLen sets the maximum size of packets which are captured
func SetSnapshotLen(sl int32) error {
	snapshotLen = sl
	minerLogger.Debug().Msgf("Snapshot length set to %d", sl)
	return nil
}

// SetTimeout set the timeout to the desired duration
func SetTimeout(d time.Duration) error {
	timeout = d
	minerLogger.Debug().Msgf("Timeout set to %s", d)
	return nil
}

// GetDevice returns the current device (interface name or capture file)
func GetDevice() string {
	return device
}

// SetDevice sets the device to listen. It can be either an interface or
// a capture file (ex: .pcap)
func SetDevice(dev string) error {
	if contains(availableDevices, dev) {
		device = dev
		iface = true
	} else {
		dev, err := filepath.Abs(dev)
		if err == nil && fileExists(dev) {
			device = dev
			iface = false
		} else {
			err := fmt.Errorf("Unknown device %s", dev)
			minerLogger.Error().Msg(err.Error())
			return err
		}
	}
	minerLogger.Info().Msgf(`Set device to "%s"`, dev)
	return nil
}
