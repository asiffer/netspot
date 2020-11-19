// Package config provide configuration for all the other modules
// In particular it parses the config file and exposes
// parameters as key/value (go-ini)
package config

import (
	js "encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

// Konf is the global koanf instance.
// Use "." as the key path delimiter. This can be "/" or any character.
var (
	konf      = koanf.New(".")
	savedKonf *koanf.Koanf // fallback
)

var (
	// logger
	configLogger zerolog.Logger
)

var defaultConfig = map[string]interface{}{
	"api.endpoint":       "tcp://localhost:11000",
	"miner.device":       "any",
	"miner.promiscuous":  true,
	"miner.snapshot_len": 65535,
	"miner.timeout":      0,
	"analyzer.period":    1 * time.Second,
	"analyzer.stats":     []string{},
	"spot.depth":         50,
	"spot.q":             1e-4,
	"spot.n_init":        1000,
	"spot.level":         0.98,
	"spot.up":            true,
	"spot.down":          false,
	"spot.alert":         true,
	"spot.bounded":       true,
	"spot.max_excess":    200,
}

// networks accpeted by golang/net/Dial
var dialNetworks = []string{
	"tcp",
	"tcp4",
	"tcp6",
	"udp",
	"udp4",
	"udp6",
	"ip",
	"ip4",
	"ip6",
	"unix",
	"unixgram",
	"unixpacket",
}

var dataFormat = []string{"csv", "json", "gob"}

// check if the network is correct
func isValidDialNetwork(network string) bool {
	for _, n := range dialNetworks {
		if n == network {
			return true
		}
	}
	return false
}

func checkPath(p string) error {
	dir, err := filepath.Abs(filepath.Dir(p))
	if err != nil {
		return err
	}
	// check if dir exists
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		return err
	}
	return nil
}

func parseAddress(conn string) (string, string, error) {
	raw := strings.Split(conn, "://")
	if len(raw) != 2 {
		return "", "", fmt.Errorf("The address is not valid, its format must be proto://address")
	}
	// Known networks are "tcp", "tcp4" (IPv4-only), "tcp6" (IPv6-only),
	// "udp", "udp4" (IPv4-only), "udp6" (IPv6-only), "ip", "ip4" (IPv4-only),
	// "ip6" (IPv6-only), "unix", "unixgram" and "unixpacket".
	if !isValidDialNetwork(raw[0]) {
		return "", "", fmt.Errorf("The network %s is not valid (see https://golang.org/pkg/net/#Dial) to get valid ones", raw[0])
	}
	return raw[0], raw[1], nil
}

// output the right parser given the extension
func guessParser(file string) (koanf.Parser, error) {
	// lowercase
	ext := strings.ToLower(filepath.Ext(file))
	switch ext {
	case ".yaml":
		return yaml.Parser(), nil
	case ".json":
		return json.Parser(), nil
	case ".toml":
		return toml.Parser(), nil
	default:
		return nil, fmt.Errorf("Extension '%s' not supported", ext)
	}
}

// InitLogger initialize the sublogger for EXPORTER
func InitLogger() {
	configLogger = log.With().Str("module", "CONFIG").Logger()
}

// InitConfig load the default config
func InitConfig() error {
	return konf.Load(confmap.Provider(defaultConfig, "."), nil)
}

// JSON return the current config
func JSON() ([]byte, error) {
	return js.Marshal(konf.Raw())
}

// Debug prints all the config
func Debug() {
	konf.Print()
}

// RegisterDefaultConfig add default parameters
// in the default config. This function should be called
// in a init() function (before other config load)
func RegisterDefaultConfig(m map[string]interface{}) {
	configLogger.Debug().Msgf("Setting default config: %v", m)
	for k, v := range m {
		defaultConfig[k] = v
	}
}

// ========================================================================== //
// API (functions to access the configuration)
// ========================================================================== //

// HasKey checks if a key is given in the config file
func HasKey(key string) bool {
	// the konf.Keys() list the 'end' keys
	// for instance 'miner' is not a valid key
	// but 'miner.device' is.
	for _, k := range konf.Keys() {
		if key == k {
			return true
		}
	}
	return false
}

// HasNotNilKey checks if a key is given in the config file
// anf if its value is not nil
func HasNotNilKey(key string) bool {
	return konf.Exists(key) && konf.Get(key) != nil
}

// GetString returns a string key
func GetString(key string) (string, error) {
	if !HasKey(key) {
		return "", fmt.Errorf("Key %s does not exist", key)
	}
	s := konf.String(key)
	if s == "" {
		return "", fmt.Errorf("Error while parsing key %s", key)
	}
	return s, nil
}

// GetStringList returns a slice of string
func GetStringList(key string) ([]string, error) {
	if !HasKey(key) {
		return nil, fmt.Errorf("Key %s does not exist", key)
	}
	s := konf.Strings(key)
	// if len(s) == 0 {
	// 	return nil, fmt.Errorf("Error while parsing key %s", key)
	// }
	return s, nil
}

// GetPath return a valid path
func GetPath(key string) (string, error) {
	if !HasKey(key) {
		return "", fmt.Errorf("Key %s does not exist", key)
	}
	p := konf.String(key)
	if err := checkPath(p); err != nil {
		return "", err
	}
	return p, nil
}

// GetSocket returns a valid socket: proto, address
func GetSocket(key string) (string, string, error) {
	if !HasKey(key) {
		return "", "", fmt.Errorf("Key %s does not exist", key)
	}
	return parseAddress(konf.String(key))
}

// GetDataFormat returns a valid data format (for socket only)
func GetDataFormat(key string) (string, error) {
	if !HasKey(key) {
		return "", fmt.Errorf("Key %s does not exist", key)
	}
	k := konf.String(key)
	for _, f := range dataFormat {
		if f == k {
			return f, nil
		}
	}
	return "", fmt.Errorf("The format %s is not accepted (only csv, json and gob)", k)
}

