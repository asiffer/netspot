// stats.go

// Package stats provides statistics implementations.
package stats

import (
	"errors"
	"fmt"
	"netspot/config"

	"github.com/asiffer/gospot"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// StatConstructor is the generic template for a stat constructor
type StatConstructor func(bs BaseStat) StatInterface

var (
	logger zerolog.Logger // package logger
	// AvailableStats is the map linking a stat name to its constructor
	AvailableStats = make(map[string]StatConstructor)
)

func init() {

}

// Register aims to add implemented stats to the slice AvailableStats
func Register(name string, sc StatConstructor) error {
	_, exists := AvailableStats[name]
	if exists {
		msg := fmt.Sprintf("The statistics %s is already available", name)
		log.Error().Msg(msg)
		return errors.New(msg)
	}
	AvailableStats[name] = sc
	return nil
}

// BaseStat is the basic structure which defines a statistic. It
// embeds a string (its unique name) and a DSpot instance which monitors
// itself.
type BaseStat struct {
	name  string
	dspot *gospot.DSpot // the spot instance
}

// StatInterface gathers the common behavior of the statistics
type StatInterface interface {
	Name() string                       // the name of the statistics
	Requirement() []string              // the names of the requested counters
	Compute(ctrvalues []uint64) float64 // only compute the statistics
	Update(val float64) int             // feed dspot
	DSpot() *gospot.DSpot               // return the DSpot instance
	Configure() error                   // load the spot parameters
}

// DSpot returns the pointer to the DSpot instance embedded
// in the BaseStat
func (m *BaseStat) DSpot() *gospot.DSpot {
	return m.dspot
}

// Update feeds the DSpot instance embedded in the BaseStat
// with a new incoming value. It returns a return code
// according to normality/abnormality of the event.
func (m *BaseStat) Update(val float64) int {
	return m.dspot.Step(val)
}

// Configure loads the DSpot parameters
// from the config file. It is common for
// all the statistics
func (m *BaseStat) Configure() error {
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
	if sc.Depth, err = config.GetStrictlyPositiveInt(keys["depth"]); err != nil {
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
	bs := BaseStat{name: statname, dspot: nil}
	if err := bs.Configure(); err != nil {
		return nil, err
	}

	statConstructor, exists := AvailableStats[statname]
	if exists {
		return statConstructor(bs), nil
	}
	log.Error().Msg("Unknown statistics")
	return nil, errors.New("Unknown statistics")
}

func main() {
}
