// file.go

package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

const (
	dataFileFormat      = "netspot_raw_%s.json"   // format to name the file where data are logged
	alarmFileNameFormat = "netspot_alarm_%s.json" // format to name the file where anomalies are logged
)

// File is the file logger
type File struct {
	data        bool
	alarm       bool
	name        string
	dataLogger  zerolog.Logger
	alarmLogger zerolog.Logger
}

func init() {
	// default config
	viper.SetDefault("file.data", dataFileFormat)
	viper.SetDefault("file.alarm", alarmFileNameFormat)
}

func getSeriesName(opts ...interface{}) (string, error) {
	if len(opts) == 0 {
		return "", fmt.Errorf("Series name is not set")
	}
	name, ok := opts[0].(string)
	if ok {
		return name, nil
	}
	return "", fmt.Errorf("The series name is not valid (type assertion failed)")
}

func (f *File) initDataLogger() error {
	var err error
	var dataFilePath string

	// get path
	format := viper.GetString("file.data")
	if strings.Contains(format, "%s") {
		dataFilePath, err = filepath.Abs(fmt.Sprintf(format, f.name))
	} else {
		dataFilePath, err = filepath.Abs(format)
	}
	// check format
	if err != nil {
		return fmt.Errorf("Error while formatting file path for data (%w)", err)
	}

	// create file
	dataFile, err := os.Create(dataFilePath)
	if err != nil {
		return fmt.Errorf("Error while creating data log file %s (%w)", dataFilePath, err)
	}
	defer dataFile.Close()

	// init logger
	f.dataLogger = zerolog.New(dataFile).With().Logger()
	return nil
}

func (f *File) initAlarmLogger() error {
	var err error
	var alarmFilePath string

	// get path
	format := viper.GetString("file.alarm")
	if strings.Contains(format, "%s") {
		alarmFilePath, err = filepath.Abs(fmt.Sprintf(format, f.name))
	} else {
		alarmFilePath, err = filepath.Abs(format)
	}
	// check format
	if err != nil {
		return fmt.Errorf("Error while formatting file path for alarm (%w)", err)
	}

	// create file
	alarmFile, err := os.Create(alarmFilePath)
	if err != nil {
		return fmt.Errorf("Error while creating alarm log file %s (%w)", alarmFilePath, err)
	}
	defer alarmFile.Close()

	// init logger
	f.alarmLogger = zerolog.New(alarmFile).With().Logger()
	return nil
}

// Name returns the name of the exporter
func (f *File) Name() string {
	return "file"
}

// Init defines the options of the exporter
// It supports a single option: the name of the series
// which should be a string. The series name could be "".
func (f *File) Init(options ...interface{}) error {
	if !viper.IsSet("file") {
		return fmt.Errorf("The section %s has not been found", "file")
	}

	// get series name
	name, err := getSeriesName(options)
	if err != nil {
		return err
	}
	f.name = name

	// set everything to false
	f.data, f.alarm = false, false
	// update options
	f.data = viper.IsSet("file.data")
	f.alarm = viper.IsSet("file.alarm")

	// init loggers
	//
	// data logger
	if f.data {
		if err = f.initDataLogger(); err != nil {
			return err
		}
	}
	// alarm logger
	if f.alarm {
		if err = f.initAlarmLogger(); err != nil {
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
	return nil
}
