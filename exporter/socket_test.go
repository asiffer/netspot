// socket_test.go

package exporter

import (
	"bytes"
	"encoding/csv"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/asiffer/netspot/config"
)

var (
	unixSocket = "/tmp/netspot_data.socket"
	socket     net.Listener
)

func init() {
	os.Remove(unixSocket)
}

func toStrMap(m map[string]interface{}) map[string]string {
	out := make(map[string]string)
	for k, v := range m {
		out[k] = fmt.Sprintf("%v", v)
	}
	return out
}

func Error(msg string, e error) {
	fmt.Printf("[\033[31;1mERROR\033[0m] %s (%v)\n", msg, e)
}

func startSocket(addr string) (net.Listener, error) {
	// data socket
	proto, endpoint, err := parseAddress(addr)
	if err != nil {
		return nil, err
	}
	// remove if it exists
	if _, err := os.Stat(endpoint); err == nil {
		// path to endpoint exists
		os.Remove(endpoint)
	}
	return net.Listen(proto, endpoint)
}

func basicSocket() *Socket {
	conn, err := net.Dial("unix", unixSocket)
	if err != nil {
		Error("cannot connect to socket", err)
		panic(err)
	}
	return &Socket{
		data:       true,
		alarm:      false,
		seriesName: "testSocket",
		tag:        "testNetspot",
		dataConn:   conn,
		alarmConn:  nil,
		format:     "csv",
	}
}

func basicAlarmSocket() *Socket {
	s := basicSocket()
	s.alarm = true
	s.data = false
	s.dataConn, s.alarmConn = s.alarmConn, s.dataConn
	return s
}

func getDataAlarmSockets() (string, string, error) {
	data, err := config.GetString("exporter.socket.data")
	if err != nil {
		return "", "", err
	}
	alarm, err := config.GetString("exporter.socket.alarm")
	if err != nil {
		return "", "", err
	}
	return data, alarm, nil
}

func initDataAlarmSockets() (net.Listener, net.Listener, error) {
	// init listening sockets
	data, alarm, err := getDataAlarmSockets()
	if err != nil {
		return nil, nil, err
	}

	dataSocket, err := startSocket(data)
	if err != nil {
		return nil, nil, err
	}

	alarmSocket, err := startSocket(alarm)
	if err != nil {
		return nil, nil, err
	}
	return dataSocket, alarmSocket, nil
}

func TestInitStartCloseSocket(t *testing.T) {
	title(t.Name())

	// init config
	if err := setFullConfig(); err != nil {
		t.Fatal(err)
	}

	// init listening sockets
	dataSocket, alarmSocket, err := initDataAlarmSockets()
	if err != nil {
		t.Fatal(err)
	}

	s := Socket{}

	checkTitle("Initialization")
	if err := s.Init(); err != nil {
		testERROR()
		t.Fatal(err)
	} else {
		testOK()
		defer Unload(s.Name())
	}

	if !s.LogsData() {
		t.Errorf("Expecting data logging")
	}

	if !s.LogsAlarm() {
		t.Errorf("Expecting alarm logging")
	}

	checkTitle("Start")
	if err := s.Start("wtf"); err != nil {
		testERROR()
		t.Error(err)
	} else {
		testOK()
	}

	checkTitle("Close")
	if err := s.Close(); err != nil {
		testERROR()
		t.Error(err)
	} else {
		testOK()
	}

	dataSocket.Close()
	alarmSocket.Close()
}