// GetInt returns a int key
func GetInt(key string) (int, error) {
	if !HasKey(key) {
		return 0, fmt.Errorf("Key %s does not exist", key)
	}
	i := konf.Int(key)
	if i == 0 {
		return 0, fmt.Errorf("Error while parsing key %s (got 0)", key)
	}
	return i, nil
}

// GetStrictlyPositiveInt returns a int key > 0
func GetStrictlyPositiveInt(key string) (int, error) {
	if !HasKey(key) {
		return 0, fmt.Errorf("Key %s does not exist", key)
	}
	i := konf.Int(key)
	if i <= 0 {
		return 0, fmt.Errorf("Error while parsing key %s (got %d)", key, i)
	}
	return i, nil
}

// GetFloat64 returns a float64 key
func GetFloat64(key string) (float64, error) {
	if !HasKey(key) {
		return 0., fmt.Errorf("Key %s does not exist", key)
	}
	f := konf.Float64(key)
	if f == 0.0 {
		return 0.0, fmt.Errorf("Error while parsing key %s (got 0.0)", key)
	}
	return f, nil
}

// GetStrictlyPositiveFloat64 returns only strictly positive float64 key
func GetStrictlyPositiveFloat64(key string) (float64, error) {
	if !HasKey(key) {
		return 0., fmt.Errorf("Key %s does not exist", key)
	}
	f := konf.Float64(key)
	if f <= 0. {
		return 0., fmt.Errorf("Error while parsing key %s (got %f)", key, f)
	}
	return f, nil
}

// GetBool returns a boolean key
func GetBool(key string) (bool, error) {
	if !HasKey(key) {
		return false, fmt.Errorf("Key %s does not exist", key)
	}
	return konf.Bool(key), nil
}

// MustBool return true if the key exists with a true-like value
// otherwise it returns false
func MustBool(key string) bool {
	if HasKey(key) {
		return konf.Bool(key)
	}
	return false
}

// GetDuration returns a duration key
func GetDuration(key string) (time.Duration, error) {
	if !HasKey(key) {
		return 0, fmt.Errorf("Key %s does not exist", key)
	}
	return konf.Duration(key), nil
}

// SetValue defines a value of a parameter
func SetValue(key string, value interface{}) error {
	m := map[string]interface{}{key: value}
	return konf.Load(confmap.Provider(m, "."), nil)
}

// ========================================================================== //
// Loaders
// ========================================================================== //

// LoadDefaults basically loads the `defaultConfig` map
// which provides default values
func LoadDefaults() error {
	if err := konf.Load(confmap.Provider(defaultConfig, "."), nil); err != nil {
		return fmt.Errorf("Error while loading default conf: %v", err)
	}
	return nil
}

// LoadFromCli init the config based on the command line arguments
func LoadFromCli(c *cli.Context) error {
	// load defaults
	configLogger.Debug().Msgf("Loading default configuration")
	if err := LoadDefaults(); err != nil {
		return err
	}
	// check if config file is given
	p := c.Path("config")
	if len(p) > 0 {
		// load config file arguments
		if err := LoadConfig(p); err != nil {
			return err
		}
	}
	// now load the cli arguments (override the config file)
	configLogger.Debug().Msgf("Loading cli parameters")
	// configLogger.Debug().Msgf("Before CLI:\n %+v", konf.All())
	if err := konf.Load(Provider(c, "."), nil); err != nil {
		return fmt.Errorf("Error loading config from cli: %v", err)
	}

	configLogger.Debug().Msgf("Final config:\n %+v", konf.All())
	configLogger.Info().Msgf("Configuration loaded")
	return nil
}

// LoadConfig inits the config package from the given file
func LoadConfig(filename string) error {
	configLogger.Debug().Msgf("Loading config file %s", filename)
	// Load JSON config.
	parser, err := guessParser(filename)
	if err != nil {
		return err
	}
	// load config (it overrides default config)
	if err := konf.Load(file.Provider(filename), parser); err != nil {
		return fmt.Errorf("Error loading config file %s: %v", filename, err)
	}
	return nil
}

// LoadForTest loads a basic config given in input
func LoadForTest(m map[string]interface{}) error {
	return konf.Load(confmap.Provider(m, "."), nil)
}

// LoadForTestToml loads a basic config file
func LoadForTestToml(toml string) error {
	return LoadConfig(toml)
}

// LoadForTestRawToml loads from raw bytes
func LoadForTestRawToml(raw []byte) error {
	return konf.Load(rawbytes.Provider(raw), toml.Parser())
}

// LoadFromRawJSON loads from raw bytes
func LoadFromRawJSON(raw []byte) error {
	data, err := json.Parser().Unmarshal(raw)
	if err != nil {
		return err
	}
	// check keys
	for key := range data {
		if !HasKey(key) {
			return fmt.Errorf("The key '%s' is unknown", key)
		}
	}
	return konf.Load(confmap.Provider(data, "."), nil)
}

// Clean remove all config keys
func Clean() {
	konf = koanf.New(".")
}

// Save/Restore ============================================================= //
// ========================================================================== //
// ========================================================================== //

// Save copy the configuration
func Save() {
	savedKonf = konf.Copy()
	configLogger.Debug().Msg("Configuration saved")
}

// Fallback keeps the last saved configurtion and
// restores it as the current config
func Fallback() error {
	if savedKonf == nil {
		return fmt.Errorf("No configuration to restore")
	}
	konf = savedKonf.Copy()
	savedKonf = nil
	configLogger.Debug().Msg("Configuration restored")
	return nil
}

func main() {}
