// socket.go

package exporter

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"netspot/config"
	"strings"
	"time"
)

var (
	supportedSocketFormats = []string{"csv", "json", "gob"}
)

// Socket is the socket logger
type Socket struct {
	data         bool
	alarm        bool
	dataProto    string
	dataAddress  string
	alarmProto   string
	alarmAddress string
	seriesName   string
	tag          string
	dataConn     net.Conn
	alarmConn    net.Conn
	format       string // csv, json, binary ...
}

func init() {
	RegisterAndSetDefaults(&Socket{}, map[string]interface{}{
		"exporter.socket.format": "json",
		"exporter.socket.tag":    "netspot",
	})
}

func checkFormat(f string) error {
	for _, s := range supportedSocketFormats {
		if f == s {
			return nil
		}
	}
	return fmt.Errorf("The format %s is not accepted (only csv, json and gob)", f)
}

// Main functions =========================================================== //
// ========================================================================== //
// ========================================================================== //

// Name returns the name of the exporter
func (s *Socket) Name() string {
	return "socket"
}

// Options return the parameters of the shipper
// func (s *Socket) Options() map[string]string {
// 	return map[string]string{
// 		"data": `A string which gives the socket where data will be sent.
// The format is the following: [proto]://[address].
// Ex: unix:///tmp/netspot_data.socket
//     tcp://127.0.0.1:4000`,
// 		"alarm": `A string which gives the socket where alarms will be sent.
// The format is the following: [proto]://[address].
// Ex: unix:///tmp/netspot_alarm.socket
//     udp://127.0.0.1:4000`,
// 		"tag": "A string which sets a tag when data are sent ('netspot' by default)",
// 		"format": `A string which defines the sending format (data/alarm).
// It currently accepts "csv", "json" and "gob" (golang binary format)`,
// 	}
// }

// Status return the status of the shipper
func (s *Socket) Status() map[string]interface{} {
	m := map[string]interface{}{
		"tag":    s.tag,
		"format": s.format,
	}
	if s.data {
		m["data"] = s.dataAddress
		if s.alarm {
			m["alarm"] = s.alarmAddress
		}
	}
	return m
}

// Init reads the config of the shipper
func (s *Socket) Init() error {
	var err error
	// update options
	s.data = config.HasKey("exporter.socket.data")
	s.alarm = config.HasKey("exporter.socket.alarm")

	// retrieve addresses (no check is done)
	if s.data {
		s.dataProto, s.dataAddress, err = config.GetSocket("exporter.socket.data")
		if err != nil {
			return err
		}
	}

	if s.alarm {
		s.alarmProto, s.alarmAddress, err = config.GetSocket("exporter.socket.alarm")
		if err != nil {
			return err
		}
	}

	// tag
	s.tag, err = config.GetString("exporter.socket.tag")
	if err != nil {
		return err
	}
	// data format
	s.format, err = config.GetString("exporter.socket.format")
	if err != nil {
		return err
	}

	if err := checkFormat(s.format); err != nil {
		return err
	}

	if s.data || s.alarm {
		return Load(s.Name())
	}
	return nil
}

// Start generate the connection from the shipper to the endpoint
func (s *Socket) Start(series string) error {
	var err error
	s.seriesName = series
	// init sockets
	//
	// data socket
	if s.data {
		if s.dataConn, err = net.Dial(s.dataProto, s.dataAddress); err != nil {
			return err
		}
	}
	// alarm socket
	if s.alarm {
		if s.alarmConn, err = net.Dial(s.alarmProto, s.alarmAddress); err != nil {
			return err
		}
	}
	return nil
}

// Write logs data
func (s *Socket) Write(t time.Time, data map[string]float64) error {
	var bin []byte
	var err error
	raw := untypeMap(data)
	raw["type"] = "data"
	raw["time"] = t.UnixNano()
	raw["series"] = s.seriesName
	raw["name"] = s.tag

	if s.data {
		switch s.format {
		case "csv":
			bin, err = toCSV(raw)
		case "json":
			bin, err = toJSON(raw)
		case "gob":
			bin, err = toGob(raw)
		default:
			return fmt.Errorf("Bad data format (%s)", s.format)
		}
		_, err = s.dataConn.Write(bin)
		return err
	}

	return nil
}

// Warn logs alarms
func (s *Socket) Warn(t time.Time, x *SpotAlert) error {
	var bin []byte
	var err error
	raw := x.toUntypedMap()
	raw["type"] = "alarm"
	raw["time"] = t.UnixNano()
	raw["series"] = s.seriesName
	raw["name"] = s.tag
	if s.alarm {
		switch s.format {
		case "csv":
			bin, err = toCSV(raw)
		case "json":
			bin, err = toJSON(raw)
		case "gob":
			bin, err = toGob(raw)
		default:
			return fmt.Errorf("Bad data format (%s)", s.format)
		}
		_, err = s.alarmConn.Write(bin)
		return err
	}
	return nil
}

// Close does nothing here
func (s *Socket) Close() error {
	if s.dataConn != nil {
		if err := s.dataConn.Close(); err != nil {
			return fmt.Errorf("Error while closing '%s' shipper (%v)", s.Name(), err)
		}
	}
	if s.alarmConn != nil {
		if err := s.alarmConn.Close(); err != nil {
			return fmt.Errorf("Error while closing '%s' shipper (%v)", s.Name(), err)
		}
	}
	return nil
}

// LogsData tells whether the shipper logs data
func (s *Socket) LogsData() bool {
	return s.data
}

// LogsAlarm tells whether the shipper logs alarm
func (s *Socket) LogsAlarm() bool {
	return s.alarm
}

// Side functions =========================================================== //
// ========================================================================== //
// ========================================================================== //

// func checkSocketFormat(format string) error {
// 	for _, f := range supportedSocketFormats {
// 		if format == f {
// 			return nil
// 		}
// 	}
// 	return fmt.Errorf("Data format '%s' is not supported", format)
// }

func toCSV(raw map[string]interface{}) ([]byte, error) {
	header := make([]string, 0)
	values := make([]string, 0)
	for k, v := range raw {
		header = append(header, k)
		values = append(values, fmt.Sprintf("%v", v))
	}
	str := fmt.Sprintf("%s\n%s",
		strings.Join(header, ","),
		strings.Join(values, ","))
	return []byte(str), nil
}

func toJSON(raw map[string]interface{}) ([]byte, error) {
	js, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	return js, err
}

func toGob(raw map[string]interface{}) ([]byte, error) {
	var buffer bytes.Buffer // Stand-in for the network.

	// Create an encoder and send a value.
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(raw)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