func TestStartSocket(t *testing.T) {
	title(t.Name())
	// Clear()
	// init config
	if err := setFullConfig(); err != nil {
		t.Fatal(err)
	}

	// init listening sockets
	dataSocket, alarmSocket, err := initDataAlarmSockets()
	if err != nil {
		t.Fatal(err)
	}

	s := Socket{}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer Unload(s.Name())

	if err := s.Start("wtf"); err != nil {
		t.Fatal(err)
	}

	checkTitle("Checking data export")
	if s.data && (s.dataConn.RemoteAddr().String() == dataSocket.Addr().String()) {
		testOK()
	} else {
		testERROR()
		t.Errorf("Expecting %s, got %s",
			dataSocket.Addr().String(),
			s.dataConn.RemoteAddr().String())
	}

	checkTitle("Checking alarm export")
	if s.alarm && (s.alarmConn.RemoteAddr().String() == alarmSocket.Addr().String()) {
		testOK()
	} else {
		testERROR()
		t.Errorf("Expecting %s, got %s",
			alarmSocket.Addr().String(),
			s.dataConn.RemoteAddr().String())
	}

	checkTitle("Closing")
	err = s.Close()
	if err != nil {
		testERROR()
		t.Error(err)
	} else {
		testOK()
	}

	dataSocket.Close()
	alarmSocket.Close()
}

func TestCSVEncoder(t *testing.T) {
	title(t.Name())
	config.Clean()
	config.InitConfig()

	// // init config
	// if err := setFullConfig(); err != nil {
	// 	t.Fatal(err)
	// }

	c := []byte(`
	[exporter.socket]
	format = "csv"
	data = "unix:///tmp/netspot_data.socket"
	tag = "netspot"`)
	if err := config.LoadForTestRawToml(c); err != nil {
		t.Fatal(err)
	}

	// delete(viper.Get("exporter.socket").(map[string]interface{}), "alarm")

	config.Debug()
	// init listening sockets
	dataAddr, err := config.GetString("exporter.socket.data")
	if err != nil {
		t.Fatal(err)
	}
	socket, err := startSocket(dataAddr)
	if err != nil {
		t.Fatal(err)
	}

	s := Socket{}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer Unload(s.Name())

	if err := s.Start("wtf"); err != nil {
		t.Fatal(err)
	}

	// accept connection
	conn, err := socket.Accept()
	if err != nil {
		t.Error(err)
	}

	// prepare data
	now := time.Now()
	data := map[string]float64{"stat0": 15.2, "stat1": -3.33333333}

	// receiver
	buffer := make([]byte, 2048)

	// t.Logf("%v+\n", s.Status())
	// send data
	if err := s.Write(now, data); err != nil {
		t.Error(err)
	}

	// read data
	if n, err := conn.Read(buffer); err != nil {
		t.Error(err)
	} else {
		t.Log(n, "bytes read")
		t.Logf("%v\n", string(buffer[:n]))
	}

	// csv
	reader := bytes.NewReader(buffer)
	creader := csv.NewReader(reader)
	header, err := creader.Read()
	if err != nil {
		t.Error(err)
	}
	t.Log(header)

	value, err := creader.Read()
	if err != nil {
		t.Error(err)
	}
	t.Log(value)

	if err = s.Close(); err != nil {
		t.Error(err)
	}

	if err = socket.Close(); err != nil {
		t.Error(err)
	}

}

func TestJSONEncoder(t *testing.T) {
	title(t.Name())
	config.Clean()
	config.InitConfig()
	// init config
	// if err := setFullConfig(); err != nil {
	// 	t.Fatal(err)
	// }
	c := []byte(`
	[exporter.socket]
	format = "json"
	data = "unix:///tmp/netspot_data.socket"
	tag = "netspot"`)
	if err := config.LoadForTestRawToml(c); err != nil {
		t.Fatal(err)
	}

	// init listening sockets
	dataAddr, err := config.GetString("exporter.socket.data")
	if err != nil {
		t.Fatal(err)
	}
	socket, err := startSocket(dataAddr)
	if err != nil {
		t.Fatal(err)
	}

	s := Socket{}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer Unload(s.Name())

	if err := s.Start("wtf"); err != nil {
		t.Fatal(err)
	}

	// accept connection
	conn, err := socket.Accept()
	if err != nil {
		t.Error(err)
	}

	// prepare data
	now := time.Now()
	data := map[string]float64{"stat0": 415.02, "stat1": -38.339}

	// receiver
	buffer := make([]byte, 2048)

	// send data
	s.format = "json"
	s.Write(now, data)

	// read data
	conn.Read(buffer)

	// json
	reader := bytes.NewReader(buffer)
	jdecoder := json.NewDecoder(reader)

	m := make(map[string]interface{})
	jdecoder.Decode(&m)

	t.Log(m)

	if err = s.Close(); err != nil {
		t.Error(err)
	}

	if err = socket.Close(); err != nil {
		t.Error(err)
	}

}

