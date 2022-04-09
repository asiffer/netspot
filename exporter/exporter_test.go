//  exporter_test.go

package exporter

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/asiffer/netspot/config"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

var (
	headerWidth = 80
	testDir     = getDir()
)

var fullConfig = []byte(`
[exporter.console]
data = true
alarm = true

[exporter.file]
data = "/tmp/netspot_%s_data.json"
alarm = "/tmp/netspot_%s_alarm.json"

[exporter.socket]
data = "unix:///tmp/netspot_data.socket"
alarm = "unix:///tmp/netspot_alarm.socket"
tag = "netspot"
format = "csv"

[exporter.influxdb]
data = true
alarm = true
address = "http://127.0.0.1:8086"
database = "netspot"
username = "netspot"
password = "netspot"
batch_size = -5
agent_name = "local"
`)

var errorConfig = []byte(`
[nothing]
nothing = 
`)

var strangeConfig = []byte(`
[exporter.file]
data = "/tmp/data_file.txt"
alarm = "/root/alarm_file.txt"

[exporter.socket]
data = "127.0.0.1"
alarm = "/tmp/file"

[exporter.influxdb]
address = "127.0.0.1:8086"
`)

func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir, _ := filepath.Abs(filepath.Dir(filename))
	// dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	testDir := path.Join(dir, "test")
	fmt.Println(testDir)
	zerolog.SetGlobalLevel(zerolog.NoLevel)
}

func parseAddress(conn string) (string, string, error) {
	raw := strings.Split(conn, "://")
	if len(raw) != 2 {
		return "", "", fmt.Errorf("The address is not valid, its format must be proto://address")
	}
	return raw[0], raw[1], nil
}

func getDir() string {
	_, filename, _, _ := runtime.Caller(0)
	dir, _ := filepath.Abs(filepath.Dir(filename))
	return path.Join(dir, "test")
}

func title(s string) {
	var l = len(s)
	var border int
	var left string
	var right string
	remaining := headerWidth - l - 2
	if remaining%2 == 0 {
		border = remaining / 2
		left = strings.Repeat("-", border) + " "
		right = " " + strings.Repeat("-", border)
	} else {
		border = (remaining - 1) / 2
		left = strings.Repeat("-", border+1) + " "
		right = " " + strings.Repeat("-", border)
	}

	fmt.Println(left + s + right)
}

func checkTitle(s string) {
	format := "%-" + fmt.Sprint(headerWidth-7) + "s"
	fmt.Printf(format, s)
}

func testOK() {
	fmt.Println("[\033[32mOK\033[0m]")
}

func testERROR() {
	fmt.Println("[\033[31mERROR\033[0m]")
}

func isRunningInDockerContainer() bool {
	// docker creates a .dockerenv file at the root
	// of the directory tree inside the container.
	// if this file exists then the viewer is running
	// from inside a container so return true
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return false
}

func setFullConfig() error {
	viper.SetConfigType("toml")
	if err := viper.ReadConfig(bytes.NewBuffer(fullConfig)); err != nil {
		return err
	}
	if isRunningInDockerContainer() {
		viper.Set("exporter.influxdb.address", "http://influxdb:8086")
	}
	return config.LoadForTest(viper.AllSettings())
}

func setErrorConfig() error {
	viper.SetConfigType("toml")
	if err := viper.ReadConfig(bytes.NewBuffer(errorConfig)); err != nil {
		return err
	}
	return config.LoadForTest(viper.AllSettings())
}

func setStrangeConfig() error {
	viper.SetConfigType("toml")
	if err := viper.ReadConfig(bytes.NewBuffer(strangeConfig)); err != nil {
		return err
	}
	return config.LoadForTest(viper.AllSettings())
}

func setNullConfig() error {
	return config.LoadForTest(nil)
}

func TestNothing(t *testing.T) {
	title("Testing exporter")
}

func TestCheckPath(t *testing.T) {
	title(t.Name())
	p := "/-xxx-/\\"
	if err := checkPath(p); err == nil {
		t.Errorf("The path %s should not exists", p)
	}
}

func TestLoadAll(t *testing.T) {
	title(t.Name())
	Zero()
	setNullConfig()
	for _, s := range available {
		err := Load(s.Name())
		if err != nil {
			t.Error(err)
		}
	}
	Zero()
}

