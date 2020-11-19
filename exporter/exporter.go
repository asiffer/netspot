// exporter.go
//

// Package exporter aims to provide a general framework to send
// netspot data to different endpoints (console, files, db, sockets etc.)
package exporter

import (
	"errors"
	"fmt"
	"netspot/config"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// loaded stores all the loaded ExportingModules
	loaded = make([]ExportingModule, 0)
	// available modules. It is filled at compile time
	available = make(map[string]ExportingModule)
	// logger
	exporterLogger zerolog.Logger
	// state
	started = false
)

// ExportingModule is the general interface which denotes
// a module which sends data from netspot
type ExportingModule interface {
	// Return the name of the module
	Name() string
	// Init the module (init some variables)
	Init() error
	// Start the module (make it ready to receive data)
	Start(series string) error
	// Aimed to write raw statistics
	Write(time.Time, map[string]float64) error
	// Aimed to write alerts
	Warn(time.Time, *SpotAlert) error
	// Close the module
	Close() error
}

func init() {

}

// reset the variables of the package
func reset() {
	// loaded stores all the loaded ExportingModules
	loaded = make([]ExportingModule, 0)
	// state
	started = false
}

// Main functions =========================================================== //
// ========================================================================== //
// ========================================================================== //

// InitLogger initialize the sublogger for EXPORTER
func InitLogger() {
	exporterLogger = log.With().Str("module", "EXPORTER").Logger()
}

// InitConfig load the modules according to the config file
// The modules must be given within the [exporter] section,
// like [exporter.file]
func InitConfig() error {
	// reset first
	if err := Zero(); err != nil {
		return err
	}
	// loop over the known exporting modules
	for _, module := range available {
		// it inits the module
		// the module manages whether to load itself or not
		if err := module.Init(); err != nil {
			return fmt.Errorf("Error while initializing '%s' module: %v",
				module.Name(), err)
		}
	}

	exporterLogger.Debug().Msgf("Loaded modules: %v", loadedModules())
	exporterLogger.Info().Msg("Exporter package configured")
	return nil
}

// Register makes a new module available
// It must be called in the init() function of the module
func Register(s ExportingModule) {
	available[s.Name()] = s
}

// RegisterAndSetDefaults makes a new module available
// Moreover it provides the default value of an exporter
// It must be called in the init() function of the module
func RegisterAndSetDefaults(s ExportingModule, m map[string]interface{}) {
	available[s.Name()] = s
	config.RegisterDefaultConfig(m)
}

// AvailableExportingModules returns the list of the available exporting modules
func AvailableExportingModules() []string {
	ah := make([]string, 0)
	for k := range available {
		ah = append(ah, k)
	}
	return ah
}

// HasStarted returns the internal state of the exporter
func HasStarted() bool {
	return started
}

// Load loads a new module (without initialization)
func Load(name string) error {
	if isLoaded(name) {
		// return nil
		return fmt.Errorf("The '%s' module is already loaded", name)
	}
	loaded = append(loaded, available[name])
	exporterLogger.Debug().Msgf("'%s' module loaded", name)
	return nil
}

// Unload removes a ExportingModule
func Unload(name string) error {
	i := findExportingModule(name)
	if i < 0 {
		return fmt.Errorf("The '%s' ExportingModule is not loaded", name)
	}

	// swap
	ls := len(loaded)
	loaded[ls-1], loaded[i] = loaded[i], loaded[ls-1]
	// remove last element
	loaded = loaded[:ls-1]
	return nil
}

// Start init all the connections from the module to their endpoint
// It triggers error when a connection fails.
func Start(series string) error {
	for _, module := range loaded {
		if err := module.Start(series); err != nil {
			return err
		}
	}
	started = true
	exporterLogger.Info().Msgf("Exporting modules are ready to receive data")
	return nil
}

// Write sends data to all the ExportingModule
func Write(t time.Time, data map[string]float64) error {
	for _, h := range loaded {
		err := h.Write(t, data)
		if err != nil {
			return fmt.Errorf("Error from %s: %v", h.Name(), err)
		}
	}
	return nil
}

// Warn sends alarm to the ExportingModules
func Warn(t time.Time, s *SpotAlert) error {
	for _, h := range loaded {
		err := h.Warn(t, s)
		if err != nil {
			return fmt.Errorf("Error from %s: %v", h.Name(), err)
		}
	}
	return nil
}

// Close does the job. It is like 'Stop' in the
// other packages
func Close() error {
	for _, h := range loaded {
		if err := h.Close(); err != nil {
			return err
		}
	}
	started = false
	return nil
}

// Zero wraps the reset function
func Zero() error {
	if HasStarted() {
		exporterLogger.Error().Msg("Cannot reload, exporter is active")
		return errors.New("Cannot reload, exporter is active")
	}
	reset()
	exporterLogger.Info().Msg("Exporter package (re)loaded")
	return nil
}

// Side functions =========================================================== //
// ========================================================================== //
// ========================================================================== //

func isAvailable(s string) bool {
	for name := range available {
		if name == s {
			return true
		}
	}
	return false
}

func isLoaded(s string) bool {
	return findExportingModule(s) >= 0
}

func untypeMap(m map[string]float64) map[string]interface{} {
	M := make(map[string]interface{})
	for key, value := range m {
		M[key] = value
	}
	return M
}

func loadedModules() []string {
	m := make([]string, len(loaded))
	for i, module := range loaded {
		m[i] = module.Name()
	}
	return m
}

func findExportingModule(name string) int {
	for i, e := range loaded {
		if e.Name() == name {
			return i
		}
	}
	return -1
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