func TestGobEncoder(t *testing.T) {
	title(t.Name())

	config.Clean()
	config.InitConfig()

	c := []byte(`
	[exporter.socket]
	format = "gob"
	data = "unix:///tmp/netspot_data.socket"
	tag = "netspot"`)
	if err := config.LoadForTestRawToml(c); err != nil {
		t.Fatal(err)
	}

	// init listening sockets
	dataAddr, err := config.GetString("exporter.socket.data")
	if err != nil {
		t.Fatal(err)
	}
	socket, err := startSocket(dataAddr)
	if err != nil {
		t.Fatal(err)
	}

	s := Socket{}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer Unload(s.Name())

	if err := s.Start("wtf"); err != nil {
		t.Fatal(err)
	}

	// accept connection
	conn, err := socket.Accept()
	if err != nil {
		t.Error(err)
	}

	// prepare data
	now := time.Now()
	data := map[string]float64{"stat0": 78415.02, "stat1": 5558.05, "stat2": -5}

	// receiver
	buffer := make([]byte, 2048)

	// send data
	s.format = "gob"
	s.Write(now, data)

	// read data
	conn.Read(buffer)

	// gob
	reader := bytes.NewReader(buffer)
	gdecoder := gob.NewDecoder(reader)

	m := make(map[string]interface{})
	err = gdecoder.Decode(&m)

	t.Log(m)

	if err = s.Close(); err != nil {
		t.Error(err)
	}

	if err = socket.Close(); err != nil {
		t.Error(err)
	}
}

func TestSocketWarn(t *testing.T) {
	title(t.Name())
	Zero()
	config.Clean()
	config.InitConfig()

	c := []byte(`
	[exporter.socket]
	format = "csv"
	alarm = "unix:///tmp/netspot_alarm.socket"
	tag = "netspot"`)
	if err := config.LoadForTestRawToml(c); err != nil {
		t.Fatal(err)
	}

	// init listening sockets
	alarmAddr, err := config.GetString("exporter.socket.alarm")
	if err != nil {
		t.Fatal(err)
	}
	socket, err := startSocket(alarmAddr)
	if err != nil {
		t.Fatal(err)
	}

	s := Socket{}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer Unload(s.Name())

	if err := s.Start("wtf"); err != nil {
		t.Fatal(err)
	}

	// accept connection
	conn, err := socket.Accept()
	if err != nil {
		t.Error(err)
	}

	// prepare data
	now := time.Now()
	alert := SpotAlert{
		Status:      "UP_ALERT",
		Stat:        "R_SYN",
		Value:       0.995,
		Code:        1,
		Probability: 1e-8,
	}

	// receiver
	buffer := make([]byte, 2048)

	// send data
	s.format = "gob"
	err = s.Warn(now, &alert)
	if err != nil {
		t.Error(err)
	}

	// read data
	conn.Read(buffer)

	// gob
	reader := bytes.NewReader(buffer)
	gdecoder := gob.NewDecoder(reader)

	m := make(map[string]interface{})
	err = gdecoder.Decode(&m)

	t.Log(m)

	if err = s.Close(); err != nil {
		t.Error(err)
	}

	if err = socket.Close(); err != nil {
		t.Error(err)
	}
}

func TestSocketStatus(t *testing.T) {
	Zero()
	title(t.Name())
	if err := setFullConfig(); err != nil {
		t.Error(err)
	}

	s := Socket{}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	defer Unload(s.Name())

	// t.Logf("%v+\n", s.Status())
}
