// influxdb1_test.go

package exporter

import (
	"math/rand"
	"testing"
	"time"

	"github.com/asiffer/netspot/config"
)

func TestInitInflux(t *testing.T) {
	title(t.Name())
	Zero()
	if err := setFullConfig(); err != nil {
		t.Error(err)
	}

	checkTitle("Initialization")
	inf := InfluxDB{}

	if err := inf.Init(); err != nil {
		testERROR()
		t.Fatal(err)
	} else {
		testOK()
		defer Unload(inf.Name())
	}
	t.Log(inf)

	if !inf.LogsAlarm() {
		t.Errorf("Expecting alarm logging activated")
	}

	if !inf.LogsData() {
		t.Errorf("Expecting data logging activated")
	}

	checkTitle("Start")
	if err := inf.Start("wtf"); err != nil {
		testERROR()
		t.Error(err)
	} else {
		testOK()
	}

	checkTitle("Checking database connection")
	duration, str, err := inf.client.Ping(1 * time.Second)
	if err != nil {
		testERROR()
		t.Fatal(err)
	} else {
		testOK()
	}
	t.Log(duration, str)

	if err = inf.Close(); err != nil {
		t.Error(err)
	}
}

func TestInitInfluxBadConfiguration(t *testing.T) {
	inf := InfluxDB{}
	conf := map[string]interface{}{
		"exporter.influxdb.data":       true,
		"exporter.influxdb.alarm":      true,
		"exporter.influxdb.address":    "http://127.0.0.1:8086",
		"exporter.influxdb.database":   "netspot",
		"exporter.influxdb.username":   "netspot",
		"exporter.influxdb.password":   "netspot",
		"exporter.influxdb.batch_size": -5,
		"exporter.influxdb.agent_name": "local",
	}

	conf["exporter.influxdb.database"] = ""
	config.LoadForTest(conf)
	if err := inf.Init(); err == nil {
		t.Errorf("an error was expected")
	}

	conf["exporter.influxdb.password"] = ""
	config.LoadForTest(conf)
	if err := inf.Init(); err == nil {
		t.Errorf("an error was expected")
	}

	conf["exporter.influxdb.username"] = ""
	config.LoadForTest(conf)
	if err := inf.Init(); err == nil {
		t.Errorf("an error was expected")
	}

	conf["exporter.influxdb.address"] = ""
	config.LoadForTest(conf)
	if err := inf.Init(); err == nil {
		t.Errorf("an error was expected")
	}

	conf["exporter.influxdb.batch_size"] = "@@@"
	config.LoadForTest(conf)
	if err := inf.Init(); err == nil {
		t.Errorf("an error was expected")
	}

	conf["exporter.influxdb.agent_name"] = ""
	config.LoadForTest(conf)
	if err := inf.Init(); err == nil {
		t.Errorf("an error was expected")
	}
}

func TestInfluxWriteAndWarn(t *testing.T) {
	title(t.Name())
	Zero()
	err := setFullConfig()
	if err != nil {
		t.Error(err)
	}

	inf := InfluxDB{}

	if err := inf.Init(); err != nil {
		t.Fatal(err)
	}
	defer Unload(inf.Name())
	inf.Start("wtf0")

	// prepare data
	now := time.Now()
	data := map[string]float64{"stat0": 15.2, "stat1": -3.33333333}

	checkTitle("Writing data")
	err = inf.Write(now, data)
	if err != nil {
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
	checkTitle("Sending alarm")
	err = inf.Warn(now, &alert)
	if err != nil {
		testERROR()
		t.Error(err)
	} else {
		testOK()
	}

	checkTitle("Flushing")
	err = inf.flush()
	if err != nil {
		testERROR()
		t.Error(err)
	} else {
		testOK()
	}

	checkTitle("Closing")
	err = inf.Close()
	if err != nil {
		testERROR()
		t.Error(err)
	} else {
		testOK()
	}

}

func TestInfluxStatus(t *testing.T) {
	title(t.Name())
	Zero()
	if err := setFullConfig(); err != nil {
		t.Error(err)
	}

	inf := InfluxDB{}
	if err := inf.Init(); err != nil {
		t.Fatal(err)
	}
	defer Unload(inf.Name())
}

func TestInfluxBatchWrite(t *testing.T) {
	title(t.Name())

	host := "http://localhost:8086"
	if isRunningInDockerContainer() {
		host = "http://influxdb:8086"
	}

	conf := map[string]interface{}{
		"exporter.influxdb.data":       true,
		"exporter.influxdb.alarm":      true,
		"exporter.influxdb.address":    host,
		"exporter.influxdb.database":   "netspot",
		"exporter.influxdb.username":   "netspot",
		"exporter.influxdb.password":   "netspot",
		"exporter.influxdb.batch_size": 20,
		"exporter.influxdb.agent_name": "local",
	}
	config.LoadForTest(conf)

	inf := available["influxdb"]

	if err := inf.Init(); err != nil {
		t.Fatal(err)
	}
	defer Unload(inf.Name())
	inf.Start("exporter-infludb-test")

	// send data multiple times
	checkTitle("Writing data")
	n, _ := config.GetStrictlyPositiveInt("exporter.influxdb.batch_size")
	for i := 0; i < 2*n; i++ {
		data := map[string]float64{"stat0": rand.Float64(), "stat1": rand.Float64()}
		if err := inf.Write(time.Now(), data); err != nil {
			testERROR()
			t.Fatal(err)
		}
	}
	testOK()
}