func TestBasics(t *testing.T) {
	title(t.Name())
	// defer Clear()
	setNullConfig()
	InitLogger()

	checkTitle("Number of available modules")
	h := AvailableExportingModules()
	if len(h) != len(available) {
		testERROR()
		t.Errorf("Wrong number of module, expected %d, got %d",
			len(available),
			len(h))
	} else {
		testOK()
	}

	checkTitle("Avaibility")
	if isAvailable("WTF") {
		testERROR()
		t.Errorf("An unknown exporter (%s) is available", "WTF")
	} else {
		testOK()
	}

	checkTitle("Loading")

	if err := Load("console"); err != nil {
		testERROR()
		t.Error(err)
	} else {
		testOK()
	}

	// prepare data
	now := time.Now()
	data := map[string]float64{"stat0": 15.2, "stat1": -3.33333333}

	checkTitle("Writing data")

	if err := Write(now, data); err != nil {
		testERROR()
		t.Error(err)
	} else {
		testOK()
	}

	// prepare data
	alert := SpotAlert{
		Status:      "UP_ALERT",
		Stat:        "R_SYN",
		Value:       0.995,
		Code:        1,
		Probability: 1e-8,
	}

	checkTitle("Sending warning")

	if err := Warn(now, &alert); err != nil {
		testERROR()
		t.Error(err)
	} else {
		testOK()
	}

	checkTitle("Releasing")

	if err := Zero(); err != nil {
		testERROR()
		t.Error(err)
	} else {
		testOK()
	}

}

func TestLoading(t *testing.T) {
	title(t.Name())
	// defer Clear()
	setFullConfig()
	// zerolog.SetGlobalLevel(zerolog.Disabled)

	if err := Zero(); err != nil {
		t.Error(err)
	}

	for _, m := range available {
		checkTitle(fmt.Sprintf("Loading '%s'", m.Name()))
		if err := Load(m.Name()); err != nil {
			testERROR()
			t.Error(err)
		} else {
			testOK()
		}
	}

	for _, m := range available {
		checkTitle(fmt.Sprintf("Reloading '%s'", m.Name()))
		if err := Load(m.Name()); err == nil {
			testERROR()
			t.Error(err)
		} else {
			testOK()
		}
	}

	checkTitle("Resetting")

	if err := Zero(); err != nil {
		testERROR()
		t.Error(err)
	} else {
		testOK()
	}

}

func TestInit(t *testing.T) {
	title(t.Name())
	// defer Clear()
	InitLogger()
	InitConfig()
	// t.Logf("%+v\n", GenericStatus())
}

func TestBasicStart(t *testing.T) {
	title(t.Name())
	Zero()
	config.Clean()
	config.InitConfig()
	defer Zero()

	// config
	configExample := []byte(`
[exporter.console]
# A boolean to activate console data logging
data = true
# A boolean to activate console alarm logging
alarm = true
`)

	if err := config.LoadForTestRawToml(configExample); err != nil {
		t.Fatal(err)
	}

	config.Debug()
	InitLogger()
	InitConfig()

	// if err := Load("console"); err != nil {
	// 	t.Fatal(err)
	// }
	// t.Logf("%+v\n", GenericStatus())

	checkTitle("Starting")
	if err := Start("test"); err != nil {
		testERROR()
		t.Error(err)
	} else {
		testOK()
	}

	checkTitle("Writing")
	data := map[string]float64{"stat": 0.25}
	if err := Write(time.Now(), data); err != nil {
		testERROR()
		t.Error(err)
	} else {
		testOK()
	}

	checkTitle("Warning\n")
	alert := &SpotAlert{
		Status:      "UP_ALERT",
		Stat:        "R_SYN",
		Value:       0.99,
		Code:        1,
		Probability: 1e-8,
	}

	if err := Warn(time.Now(), alert); err != nil {
		testERROR()
		t.Error(err)
	} else {
		testOK()
	}

	checkTitle("Closing\n")
	if err := Close(); err != nil {
		testERROR()
		t.Error(err)
	} else {
		testOK()
	}

	// end stdout
	// w.Close()
	// out, _ := ioutil.ReadAll(os.Stdout)
	// os.Stdout = rescueStdout

	// fmt.Println(out, string(out))
}

func TestStartCloseAll(t *testing.T) {
	title("Testing start/close function")
	// defer Clear()
	if err := setFullConfig(); err != nil {
		t.Error(err)
	}

	checkTitle("Init full config")

	if err := InitConfig(); err != nil {
		testERROR()
		t.Error(err)
	} else {
		testOK()
	}

	dataAddr, err := config.GetString("exporter.socket.data")
	if err != nil {
		testERROR()
		t.Fatal(err)
	}
	startSocket(dataAddr)

	alarmAddr, err := config.GetString("exporter.socket.alarm")
	if err != nil {
		testERROR()
		t.Fatal(err)
	}
	startSocket(alarmAddr)

	checkTitle("Start all modules")
	series := fmt.Sprintf("full-%d", rand.Int())
	if err := Start(series); err != nil {
		testERROR()
		t.Error(err)
	} else {
		testOK()
	}

	checkTitle("Close all modules")
	if err := Close(); err != nil {
		testERROR()
		t.Error(err)
	} else {
		testOK()
	}
}
