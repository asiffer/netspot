// file_test.go

package exporter

import (
	"fmt"
	"io/ioutil"
	"netspot/config"
	"testing"
	"time"
)

const filePrefix = "exporter.file"

func init() {
	m := map[string]interface{}{
		filePrefix + ".data":  "/tmp/netspot_%s_data.json",
		filePrefix + ".alarm": "/tmp/netspot_%s_alarm.json",
	}
	if err := config.LoadForTest(m); err != nil {
		panic(err)
	}
}

func TestInitStartCloseFile(t *testing.T) {
	title("Testing File exporter")

	if err := setFullConfig(); err != nil {
		t.Fatal(err)
	}

	f := File{}

	checkTitle("Initialization")
	if err := f.Init(); err != nil {
		testERROR()
		t.Fatal(err)
	} else {
		testOK()
	}

	checkTitle("Start")
	if err := f.Start("wtf"); err != nil {
		testERROR()
		t.Error(err)
	} else {
		testOK()
	}

	checkTitle("Close")
	if err := f.Close(); err != nil {
		testERROR()
		t.Error(err)
	} else {
		testOK()
	}
}
func TestStartFile(t *testing.T) {
	title("Testing start of File exporter")

	if err := setFullConfig(); err != nil {
		t.Error(err)
	}

	f := File{}
	if err := f.Init(); err != nil {
		t.Fatal(err)
	}

	if err := f.Start("wtf"); err != nil {
		t.Fatal(err)
	}

	checkTitle("Checking data export")
	t.Log(f.dataFileHandler)
	fd, err := f.dataFileHandler.Stat()
	if err != nil {
		t.Fatal(err)
	}

	if f.LogsData() && (fd.Name() == "netspot_wtf_data.json") {
		testOK()
	} else {
		testERROR()
		t.Errorf("Expecting %s, got %s",
			"netspot_wtf_data.json",
			fd.Name())
	}

	checkTitle("Checking alarm export")
	fd, _ = f.alarmFileHandler.Stat()

	if f.LogsAlarm() && (fd.Name() == "netspot_wtf_alarm.json") {
		testOK()
	} else {
		testERROR()
		t.Errorf("Expecting %s, got %s",
			"netspot_wtf_alarm.json",
			fd.Name())
	}

	if err = f.Close(); err != nil {
		t.Error(err)
	}
}

func TestFileWriteAndWarn(t *testing.T) {
	title("Testing data writing (File exporter)")

	if err := setFullConfig(); err != nil {
		t.Error(err)
	}

	f := File{}
	if err := f.Init(); err != nil {
		t.Fatal(err)
	}

	if err := f.Start("wtf"); err != nil {
		t.Fatal(err)
	}

	// prepare data
	now := time.Now()
	data := map[string]float64{"stat0": 15.2, "stat1": -3.33333333}

	f.Write(now, data)

	// prepare data
	alert := SpotAlert{
		Status:      "UP_ALERT",
		Stat:        "R_SYN",
		Value:       0.995,
		Code:        1,
		Probability: 1e-8,
	}
	f.Warn(now, &alert)

	f.Close()

	d, err := ioutil.ReadFile(fmt.Sprintf(f.dataAddress, f.seriesName))
	if err != nil {
		t.Error(err)
	}
	t.Log(string(d))

	a, err := ioutil.ReadFile(fmt.Sprintf(f.alarmAddress, f.seriesName))
	if err != nil {
		t.Error(err)
	}
	t.Log(string(a))
}

func TestFileStatus(t *testing.T) {
	if err := setFullConfig(); err != nil {
		t.Error(err)
	}

	// config
	configExample := []byte(`
[exporter.file]
# A boolean to activate console data logging
data = "/tmp/netspot.data"
# A boolean to activate console alarm logging
alarm = "/tmp/netspot.alarm"
`)

	if err := config.LoadForTestRawToml(configExample); err != nil {
		t.Fatal(err)
	}

	f := File{}
	if err := f.Init(); err != nil {
		t.Fatal(err)
	}

	t.Logf("%v+\n", f.Status())

	f.Start("noformat")
	f.Close()
}
