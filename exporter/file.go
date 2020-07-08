// file.go

package exporter

import (
	"fmt"
	"netspot/config"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const (
	dataFileFormat  = "netspot_data_%s.json"  // format to name the file where data are logged
	alarmFileFormat = "netspot_alarm_%s.json" // format to name the file where anomalies are logged
)

// File is the file logger
type File struct {
	data             bool
	alarm            bool
	dataAddress      string
	alarmAddress     string
	seriesName       string
	dataLogger       zerolog.Logger
	alarmLogger      zerolog.Logger
	dataFileHandler  *os.File
	alarmFileHandler *os.File
}

func init() {
	Register(&File{})

}

// Main functions =========================================================== //
// ========================================================================== //
// ========================================================================== //

// Name returns the name of the exporter
func (f *File) Name() string {
	return "file"
}

// Options return the parameters of the shipper
// func (f *File) Options() map[string]string {
// 	return map[string]string{
// 		"data": `A string which gives the file where data will be stored.
// The value can contain a '%s' which will be replaced by the series name.
// Ex: netspot_%s_data.json`,
// 		"alarm": `A string which gives the file where alarms will be stored.
// The value can contain a '%s' which will be replaced by the series name.
// Ex: netspot_%s_alarm.json`,
// 	}
// }

// Status return the status of the module
func (f *File) Status() map[string]interface{} {
	m := make(map[string]interface{})
	if f.data {
		m["data"] = f.dataAddress
	}
	if f.alarm {
		m["alarm"] = f.alarmAddress
	}
	return m
}

// Init reads the config of the modules
func (f *File) Init() error {
	var err error

	f.data = config.HasKey("exporter.file.data")
	f.alarm = config.HasKey("exporter.file.alarm")

	if f.data {
		f.dataAddress, err = config.GetPath("exporter.file.data")
		if err != nil {
			return err
		}
	}

	if f.alarm {
		f.alarmAddress, err = config.GetPath("exporter.file.alarm")
		if err != nil {
			return err
		}
	}

	if f.data || f.alarm {
		return Load(f.Name())
	}
	return nil
}

// Start generate the connection from the module to the endpoint
func (f *File) Start(series string) error {
	f.seriesName = series

	// init loggers
	//
	// data logger
	if f.data {
		if err := f.initDataLogger(); err != nil {
			return err
		}
	}
	// alarm logger
	if f.alarm {
		if err := f.initAlarmLogger(); err != nil {
			return err
		}
	}
	return nil
}

// Write logs data
func (f *File) Write(t time.Time, data map[string]float64) error {
	if f.data {
		dlog := f.dataLogger.Log().Time("time", t)
		for key, value := range data {
			dlog.Float64(key, value)
		}
		dlog.Send()
	}
	return nil
}

// Warn logs alarms
func (f *File) Warn(t time.Time, s *SpotAlert) error {
	if f.alarm {
		wlog := f.alarmLogger.Log().Time("time", t)
		wlog.Str("Status", s.Status).
			Str("Stat", s.Stat).
			Float64("Value", s.Value).
			Int("Code", s.Code).
			Float64("Probability", s.Probability).Send()
	}
	return nil
}

// Close does nothing here
func (f *File) Close() error {
	if f.alarmFileHandler != nil {
		if err := f.alarmFileHandler.Close(); err != nil {
			return fmt.Errorf("Error while closing '%s' module (%v)", f.Name(), err)
		}
	}
	if f.dataFileHandler != nil {
		if err := f.dataFileHandler.Close(); err != nil {
			return fmt.Errorf("Error while closing '%s' module (%v)", f.Name(), err)
		}
	}
	return nil
}

// LogsData tells whether the module logs data
func (f *File) LogsData() bool {
	return f.data
}

// LogsAlarm tells whether the module logs alarm
func (f *File) LogsAlarm() bool {
	return f.alarm
}

// Side functions =========================================================== //
// ========================================================================== //
// ========================================================================== //

func (f *File) initDataLogger() error {
	var err error
	var dataFilePath string

	// get path
	format := f.dataAddress
	if strings.Contains(format, "%s") {
		dataFilePath, err = filepath.Abs(fmt.Sprintf(format, f.seriesName))
	} else {
		dataFilePath, err = filepath.Abs(format)
	}
	// check format
	if err != nil {
		return fmt.Errorf("Error while formatting file path for data (%w)", err)
	}

	// create file
	f.dataFileHandler, err = os.Create(dataFilePath)
	if err != nil {
		return fmt.Errorf("Error while creating data log file %s (%w)", dataFilePath, err)
	}

	// init logger
	f.dataLogger = zerolog.New(f.dataFileHandler).With().Logger()
	return nil
}

func (f *File) initAlarmLogger() error {
	var err error
	var alarmFilePath string

	// get path
	format := f.alarmAddress
	if strings.Contains(format, "%s") {
		alarmFilePath, err = filepath.Abs(fmt.Sprintf(format, f.seriesName))
	} else {
		alarmFilePath, err = filepath.Abs(format)
	}
	// check format
	if err != nil {
		return fmt.Errorf("Error while formatting file path for alarm (%w)", err)
	}

	// create file
	f.alarmFileHandler, err = os.Create(alarmFilePath)
	if err != nil {
		return fmt.Errorf("Error while creating alarm log file %s (%w)", alarmFilePath, err)
	}

	// init logger
	f.alarmLogger = zerolog.New(f.alarmFileHandler).With().Logger()
	return nil
}
