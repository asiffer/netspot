// test_miner.go
package miner

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

var (
	headerWidth    = 80
	pcapTestFile   = "test/test.pcap"
	configTestFile = "test/miner.toml"
)

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

func init() {
	DisableLogging()
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

func TestSetDevice(t *testing.T) {
	title("Testing device setting")
	checkTitle(fmt.Sprintf("Setting device to %s... ", pcapTestFile))
	r := SetDevice(pcapTestFile)
	if r != 0 || GetDevice() != pcapTestFile {
		testERROR()
		t.Error("Fail: setting device (file)")
	}
	if IsDeviceInterface() {
		testERROR()
		t.Error("Fail: setting device (file)")
	} else {
		testOK()
	}

	dev := AvailableDevices[0]
	checkTitle(fmt.Sprintf("Setting device to %s... ", dev))
	r = SetDevice(dev)
	if r != 0 || GetDevice() != dev {
		testERROR()
		t.Error("Fail: setting device (interface)")
	}
	if !IsDeviceInterface() {
		testERROR()
		t.Error("Fail: setting device (interface)")
	} else {
		testOK()
	}
}

func TestGetNumberOfDevice(t *testing.T) {
	title("Testing getting the number of devices")
	checkTitle("Getting the right number of devices...")
	if GetNumberOfDevices() != len(AvailableDevices) {
		testERROR()
		t.Error("Fail: getting the number of devices")
	} else {
		testOK()
	}
}

func TestSetTimeout(t *testing.T) {
	title("Testing timeout setting")
	to := 500 * time.Millisecond
	checkTitle(fmt.Sprintf("Setting timeout to %s... ", to))
	SetTimeout(to)
	if timeout == to {
		testOK()
	} else {
		testERROR()
		t.Error("Fail: setting timeout")
	}
}

func TestSetSnapshotLen(t *testing.T) {
	title("Testing snapshot length setting")
	var sl int32 = 4096
	checkTitle(fmt.Sprintf("Setting snapshot length to %d... ", sl))
	SetSnapshotLen(sl)
	if snapshotLen == sl {
		testOK()
	} else {
		testERROR()
		t.Error("Fail: setting snapshot length")
	}
}

func TestSetPromiscuous(t *testing.T) {
	title("Testing promiscuous mode setting")
	checkTitle(fmt.Sprintf("Setting promiscuous to %v... ", true))
	SetPromiscuous(true)
	if IsPromiscuous() {
		testOK()
	} else {
		testERROR()
		t.Error("Fail: setting promiscuous")
	}
}

func TestInitConfig(t *testing.T) {
	title("Loading config file")
	f, e := filepath.Abs(configTestFile)
	if e != nil {
		t.Error("Fail: retrieving config file")
	}
	viper.SetConfigFile(f)
	viper.ReadInConfig()
	InitConfig()
	checkTapConfig(t)
}

func checkTapConfig(t *testing.T) {
	checkTitle("Checking device...")
	if GetDevice() != "test/test.pcap" {
		testERROR()
		fmt.Println(GetDevice())
		t.Error("Fail: setting device (file)")
	} else {
		testOK()
	}

	checkTitle("Checking snapshot length...")
	if snapshotLen != 9999 {
		testERROR()
		t.Error("Fail: setting snapshot length")
	} else {
		testOK()
	}

	checkTitle("Checking timeout...")
	if timeout != 7*time.Second {
		testERROR()
		t.Error("Fail: setting timeout")
	} else {
		testOK()
	}

	checkTitle("Checking promiscuous...")
	if !IsPromiscuous() {
		testERROR()
		t.Error("Fail: setting promiscuous")
	} else {
		testOK()
	}
}

func TestRawStatus(t *testing.T) {
	title("Testing raw status")
	viper.SetConfigFile(configTestFile)
	viper.ReadInConfig()
	InitConfig()
	m := RawStatus()

	checkTitle("Checking device...")
	if m["device"] != "test/test.pcap" {
		testERROR()
		fmt.Println(GetDevice())
		t.Error("Fail: setting device (file)")
	} else {
		testOK()
	}

	checkTitle("Checking snapshot length...")
	if m["snapshot length"] != "9999" {
		testERROR()
		t.Error("Fail: setting snapshot length")
	} else {
		testOK()
	}
	checkTitle("Checking timeout...")
	if m["timeout"] != "7s" {
		testERROR()
		t.Error("Fail: setting timeout")
	} else {
		testOK()
	}

	checkTitle("Checking promiscuous...")
	if m["promiscuous"] != "true" {
		testERROR()
		t.Error("Fail: setting promiscuous")
	} else {
		testOK()
	}
}

func TestPcapSniffing(t *testing.T) {
	title("Testing pcap sniffing")
	setNormalConfig()
	StartSniffingAndWait()

	checkTitle("Checking number of parsed packets... ")
	if GetNbParsedPkts() != 908 {
		testERROR()
		t.Error("Fail: getting number of parsed packets")
	} else {
		testOK()
	}
}

func pcapSniffing() {
	SetDevice(pcapTestFile)
	SetTimeout(30 * time.Second)
	SetPromiscuous(false)
	SetSnapshotLen(65537)
	StartSniffingAndWait()
}

func setNormalConfig() {
	SetDevice(pcapTestFile)
	SetTimeout(1 * time.Second)
	SetPromiscuous(false)
	SetSnapshotLen(65537)
}

func TestZero(t *testing.T) {
	title("Testing reset (Zero)")

	viper.SetConfigFile(configTestFile)
	viper.ReadInConfig()
	InitConfig()
	pcapSniffing()

	Zero()
	checkTapConfig(t)

	checkTitle("Checking number of parsed packets...")
	if nbParsedPkts != 0 {
		testERROR()
		t.Error("Fail: resetting number of parsed packets")
	} else {
		testOK()
	}

	checkTitle("Checking counter map and counter id...")
	if counterID != 0 {
		testERROR()
		t.Error("Fail: resetting counter id")
	}
	if len(counterMap) > 0 {
		testERROR()
		t.Error("Fail: resetting counter map")
	} else {
		testOK()
	}

	checkTitle("Checking status variables...")
	if sendTicks {
		testERROR()
		t.Error("Fail: resetting sendTicks")
	}
	if sniffing {
		testERROR()
		t.Error("Fail: resetting sniffing")
	} else {
		testOK()
	}

}

func TestUnloading(t *testing.T) {
	title("Testing counter unloading")
	LoadFromName("IP")
	l := GetLoadedCounters()
	if len(l) > 0 {
		checkTitle("Checking counter unloading...")
		ctrname := l[0]
		if UnloadFromName(ctrname) != 0 {
			testERROR()
			t.Error("Fail: unloading counter")
		} else {
			testOK()
		}

		checkTitle("Checking if counter remains...")
		for _, n := range GetLoadedCounters() {
			if n == ctrname {
				testERROR()
				t.Error("Fail: the counter remains")
			}
		}
		testOK()

	}
}

func TestLoading(t *testing.T) {
	title("Testing counter loading")
	ok := LoadFromName("SYN")
	nok := LoadFromName("WTF")

	checkTitle("Checking existing counter loading...")
	if ok < 0 {
		testERROR()
		t.Error("Fail: counter not loaded")
	} else {
		testOK()
	}

	checkTitle("Checking not existing counter loading...")
	if nok >= 0 {
		testERROR()
		t.Error("Fail: counter not loaded")
	} else {
		testOK()
	}
}

func TestGettingCounterValue(t *testing.T) {
	title("Testing counter values")
	Zero()
	setNormalConfig()
	idIP := LoadFromName("IP")
	idSYN := LoadFromName("SYN")

	StartSniffing()
	time.Sleep(1 * time.Second)
	checkTitle("Checking IP counter value...")
	if GetCounterValue(idIP) != 908 {
		testERROR()
		t.Errorf("Fail: bad counter (get %d instead of 907)", GetCounterValue(idIP))
	} else {
		testOK()
	}

	checkTitle("Checking SYN counter value...")
	if GetCounterValue(idSYN) != 56 {
		testERROR()
		t.Errorf("Fail: bad counter (get %d instead of 56)", GetCounterValue(idSYN))
	} else {
		testOK()
	}

}
