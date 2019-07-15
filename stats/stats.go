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
	Update(val float64) int             // feed dspot
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
func (m *BaseStat) Update(val float64) int {
	return m.dspot.Step(val)
}

// SetDSpotConfig sets the config of the DSpot instance embedded
// in the statistics.
func (m *BaseStat) SetDSpotConfig(sc gospot.DSpotConfig) {
	log.Debug().Msgf("(%s) Setting new config", m.name)
	m.dspot = gospot.NewDSpotFromConfig(&sc)
}

// changeConfigFromMap changes an attribute of the DSpot configuration
func (m *BaseStat) changeConfigFromMap(attributes map[string]interface{}) error {
	log.Debug().Msgf("(%s) Changing DSpot config", m.name)
	initialConf := m.dspot.Config()
	var ok bool
	var integerValue int
	var floatValue float64
	var boolValue bool

	// message errors
	var msg string
	//"The following keys have created some errors: "

	for key, value := range attributes {
		ok = true

		switch key {
		case "depth", "Depth":
			integerValue, ok = value.(int)
			if !ok {
				log.Error().Msgf("Error while changing depth to %v", value)
			} else {
				log.Debug().Msgf("Changing depth to %d", integerValue)
				initialConf.Depth = integerValue
			}
		case "q", "Q":
			floatValue, ok = value.(float64)
			if !ok {
				log.Error().Msgf("Error while changing q to %v", value)
			} else {
				log.Debug().Msgf("Changing q to %f", floatValue)
				initialConf.Q = floatValue
			}
		case "n_init", "Ninit":
			integerValue, ok = value.(int)
			if !ok {
				log.Error().Msgf("Error while changing n_init to %v", value)
			} else {
				log.Debug().Msgf("Changing n_init to %d", integerValue)
				initialConf.Ninit = integerValue
			}
		case "level", "Level":
			floatValue, ok = value.(float64)
			if !ok {
				log.Error().Msgf("Error while changing level to %v", value)
			} else {
				log.Debug().Msgf("Changing level to %f", floatValue)
				initialConf.Level = floatValue
			}
		case "up", "Up":
			boolValue, ok = value.(bool)
			if !ok {
				log.Error().Msgf("Error while changing up to %v", value)
			} else {
				log.Debug().Msgf("Changing up to %t", boolValue)
				initialConf.Up = boolValue
			}
		case "down", "Down":
			boolValue, ok = value.(bool)
			if !ok {
				log.Error().Msgf("Error while changing down to %v", value)
			} else {
				log.Debug().Msgf("Changing down to %t", boolValue)
				initialConf.Down = boolValue
			}
		case "alert", "Alert":
			boolValue, ok = value.(bool)
			if !ok {
				log.Error().Msgf("Error while changing alert to %v", value)
			} else {
				log.Debug().Msgf("Changing alert to %t", boolValue)
				initialConf.Alert = boolValue
			}
		case "bounded", "Bounded":
			boolValue, ok = value.(bool)
			if !ok {
				log.Error().Msgf("Error while changing bounded to %v", value)
			} else {
				log.Debug().Msgf("Changing bounded to %t", boolValue)
				initialConf.Bounded = boolValue
			}
		case "max_excess", "MaxExcess":
			integerValue, ok = value.(int)
			if !ok {
				log.Error().Msgf("Error while changing max_excess to %v", value)
			} else {
				log.Debug().Msgf("Changing max_excess to %d", integerValue)
				initialConf.MaxExcess = integerValue
			}
		default:
			ok = false
		}

		if !ok {
			msg += fmt.Sprintf("%s (%v) ", key, value)
		}
	}

	// setting the new config
	m.dspot = gospot.NewDSpotFromConfig(&initialConf)

	if len(msg) > 0 {
		// return errors
		return fmt.Errorf("The following keys have created some errors: %s", msg)
	}
	return nil
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
		conf.Ninit = viper.GetInt(setting)
	} else {
		conf.Ninit = viper.GetInt("dspot.n_init")
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
		conf.MaxExcess = viper.GetInt(setting)
	} else {
		conf.MaxExcess = viper.GetInt("dspot.max_excess")
	}

	return conf
}

// StatFromName returns the StatInterface related to the
// given name. It returns an error when the desired statistic does
// not exist.
func StatFromName(statname string) (StatInterface, error) {
	conf := setCustomConfig(statname)
	s := gospot.NewDSpotFromConfig(&conf)
	bs := BaseStat{name: statname, dspot: s}
	statConstructor, exists := AvailableStats[statname]
	if exists {
		return statConstructor(bs), nil
	}
	log.Error().Msg("Unknown statistics")
	return nil, errors.New("Unknown statistics")
}

// StatFromNameWithCustomConfig returns the StatInterface related to the
// given name. It returns an error when the desired statistic does
// not exist. Moreover you can pass a map giving the specific values
// to the DSpot instance which monitors the stat.
func StatFromNameWithCustomConfig(statname string, config map[string]interface{}) (StatInterface, error) {
	// set up the default conf (of those in the config file)
	conf := setCustomConfig(statname)
	s := gospot.NewDSpotFromConfig(&conf)
	bs := BaseStat{name: statname, dspot: s}
	// update the conf according to the passed parameters
	if config != nil {
		bs.changeConfigFromMap(config)
	}
	// build the stat
	statConstructor, exists := AvailableStats[statname]
	if exists {
		return statConstructor(bs), nil
	}
	log.Error().Msg("Unknown statistics")
	return nil, errors.New("Unknown statistics")
}

func main() {
}
