// file.go

package exporter

import (
	"fmt"
	"netspot/config"
	"os"
	"strings"
	"time"
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
	dataFileHandler  *os.File
	alarmFileHandler *os.File
}

func init() {
	RegisterAndSetDefaults(&File{},
		map[string]interface{}{
			"exporter.file.data":  nil,
			"exporter.file.alarm": nil,
		})

}

// Main functions =========================================================== //
// ========================================================================== //
// ========================================================================== //

// Name returns the name of the exporter
func (f *File) Name() string {
	return "file"
}

// Init reads the config of the modules
func (f *File) Init() error {
	var err error

	f.data = config.HasNotNilKey("exporter.file.data")
	f.alarm = config.HasNotNilKey("exporter.file.alarm")

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
	var err error
	f.seriesName = series
	f.updateFileFromSeriesName()

	// init file handlers
	// data logger
	if f.data {
		if f.dataFileHandler, err = os.Create(f.dataAddress); err != nil {
			return err
		}
	}
	// alarm logger
	if f.alarm {
		if f.alarmFileHandler, err = os.Create(f.alarmAddress); err != nil {
			return err
		}
	}
	return nil
}

// Write logs data
func (f *File) Write(t time.Time, data map[string]float64) error {

	if f.data {
		// hope there is no problem
		f.dataFileHandler.WriteString(jsonifyWithTime(t, data))
		f.dataFileHandler.Write([]byte{'\n'})
	}
	return nil
}

// Warn logs alarms
func (f *File) Warn(t time.Time, s *SpotAlert) error {
	if f.alarm {
		f.alarmFileHandler.WriteString(s.toJSONwithTime(t))
		f.alarmFileHandler.Write([]byte{'\n'})
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

// // Side functions =========================================================== //
// // ========================================================================== //
// // ========================================================================== //

func (f *File) updateFileFromSeriesName() {
	if strings.Contains(f.dataAddress, "%s") {
		f.dataAddress = fmt.Sprintf(f.dataAddress, f.seriesName)
	}

	if strings.Contains(f.alarmAddress, "%s") {
		f.alarmAddress = fmt.Sprintf(f.alarmAddress, f.seriesName)
	}
}
