// netspot.go

// Netspot is a basic IDS with statistical learning. It works as a server
// which either listens on interface or reads a network capture file. The server
// is controlled by a client `netspotctl`.
package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"netspot/analyzer"
	"netspot/miner"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

var (
	app         *cli.App
	minerEvents chan int
)

// Netspot is the object to build the API
type Netspot struct{}

//------------------------------------------------------------------------------
// SIDE FUNCTIONS
//------------------------------------------------------------------------------

//fileExists returns whether the given file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

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
	err = miner.Zero()
	if err != nil {
		*i = 1
		return err
	}
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

// SetPromiscuous change the promiscuous mode (relevant to iface only)
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

// SetPeriod change period of stat computation
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

	// start the counters
	miner.StartSniffing()
	if !miner.IsSniffing() {
		*i = 1
		return errors.New("The sniffer is not well started")
	}

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
	analyzer.StopStats()
	miner.StopSniffing()
	*i = 0
	return nil
}

//------------------------------------------------------------------------------
// INTERNAL FUNCTIONS
//------------------------------------------------------------------------------

// InitConsoleWriter initializes the console outputing details about the
// netspot events.
func InitConsoleWriter() {
	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.StampMicro}

	output.FormatLevel = func(i interface{}) string {
		switch fmt.Sprintf("%s", i) {
		case "warn":
			return "\033[1;33mWARNING\033[0m"
		case "info":
			return "\033[1;32m   INFO\033[0m"
		case "fatal":
			return "\033[1;31m  FATAL\033[0m"
		case "error":
			return "\033[1;31m  ERROR\033[0m"
		case "debug":
			return "\033[0;37m  DEBUG\033[0m"
		case "panic":
			return "\033[1;31m  PANIC\033[0m"
		default:
			return fmt.Sprintf("%s", i)
		}
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		field := fmt.Sprintf("%s", i)
		switch field {
		case "type", "module":
			return ""
		default:
			return "\033[2m\033[37m" + field + ":" + "\033[0m"
		}
	}

	output.FormatFieldValue = func(i interface{}) string {
		switch i.(type) {
		case float64:
			f := i.(float64)
			if f < 1e-3 {
				return fmt.Sprintf("%e", f)
			}
			return fmt.Sprintf("%.5f", f)
		case int32, int16, int8, int:
			return fmt.Sprintf("%d", i)
		default:
			return strings.ToUpper(fmt.Sprintf("%s", i))
		}
	}

	output.PartsOrder = []string{"time", "level", "message"}
	log.Logger = log.Output(output)
	zerolog.TimeFieldFormat = time.StampNano
}

// InitConfig set the global config file to read. It must be done before the subpackages
// are initialized.
func InitConfig(file string) {
	if !fileExists(file) {
		log.Error().Msgf("Config file %s does not exist. Falling back to default configuration", file)
	}
	viper.SetConfigFile(file)

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Info().Msgf("Config file changed: %s", e.Name)
	})
	viper.ReadInConfig()
	log.Info().Msgf("Config file %s loaded", viper.ConfigFileUsed())
}

// InitSubpackages initialize the config of the miner and the analyzer.
func InitSubpackages() {
	miner.InitConfig()
	analyzer.InitConfig()
}

// StartServer (it receives the cli arguments)
func StartServer(c *cli.Context) {
	InitConfig(c.String("config"))
	zerolog.SetGlobalLevel(zerolog.Level(c.Int("log-level")))
	InitSubpackages()

	ns := new(Netspot)
	rpc.Register(ns)
	rpc.HandleHTTP()

	addr := fmt.Sprintf("%s:%d", c.String("address"), c.Int("port"))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal().Msgf("listen error: %s", err.Error())
	}
	log.Info().Msgf("Listening on %s", addr)
	http.Serve(listener, nil)
}

// InitApp starts NetSpot
func InitApp() {
	app = cli.NewApp()
	app.Name = "NetSpot"
	app.Usage = "A basic IDS with statistical learning"
	app.Version = "1.0"

	// CLI arguments
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "/etc/netspot/netspot.toml",
			Usage: "Load configuration from `FILE`",
		},
		cli.StringFlag{
			Name:  "address, a",
			Value: "localhost",
			Usage: "NetSpot server listening address",
		},
		cli.IntFlag{
			Name:  "port, p",
			Value: 11000,
			Usage: "NetSpot server listening port",
		},
		cli.IntFlag{
			Name:  "log-level, l",
			Value: 1,
			Usage: "Minimum logging level (0: Debug, 1: Info, 2: Warn, 3: Error)",
		},
	}

	// it calls StartServer to
	app.Action = StartServer
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
}

func main() {
	InitConsoleWriter()
	InitApp()
}
