// influxdb1_test.go

package exporter

import (
	"netspot/config"
	"os"
	"testing"
	"time"
)

func init() {
	if isRunningInDockerContainer() {
		config.SetValue("exporter.influxdb.address", "http://influxdb:8086")
	}
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
