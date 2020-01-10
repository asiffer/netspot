// console.go
//

package exporter

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// Console is the basic console exporter
type Console struct {
	data        bool
	alarm       bool
	dataLogger  zerolog.Logger
	alarmLogger zerolog.Logger
}

// Name returns the name of the exporter
func (c *Console) Name() string {
	return "console"
}

// Init defines the options of the exporter
func (c *Console) Init(...interface{}) error {
	if !viper.IsSet("console") {
		return fmt.Errorf("The section %s has not been found", "console")
	}
	// set everything to false
	c.data, c.alarm = false, false

	// update options
	if viper.IsSet("console.data") {
		c.data = viper.GetBool("console.data")
	}

	if viper.IsSet("console.alarm") {
		c.alarm = viper.GetBool("console.alarm")
	}

	// init loggers
	c.dataLogger = zerolog.New(os.Stdout).With().Logger()
	c.alarmLogger = zerolog.New(os.Stderr).With().Logger()
	return nil
}

// Write logs data
func (c *Console) Write(t time.Time, data map[string]float64) error {
	if c.data {
		dlog := c.dataLogger.Log().Time("time", t)
		for key, value := range data {
			dlog.Float64(key, value)
		}
		dlog.Send()
	}
	return nil
}

// Warn logs alarms
func (c *Console) Warn(t time.Time, s *SpotAlert) error {
	if c.alarm {
		wlog := c.alarmLogger.Log().Time("time", t)
		wlog.Str("Status", s.Status).
			Str("Stat", s.Stat).
			Float64("Value", s.Value).
			Int("Code", s.Code).
			Float64("Probability", s.Probability).Send()
	}
	return nil
}

// Close does nothing here
func (c *Console) Close() error {
	return nil
}
