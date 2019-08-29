// config.go

package miner

import (
	"fmt"
	"netspot/miner/counters"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

//----------------------------------------------------------------------------//
//---------------------------- EXPORTED FUNCTIONS ----------------------------//
//----------------------------------------------------------------------------//

// InitConfig initializes the miner package from the config file
func InitConfig() {
	SetDevice(viper.GetString("miner.device"))
	SetSnapshotLen(viper.GetInt32("miner.snapshot_len"))
	SetPromiscuous(viper.GetBool("miner.promiscuous"))
	SetTimeout(viper.GetDuration("miner.timeout"))

	minerLogger.Debug().Msg(fmt.Sprint("Available counters: ", counters.GetAvailableCounters()))
	minerLogger.Info().Msg("Miner package configured")
}

// RawStatus returns the current status of the miner through a
// basic map. It is designed to a future print.
func RawStatus() map[string]string {
	return map[string]string{
		"promiscuous":     fmt.Sprintf("%v", promiscuous),
		"timeout":         fmt.Sprint(timeout),
		"device":          device,
		"snapshot_length": fmt.Sprintf("%d", snapshotLen),
	}
}

// GenericStatus returns the current status of the analyzer through a
// basic map. It is designed to JSON marshalling.
func GenericStatus() map[string]interface{} {
	return map[string]interface{}{
		"promiscuous":     promiscuous,
		"timeout":         timeout,
		"device":          device,
		"snapshot_length": snapshotLen,
	}
}

// DisableLogging sets the global zerolog log level to 0
func DisableLogging() {
	minerLogger.Warn().Msg("Disabling logging")
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

// SetLogging sets the global zerolog log level
func SetLogging(level int) {
	l := zerolog.Level(level)
	zerolog.SetGlobalLevel(l)
	minerLogger.Warn().Msgf("Enabling logging (level %s)", l.String())
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
func SetPromiscuous(b bool) int {
	promiscuous = b
	minerLogger.Debug().Msgf("Promiscuous set to %v", b)
	return 0
}

// SetSnapshotLen sets the maximum size of packets which are captured
func SetSnapshotLen(sl int32) int {
	snapshotLen = sl
	minerLogger.Debug().Msgf("Snapshot length set to %d", sl)
	return 0
}

// SetTimeout set the timeout to the desired duration
func SetTimeout(d time.Duration) {
	timeout = d
	minerLogger.Debug().Msgf("Timeout set to %s", d)
}

// GetDevice returns the current device (interface name or capture file)
func GetDevice() string {
	return device
}

// SetDevice sets the device to listen. It can be either an interface or
// a capture file (ex: .pcap)
func SetDevice(dev string) int {
	if contains(AvailableDevices, dev) {
		device = dev
		iface = true
	} else {
		dev, err := filepath.Abs(dev)
		if fileExists(dev) && err == nil {
			device = dev
			iface = false
		} else {
			minerLogger.Error().Msgf("Unknown device (%s)", dev)
			return 1
		}
	}
	minerLogger.Info().Msgf(`Set device to "%s"`, dev)
	return 0
}

// // SetTickPeriod defines time between two ticks
// func SetTickPeriod(d time.Duration) {
// 	tickPeriod = d
// 	sendTicks = true
// }

// SetTickPeriod defines time between two ticks
// func SetTickPeriod(d time.Duration, c TimeChannel) {
// 	tickPeriod = d
// 	sendTicks = true
// 	remoteTimeChannel = c
// }
