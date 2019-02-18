// stats.go

// Package stats provides statistics implementations.
package stats

import (
	"errors"
	"fmt"
	"strings"

	"github.com/asiffer/gospot"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// StatConstructor is the generic template for a stat constructor
type StatConstructor func(bs BaseStat) StatInterface

var (
	logger zerolog.Logger // package logger
	// AvailableStats is the map linking a stat name to its constructor
	AvailableStats = make(map[string]StatConstructor)
)

func init() {
	// default values
	viper.SetDefault("dspot.depth", 50)
	viper.SetDefault("dspot.q", 1e-4)
	viper.SetDefault("dspot.n_init", 1000)
	viper.SetDefault("dspot.level", 0.98)
	viper.SetDefault("dspot.up", true)
	viper.SetDefault("dspot.down", false)
	viper.SetDefault("dspot.alert", true)
	viper.SetDefault("dspot.bounded", true)
	viper.SetDefault("dspot.max_excess", 200)
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
	Update(val float64) int32           // feed dspot
	DSpot() *gospot.DSpot               // return the DSpot instance
}

// DSpot returns the pointer to the DSpot instance embedded
// in the BaseStat
func (m *BaseStat) DSpot() *gospot.DSpot {
	return m.dspot
}

// Update feeds the DSpot instance embedded in the BaseStat
// with a new incoming value. It returns a return code
// according to normality/abnormality of the event.
func (m *BaseStat) Update(val float64) int32 {
	return m.dspot.Step(val)
}

// SetDSpotConfig sets the config of the DSpot instance embedded
// in the statistics.
func (m *BaseStat) SetDSpotConfig(sc gospot.DSpotConfig) {
	log.Debug().Msgf("(%s) Setting new config", m.name)
	m.dspot = gospot.NewDSpotFromConfig(sc)
}

// setCustomConfig builds a DSpotConfig instance according to the
// settings written in the config file
func setCustomConfig(statname string) gospot.DSpotConfig {
	statname = strings.ToLower(statname)

	prefix := "dspot." + statname + "."
	var conf gospot.DSpotConfig
	var setting string

	setting = prefix + "depth"
	if viper.IsSet(setting) {
		conf.Depth = viper.GetInt(setting)
	} else {
		conf.Depth = viper.GetInt("dspot.depth")
	}

	setting = prefix + "q"
	if viper.IsSet(setting) {
		conf.Q = viper.GetFloat64(setting)
	} else {
		conf.Q = viper.GetFloat64("dspot.q")
	}

	setting = prefix + "n_init"
	if viper.IsSet(setting) {
		conf.Ninit = viper.GetInt32(setting)
	} else {
		conf.Ninit = viper.GetInt32("dspot.n_init")
	}

	setting = prefix + "level"
	if viper.IsSet(setting) {
		conf.Level = viper.GetFloat64(setting)
	} else {
		conf.Level = viper.GetFloat64("dspot.level")
	}

	setting = prefix + "up"
	if viper.IsSet(setting) {
		conf.Up = viper.GetBool(setting)
	} else {
		conf.Up = viper.GetBool("dspot.up")
	}

	setting = prefix + "down"
	if viper.IsSet(setting) {
		conf.Down = viper.GetBool(setting)
	} else {
		conf.Down = viper.GetBool("dspot.down")
	}

	setting = prefix + "alert"
	if viper.IsSet(setting) {
		conf.Alert = viper.GetBool(setting)
	} else {
		conf.Alert = viper.GetBool("dspot.alert")
	}

	setting = prefix + "bounded"
	if viper.IsSet(setting) {
		conf.Bounded = viper.GetBool(setting)
	} else {
		conf.Bounded = viper.GetBool("dspot.bounded")
	}

	setting = prefix + "max_excess"
	if viper.IsSet(setting) {
		conf.MaxExcess = viper.GetInt32(setting)
	} else {
		conf.MaxExcess = viper.GetInt32("dspot.max_excess")
	}

	return conf
}

// StatFromName returns the StatInterface related to the
// given name. It returns an error when the desired statistic does
// not exist.
func StatFromName(statname string) (StatInterface, error) {
	s := gospot.NewDSpotFromConfig(setCustomConfig(statname))
	bs := BaseStat{name: statname, dspot: s}
	statConstructor, exists := AvailableStats[statname]
	if exists {
		return statConstructor(bs), nil
	}
	log.Error().Msg("Unknown statistics")
	return nil, errors.New("Unknown statistics")
}

func main() {
}
