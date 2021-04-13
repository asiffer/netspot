package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/asiffer/netspot/analyzer"
	"github.com/asiffer/netspot/api"
	"github.com/asiffer/netspot/config"
	"github.com/asiffer/netspot/exporter"
	"github.com/asiffer/netspot/miner"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	cli "github.com/urfave/cli/v2"
)

// defaultConsoleWidth is the fallback width of the
// current terminal
const defaultConsoleWidth = 80

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

func byteToInt(array []byte) (int, int) {
	size := len(array)
	out := 0
	i := 0
	for ; (i < size) && (array[i] >= 48) && (array[i] <= 57); i++ {
		out = 10*out + (int(array[i]) - 48)
	}
	return out, i + 1
}

func consoleWidth() int {
	command := exec.Command("stty", "size")
	if command == nil {
		return defaultConsoleWidth
	}
	command.Stdin = os.Stdin
	out, err := command.Output()
	if err != nil || len(out) == 0 {
		return defaultConsoleWidth
	}
	// height
	_, index := byteToInt(out)
	// width
	width, _ := byteToInt(out[index:])
	return width
}

//------------------------------------------------------------------------------
// INTERNAL FUNCTIONS
//------------------------------------------------------------------------------

// initLoggers initializes all the subloggers. It must
// be done after the initialization of the general logger
func initLoggers() {
	analyzer.InitLogger()
	miner.InitLogger()
	exporter.InitLogger()
	config.InitLogger()
	api.InitLogger()
}

// InitConsoleWriter initializes the console outputing details about the
// netspot events.
func initConsoleWriter() {
	width := consoleWidth()
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
		for k := 55; k < width; k += 10 {
			format := "%-" + fmt.Sprintf("%d", k) + "s"
			if size < k {
				return fmt.Sprintf(format, s)
			}
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
