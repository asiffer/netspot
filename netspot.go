// netspot.go

// Netspot is a basic IDS with statistical learning. It works as a server
// which either listens on interface or reads a network capture file. The server
// is controlled by a client `netspotctl`.
package main

import (
	"fmt"
	"netspot/analyzer"
	"netspot/api"
	"netspot/miner"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

// ServerConfig is a basic structure which stores
// the configuration of the server
type ServerConfig struct {
	// LogLevel defines the logging level. Possible values are:
	// - panic (zerolog.PanicLevel, 5)
	// - fatal (zerolog.FatalLevel, 4)
	// - error (zerolog.ErrorLevel, 3)
	// - warn (zerolog.WarnLevel, 2)
	// - info (zerolog.InfoLevel, 1)
	// - debug (zerolog.DebugLevel, 0)
	LogLevel int
	// HTTP activates the HTTP REST endpoint
	HTTP bool
	// HTTPAddress defines the ip address and tcp port
	// of the HTTP endpoint
	HTTPAddress string
	// TLS activates HTTPS on HTTP endpoint
	TLS bool
	// CertFile is the server public certificate
	CertFile string
	// KeyFile is the server private key
	KeyFile string
	// RPC activates the Golang RPC server
	RPC bool
	// RPCAddress defines the ip address and tcp port
	// of the RPC endpoint
	RPCAddress string
}

var (
	app          *cli.App
	serverConfig ServerConfig
)

func init() {
	// default config
	viper.SetDefault("server.log_level", zerolog.InfoLevel)
	viper.SetDefault("server.http", true)
	viper.SetDefault("server.http_addr", "127.0.0.1:11000")
	viper.SetDefault("server.tls", false)
	viper.SetDefault("server.cert", "./cert.pem")
	viper.SetDefault("server.key", "./key.pem")
	viper.SetDefault("server.rpc", false)
	viper.SetDefault("server.rpc_addr", "127.0.0.1:11001")

	// init console
	InitConsoleWriter()
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

//------------------------------------------------------------------------------
// INTERNAL FUNCTIONS
//------------------------------------------------------------------------------

// InitConsoleWriter initializes the console outputing details about the
// netspot events.
func InitConsoleWriter() {
	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: zerolog.TimeFormatUnix}
	// output := zerolog.ConsoleWriter{Out: os.Stderr}
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
	log.Logger = log.Output(output)
	zerolog.TimeFieldFormat = time.StampNano

	// At the beginning the debug level is set
	zerolog.SetGlobalLevel(0)
	// initialize the sub-loggers
	analyzer.InitLogger()
	miner.InitLogger()
	api.InitLogger()
}

// LoadConfig set the global config file to read. It must be done before the subpackages
// are initialized.
func LoadConfig(file string) {
	if !fileExists(file) {
		log.Warn().Msgf("Config file %s does not exist. Falling back to default configuration", file)
	}
	viper.SetConfigFile(file)

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Info().Msgf("Config file changed: %s", e.Name)
	})
	viper.ReadInConfig()

	// server configuration
	log.Info().Msg("Config loaded")
}

// InitConfig set the server configuration from the config file
func InitConfig() {
	serverConfig = ServerConfig{
		LogLevel:    viper.GetInt("server.log_level"),
		HTTP:        viper.GetBool("server.http"),
		HTTPAddress: viper.GetString("server.http_addr"),
		TLS:         viper.GetBool("server.tls"),
		CertFile:    viper.GetString("server.cert"),
		KeyFile:     viper.GetString("server.key"),
		RPC:         viper.GetBool("server.rpc"),
		RPCAddress:  viper.GetString("server.rpc_addr"),
	}
}

// UpdateServerConfigFromCli override the options passed in the config file
// with the options passed in CLI
func UpdateServerConfigFromCli(c *cli.Context) {
	// logging level
	if c.IsSet("log-level") {
		serverConfig.LogLevel = c.Int("log-level")
	}

	// HTTP
	if c.IsSet("no-http") {
		serverConfig.HTTP = false
	}
	if c.IsSet("http") {
		serverConfig.HTTPAddress = c.String("http")
	}

	// RPC
	if c.IsSet("no-rpc") {
		serverConfig.RPC = false
	}
	if c.IsSet("rpc") {
		serverConfig.RPCAddress = c.String("rpc")
	}

	// TLS
	if c.IsSet("tls") {
		serverConfig.TLS = true
	}
	if c.IsSet("cert") {
		serverConfig.CertFile = c.String("cert")
	}
	if c.IsSet("key") {
		serverConfig.KeyFile = c.String("key")
	}

}

// InitSubpackages initialize the config of the miner and
// the analyzer.
func InitSubpackages() {
	miner.InitConfig()
	analyzer.InitConfig()
}

// StartServer (it receives the cli arguments)
func StartServer(c *cli.Context) {
	// load config
	if c.IsSet("config") {
		LoadConfig(c.String("config"))
	} else {
		LoadConfig("./netspot.toml")
	}

	// Initialize the server configuration
	InitConfig()

	// add passed cli arguments
	UpdateServerConfigFromCli(c)

	// set the logging level (cli override config file)
	zerolog.SetGlobalLevel(zerolog.Level(serverConfig.LogLevel))

	// Initialize the subpackages
	InitSubpackages()

	// com channel
	com := make(chan error)

	// if the flag -no-rpc has not been set AND
	// the config file does not activate RPC
	if serverConfig.RPC {
		go api.RunRPC(c.String("rpc"), com)
	}

	if serverConfig.HTTP {
		if serverConfig.TLS {
			// with TLS
			go api.RunHTTPS(c.String("http"),
				serverConfig.CertFile,
				serverConfig.KeyFile,
				com)
		} else {
			// without TLS
			go api.RunHTTP(c.String("http"), com)
		}

	}

	// wait
	if err := <-com; err != nil {
		log.Fatal().Msgf("server error: %v", err)
	}

}

// InitApp starts NetSpot
func InitApp() {
	app = cli.NewApp()
	app.Name = "NetSpot"
	app.Usage = "A basic IDS with statistical learning"
	app.Version = "1.3"

	// CLI arguments
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "/etc/netspot/netspot.toml",
			Usage: "Load configuration from `FILE`",
		},
		cli.StringFlag{
			Name:  "http",
			Value: "localhost:11000",
			Usage: "NetSpot server HTTP endpoint",
		},
		cli.StringFlag{
			Name:  "rpc",
			Value: "localhost:11001",
			Usage: "NetSpot server RPC endpoint",
		},
		cli.BoolFlag{
			Name:  "no-rpc",
			Usage: "Disable the golang RPC endpoint",
		},
		cli.BoolFlag{
			Name:  "no-http",
			Usage: "Disable the HTTP endpoint",
		},
		cli.BoolFlag{
			Name:  "tls",
			Usage: "Activate TLS on HTTP endpoint",
		},
		cli.StringFlag{
			Name:  "cert",
			Usage: "Path to the public certificate",
		},
		cli.StringFlag{
			Name:  "key",
			Usage: "Path to the private key",
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
