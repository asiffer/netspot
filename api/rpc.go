// rpc.go

package api

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"netspot/analyzer"
	"netspot/miner"
	"time"

	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
)

// Netspot is the object to build the API
type Netspot struct{}

//------------------------------------------------------------------------------
// HTTP API
//------------------------------------------------------------------------------

// Zero resets the analyzer and the miner. It returns
// - 0 if everything is ok
// - 1 when the miner did not reset
// - 2 when the analyzer did not reset
func (ns *Netspot) Zero(none *int, i *int) error {
	err := analyzer.Zero()
	if err != nil {
		*i = 2
		return err
	}
	//
	// The analyzer.Zero resets the Miner
	//
	// err = miner.Zero()
	// if err != nil {
	// 	*i = 1
	// 	return err
	// }
	*i = 0
	return nil
}

// SetDevice change the device to sniff (interface of pcap)
func (ns *Netspot) SetDevice(device string, i *int) error {
	*i = miner.SetDevice(device)
	if *i == 1 {
		return fmt.Errorf("Unknown device (%s)", device)
	}
	return nil
}

// SetPromiscuous changes the promiscuous mode (relevant to iface only)
func (ns *Netspot) SetPromiscuous(b bool, i *int) error {
	if miner.IsPromiscuous() == b {
		*i = -1
		if b {
			return errors.New("Promiscuous mode already activated")
		}
		return errors.New("Promiscuous mode already desactivated")

	}
	*i = miner.SetPromiscuous(b)
	if *i != 0 {
		return errors.New("Unhandled error")
	}
	return nil
}

// SetPeriod changes period of stat computation
func (ns *Netspot) SetPeriod(duration string, i *int) error {
	d, e := time.ParseDuration(duration)
	if e != nil {
		*i = -1
		return e
	}
	*i = 0
	analyzer.SetPeriod(d)
	return e
}

// SetOutputDir changes the directory of the netspot output
func (ns *Netspot) SetOutputDir(dir string, i *int) error {
	if err := analyzer.SetOutputDir(dir); err != nil {
		*i = 1
		return err
	}
	*i = 0
	return nil
}

// SetFileLogging (des)activate the data/thresholds/anomaly logging
// into files (files are saved in the outputDir directory)
func (ns *Netspot) SetFileLogging(save bool, i *int) error {
	if err := analyzer.SetFileLogging(save); err != nil {
		*i = 1
		return err
	}
	*i = 0
	return nil
}

// SetInfluxDBLogging (des)activate the data/thresholds logging
// into influxdb (! anomalies are not logged into influxdb !)
func (ns *Netspot) SetInfluxDBLogging(save bool, i *int) error {
	if err := analyzer.SetInfluxDBLogging(save); err != nil {
		*i = 1
		return err
	}
	*i = 0
	return nil
}

// AvailableInterface returns a slice of the interfaces which can be sniffed
func (ns *Netspot) AvailableInterface(none *int, deviceList *[]string) error {
	for _, s := range miner.GetAvailableDevices() {
		*deviceList = append(*deviceList, s)
	}
	return nil
}

// Load loads a stat from the given name. It returns the id of the stat (it may be useless).
func (ns *Netspot) Load(statName string, i *int) error {
	id, err := analyzer.LoadFromName(statName)
	*i = id
	return err
}

// Alive returns true. If you can call this function, it means that the
// server is running.
func (ns *Netspot) Alive(none *int, b *bool) error {
	*b = true
	return nil
}

// ListLoaded returns a slice of the statistics which are curently loaded
func (ns *Netspot) ListLoaded(none *int, statList *[]string) error {
	for _, s := range analyzer.GetLoadedStats() {
		*statList = append(*statList, s)
	}
	return nil
}

// ListAvailable returns a slice of the statistics which can be loaded (already
// loaded statistics are also present in this list)
func (ns *Netspot) ListAvailable(none *int, statList *[]string) error {
	for _, s := range analyzer.GetAvailableStats() {
		*statList = append(*statList, s)
	}
	return nil
}

// StatStatus returns a raw status of the DSpot instance monitoring the given
// statistic.
func (ns *Netspot) StatStatus(statName string, rawstatus *string) error {
	status, err := analyzer.StatStatus(statName)
	if err != nil {
		return err
	}
	*rawstatus = status.String()
	return nil

}

// StatValues return a current snapshot of the stat values (and their thresholds)
func (ns *Netspot) StatValues(none *int, values *map[string]float64) error {
	val := analyzer.StatValues()
	for s, v := range val {
		(*values)[s] = v
	}
	return nil

}

// Unload removes a loaded statistics. See analyzer.UnloadFromName to get
// the detail of the return values.
func (ns *Netspot) Unload(statName string, i *int) error {
	id, err := analyzer.UnloadFromName(statName)
	*i = id
	return err
}

// UnloadAll removes a loaded statistics. See analyzer.UnloadFromName to get
// the detail of the return values.
func (ns *Netspot) UnloadAll(none string, i *int) error {
	analyzer.UnloadAll()
	if analyzer.GetNumberOfLoadedStats() != 0 {
		*i = -1
		return errors.New("Statistics remain")
	}
	*i = 0
	return nil
}

// Config returns the configurations of the miner and the analyzer.
func (ns *Netspot) Config(none *int, s *string) error {
	bold := color.New(color.FgWhite, color.Bold)
	format := "%20s   %s\n"
	confAnalyzer := analyzer.RawStatus()
	confMiner := miner.RawStatus()

	*s += bold.Sprint("Miner\n")
	for k, v := range confMiner {
		*s += fmt.Sprintf(format, k, v)
	}

	*s += "\n" + bold.Sprint("Analyzer\n")
	for k, v := range confAnalyzer {
		*s += fmt.Sprintf(format, k, v)
	}

	return nil
}

// Start runs the miner and then the stats
func (ns *Netspot) Start(none *int, i *int) error {
	if analyzer.IsRunning() {
		*i = 3
		return errors.New("The statistics are currently computed")
	}

	if miner.IsSniffing() {
		*i = 2
		return errors.New("The sniffer is already running")
	}

	// start the analyzer (it also starts the miner)
	analyzer.StartStats()
	*i = 0
	return nil
}

// Stop stops the stat computation (and the miner too)
func (ns *Netspot) Stop(none *int, i *int) error {
	if !analyzer.IsRunning() {
		*i = 1
		return errors.New("The statistics are not currently monitored")
	}
	// stop the analyzer. It also stops the miner.
	analyzer.StopStats()
	*i = 0
	return nil
}

// RunRPC starts the golang RPC server
func RunRPC(addr string, com chan error) {
	// Register the interface
	ns := new(Netspot)
	rpc.Register(ns)
	rpc.HandleHTTP()

	// Start listening
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		// log.Fatal().Msgf("listen error: %v", err)
		com <- err
	}
	log.Info().Msgf("RPC listening on %s", addr)

	// Serve
	err = http.Serve(listener, nil)
	if err != nil {
		log.Fatal().Msgf("server error: %v", err)
		com <- err
	}
}
