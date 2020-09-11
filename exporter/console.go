// console.go
//

package exporter

import (
	"encoding/json"
	"fmt"
	"netspot/config"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Console is the basic console exporter
type Console struct {
	data        bool
	alarm       bool
	dataLogger  zerolog.Logger
	alarmLogger zerolog.Logger
}

func init() {
	// register the exporter
	Register(&Console{})
}

// Name returns the name of the exporter
func (c *Console) Name() string {
	return "console"
}

// Status return the status of the shipper
func (c *Console) Status() map[string]interface{} {
	return map[string]interface{}{
		"data":  c.data,
		"alarm": c.alarm,
	}
}

// Init defines the options of the module from the config
func (c *Console) Init() error {
	c.data = config.MustBool("exporter.console.data")
	c.alarm = config.MustBool("exporter.console.alarm")

	if c.data || c.alarm {
		return Load(c.Name())
	}
	return nil
}

// Start generate the connection from the shipper to the endpoint
func (c *Console) Start(string) error {
	// init loggers
	c.dataLogger = zerolog.New(os.Stdout).With().Logger()
	c.alarmLogger = zerolog.New(os.Stdout).With().Logger()
	return nil
}

// Write logs data
func (c *Console) Write(t time.Time, data map[string]float64) error {
	// fmt.Println(data)
	if c.data {
		fmt.Println(jsonifyWithTime(t, data))
	}
	return nil
}

// Warn logs alarms
func (c *Console) Warn(t time.Time, s *SpotAlert) error {
	if c.alarm {
		alarm := map[string]interface{}{
			"status":      s.Status,
			"stat":        s.Stat,
			"value":       s.Value,
			"code":        s.Code,
			"probability": s.Probability,
		}
		if b, err := json.Marshal(alarm); err == nil {
			fmt.Println(string(b))
		}
	}
	return nil
}

// Close does nothing here
func (c *Console) Close() error {
	return nil
}

// LogsData tells whether the shipper logs data
func (c *Console) LogsData() bool {
	return c.data
}

// LogsAlarm tells whether the shipper logs alarm
func (c *Console) LogsAlarm() bool {
	return c.alarm
}
