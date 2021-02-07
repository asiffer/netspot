// stats.go

// Package stats provides statistics implementations.
package stats

import (
	"fmt"
	"netspot/config"
	"sync"

	"github.com/asiffer/gospot"
	"github.com/rs/zerolog"
)

// StatConstructor is the generic template for a stat constructor
type StatConstructor func(bs BaseStat) StatInterface

var (
	logger zerolog.Logger // package logger
	// AvailableStats is the map linking a stat name to its constructor
	AvailableStats = make(map[string]StatInterface)
)

func init() {

}

// Register aims to add implemented stats to the slice AvailableStats
func Register(s StatInterface) error {
	if _, exists := AvailableStats[s.Name()]; exists {
		return fmt.Errorf("The statistics %s is already available", s.Name())
	}
	AvailableStats[s.Name()] = s
	return nil
}

// BaseStat is the basic structure which defines a statistic. It
// embeds a string (its unique name) and a DSpot instance which monitors
// itself.
type BaseStat struct {
	name  string
	dspot *gospot.DSpot // the spot instance
	mutex sync.Mutex    // mutex for thread safety
}

// StatInterface gathers the common behavior of the statistics
type StatInterface interface {
	Name() string // the name of the statistics
	Configure() error
	Requirement() []string              // the names of the requested counters
	Compute(ctrvalues []uint64) float64 // only compute the statistics
	Update(val float64) int             // feed DSpot
	UpProbability(quantile float64) float64
	DownProbability(quantile float64) float64
	GetThresholds() (float64, float64)
	Status() gospot.DSpotStatus
}

// Update feeds the DSpot instance embedded in the BaseStat
// with a new incoming value. It returns a return code
// according to normality/abnormality of the event.
func (m *BaseStat) Update(val float64) int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.dspot.Step(val)
}

// UpProbability computes the probability to get
// a value higher than the given threshold
func (m *BaseStat) UpProbability(q float64) float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.dspot.UpProbability(q)
}

// DownProbability computes the probability to get
// a value lower than the given threshold
func (m *BaseStat) DownProbability(q float64) float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.dspot.DownProbability(q)
}

// Status returns the status of the DSpot instance
// embedded in the stat
func (m *BaseStat) Status() gospot.DSpotStatus {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.dspot.Status()
}

// GetThresholds returns upper and lower decision thresholds
func (m *BaseStat) GetThresholds() (float64, float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.dspot.GetLowerThreshold(), m.dspot.GetUpperThreshold()
}

// Configure loads the DSpot parameters
// from the config file. It is common for
// all the statistics
func (m *BaseStat) Configure() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var err error
	sc := gospot.DSpotConfig{}
	prefix := "spot." + m.name
	keys := make(map[string]string)
	parameters := []string{
		"depth",
		"q",
		"n_init",
		"level",
		"up",
		"down",
		"alert",
		"bounded",
		"max_excess",
	}

	// get the right key
	for _, op := range parameters {
		keys[op] = prefix + "." + op
		if !config.HasKey(keys[op]) {
			// change key (fallback)
			keys[op] = "spot." + op
		}
	}

	if sc.Q, err = config.GetStrictlyPositiveFloat64(keys["q"]); err != nil {
		return err
	}
	if sc.Level, err = config.GetStrictlyPositiveFloat64(keys["level"]); err != nil {
		return err
	}

	if sc.Ninit, err = config.GetStrictlyPositiveInt(keys["n_init"]); err != nil {
		return err
	}
	if sc.Depth, err = config.GetInt(keys["depth"]); err != nil {
		return err
	}
	if sc.MaxExcess, err = config.GetStrictlyPositiveInt(keys["max_excess"]); err != nil {
		return err
	}

	if sc.Up, err = config.GetBool(keys["up"]); err != nil {
		return err
	}
	if sc.Down, err = config.GetBool(keys["down"]); err != nil {
		return err
	}
	if sc.Alert, err = config.GetBool(keys["alert"]); err != nil {
		return err
	}
	if sc.Bounded, err = config.GetBool(keys["bounded"]); err != nil {
		return err
	}

	m.dspot = gospot.NewDSpotFromConfig(&sc)
	return nil
}

// StatFromName returns the StatInterface related to the
// given name. It returns an error when the desired statistic does
// not exist.
func StatFromName(statname string) (StatInterface, error) {
	if stat, exists := AvailableStats[statname]; exists {
		if err := stat.Configure(); err != nil {
			return nil, fmt.Errorf("Error while configuring %s: %v",
				stat.Name(), err)
		}
		return stat, nil
	}
	return nil, fmt.Errorf("Unknown stat %s", statname)
}

func main() {}
