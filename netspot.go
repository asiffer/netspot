// netspot.go

// Netspot is a basic IDS with statistical learning. It works as a server
// which either listens on interface or reads a network capture file. The server
// is controlled by a client `netspotctl`.
package main

import (
	"fmt"
	"netspot/analyzer"
	"netspot/api"
	"netspot/config"
	"netspot/exporter"
	"netspot/miner"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	cli "github.com/urfave/cli/v2"
)

// Version is the netspot version
var Version = "2.0a"

var (
	configFile string
	logLevel   int
)

var (
	desc = `
netspot is a simple IDS powered by statistical learning. 
It actually monitors network statistics and detect abnormal events. 
Its core mainly relies on the SPOT algorithm (https://asiffer.github.io/libspot/) 
which flags extreme events on high throughput streaming data.
`
)

var (
	commonFlags = []cli.Flag{
		&cli.PathFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "Load configuration from `FILE`",
			// Value:   "./netspot.toml",
		},
		&cli.IntFlag{
			Name:    "log-level",
			Aliases: []string{"l"},
			Usage:   "Level of debug (0 is the most verbose)",
			Value:   1, // INFO
		},
	}

	minerFlags = []cli.Flag{
		&cli.StringFlag{
			Name:    "miner.device",
			Aliases: []string{"d"},
			Usage:   "Sniff `DEVICE` (pcap or interface)",
			Value:   "any",
		},
		&cli.BoolFlag{
			Name:  "miner.promiscuous",
			Value: false,
			Usage: "Activate promiscuous mode",
		},
		&cli.IntFlag{
			Name:  "miner.snapshot_len",
			Value: 65535,
			Usage: "Amount of captured bytes for each packet",
		},
		&cli.DurationFlag{
			Name:  "miner.timeout",
			Value: 30 * time.Second,
			Usage: "Time to wait before stopping if no packets is received",
		},
	}

	apiFLags = []cli.Flag{
		&cli.StringFlag{
			Name:    "api.endpoint",
			Aliases: []string{"e"},
			Usage:   "Listen to `ENDPOINT` (format: proto://addr)",
			Value:   "tcp://localhost:11000",
		},
	}

	analyzerFlags = []cli.Flag{
		&cli.DurationFlag{
			Name:    "analyzer.period",
			Aliases: []string{"p"},
			Value:   1 * time.Second,
			Usage:   "Time between two stats computations",
		},
		&cli.StringSliceFlag{
			Name:    "analyzer.stats",
			Aliases: []string{"s"},
			Usage:   "List of statistics to monitor",
		},
	}

	exporterFlags = []cli.Flag{
		&cli.BoolFlag{
			Name:    "exporter.console.data",
			Aliases: []string{"v"},
			Value:   false,
			Usage:   "Display statistics on the console",
		},
		&cli.StringFlag{
			Name:    "exporter.file.data",
			Aliases: []string{"f"},
			Value:   "netspot_%s_data.json",
			Usage:   "Log statistics to the given file",
		},
	}
)

var (
	app = &cli.App{
		Name:                 "netspot",
		Usage:                "A simple IDS with statistical learning",
		Authors:              []*cli.Author{{Name: "asr"}},
		Version:              Version,
		Copyright:            "GPLv3",
		Description:          removeCharacters(desc, []string{"\n"}),
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:      "serve",
				Usage:     "Start netspot as a server",
				UsageText: "netspot serve [options]",
				Action:    RunServer,
				Flags:     concatFlags(commonFlags, minerFlags, analyzerFlags, apiFLags, exporterFlags),
			},
			{
				Name:   "run",
				Usage:  "Run netspot directly on the device",
				Action: RunCli,
				Flags:  concatFlags(commonFlags, minerFlags, analyzerFlags, exporterFlags),
			},
		},
	}
)

func concatFlags(flags ...[]cli.Flag) []cli.Flag {
	out := make([]cli.Flag, 0)
	for _, f := range flags {
		out = append(out, f...)
	}
	return out
}

// console init
func init() {
	initConsoleWriter()
	initLoggers()
}

// InitSubpackages initialize the config the
// netspot sub-packages
func initSubpackages() error {

	if err := miner.InitConfig(); err != nil {
		return fmt.Errorf("Error while initializing the Miner: %v", err)
	}

	if err := analyzer.InitConfig(); err != nil {
		return fmt.Errorf("Error while initializing the Analyzer: %v", err)
	}

	if err := exporter.InitConfig(); err != nil {
		return fmt.Errorf("Error while initializing the Exporter: %v", err)
	}

	if err := api.InitConfig(); err != nil {
		return fmt.Errorf("Error while initializing the API: %v", err)
	}

	return nil
}

//------------------------------------------------------------------------------
// SIDE FUNCTIONS
//------------------------------------------------------------------------------

//fileExists returns whether the given file exists
func fileExists(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	_, err = os.Stat(absPath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func removeCharacters(s string, char []string) string {
	for _, c := range char {
		s = strings.Replace(s, c, "", -1)
	}
	return s
}

//------------------------------------------------------------------------------
// INTERNAL FUNCTIONS
//------------------------------------------------------------------------------

// initLoggers initializes all the subloggers. It must
// be done after the initialization of the general logger
func initLoggers() {
	// manager.InitLogger()
	analyzer.InitLogger()
	miner.InitLogger()
	exporter.InitLogger()
	config.InitLogger()
	api.InitLogger()
}

// InitConsoleWriter initializes the console outputing details about the
// netspot events.
func initConsoleWriter() {
	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: zerolog.TimeFormatUnix}
	// output := zerolog.ConsoleWriter{Out: os.Stderr}
	output.FormatLevel = func(i interface{}) string {
		switch fmt.Sprintf("%s", i) {
		case "debug":
			return "\033[0;090m  DEBUG\033[0m"
		case "info":
			return "\033[0;092m   INFO\033[0m"
		case "warn":
			return "\033[0;093mWARNING\033[0m"
		case "error":
			return "\033[0;091m  ERROR\033[0m"
		case "fatal":
			return "\033[0;101m  FATAL\033[0m"
		case "panic":
			return "\033[0;101m  PANIC\033[0m"
		default:
			return fmt.Sprintf("%s", i)
		}
	}

	output.FormatMessage = func(i interface{}) string {
		if i == nil {
			return ""
		}
		s, ok := i.(string)
		if !ok {
			log.Debug().Msgf("Console format error with message: %v", i)
		}
		size := len(s)
		if size < 50 {
			return fmt.Sprintf("%-50s", i)
		}
		if size < 80 {
			return fmt.Sprintf("%-80s", i)
		}
		if size < 100 {
			return fmt.Sprintf("%-100s", i)
		}
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
			return "\033[1m" + strings.ToUpper(fmt.Sprintf("%s", i)) + "\033[0m"
		}
	}

	output.PartsOrder = []string{"time", "level", "caller", "message"}

	// set the logger
	log.Logger = log.Output(output)
	// time format
	zerolog.TimeFieldFormat = time.StampNano
	// zerolog.TimeFieldFormat = time.RFC3339Nano

}

// setLogging set the minimum level of the output logs.
// - panic (zerolog.PanicLevel, 5)
// - fatal (zerolog.FatalLevel, 4)
// - error (zerolog.ErrorLevel, 3)
// - warn (zerolog.WarnLevel, 2)
// - info (zerolog.InfoLevel, 1)
// - debug (zerolog.DebugLevel, 0)
func setLogging(level int) {
	l := zerolog.Level(level)
	zerolog.SetGlobalLevel(l)
}

// disableLogging disable the log output. Warning! It disables the log
// for all the modules using zerolog
func disableLogging() {
	// managerLogger.Info().Msg("Disabling logging")
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

func initConfig(c *cli.Context) error {
	// update logging level
	setLogging(c.Int("log-level"))

	// init the 'config' package only
	if err := config.InitConfig(); err != nil {
		return fmt.Errorf("Error while initializing the 'config' package: %v", err)
	}

	// load config
	if err := config.LoadFromCli(c); err != nil {
		return err
	}

	// init other packages
	return initSubpackages()
}

// RunServer is the entrypoint of the cli
func RunServer(c *cli.Context) error {
	if err := initConfig(c); err != nil {
		return err
	}
	// run server
	return api.Serve()
}

// RunCli starts directly the analyzer
func RunCli(c *cli.Context) error {
	if err := initConfig(c); err != nil {
		return err
	}
	// run cli
	return analyzer.Run()
}

func main() {
	// run cli (parse arguments)
	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().Msgf("%v", err)
	}

	config.Debug()
}
