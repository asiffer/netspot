// test_miner.go

package miner

import (
	"fmt"
	"netspot/miner/counters"
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
	UnloadAll()
	ok := LoadFromName("SYN")
	nok := LoadFromName("WTF")

	checkTitle("Checking number of loaded counters...")
	n := GetNumberOfLoadedCounters()
	if n != 1 {
		testERROR()
		t.Errorf("Fail: Expected 1 counter, got %d", n)
	} else {
		testOK()
	}

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

	checkTitle("Checking loading twice...")
	syn := counters.AvailableCounters["SYN"]()
	_, err := load(syn)
	if err == nil {
		testERROR()
		t.Error("Fail: counter wrongly loaded")
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

	StartSniffingAndWait()
	time.Sleep(2 * time.Second)
	checkTitle("Checking IP counter value...")
	ipCtrValue, _ := GetCounterValue(idIP)
	if ipCtrValue != 908 {
		testERROR()
		t.Errorf("Fail: bad counter (get %d instead of 907)", ipCtrValue)
	} else {
		testOK()
	}

	checkTitle("Checking SYN counter value...")
	synCtrValue, _ := GetCounterValue(idSYN)
	if synCtrValue != 56 {
		testERROR()
		t.Errorf("Fail: bad counter (get %d instead of 56)", synCtrValue)
	} else {
		testOK()
	}

	checkTitle("Checking unloaded counter value...")
	_, err := GetCounterValue(-1)
	if err == nil {
		testERROR()
		t.Error("Fail: bad counter")
	} else {
		testOK()
	}

}

func TestStartStop(t *testing.T) {
	title("Testing start/stop")
	Zero()
	setNormalConfig()
	StartSniffing()
	time.Sleep(1 * time.Millisecond)
	StopSniffing()
	time.Sleep(10 * time.Millisecond)

	checkTitle("Checking sniffing status...")
	if IsSniffing() {
		testERROR()
		t.Errorf("Fail: sniffing has not stopped")
	}
	testOK()
}

func TestResetCounter(t *testing.T) {
	title("Testing counter reset")
	Zero()
	setNormalConfig()
	// SetLogging(1)
	idIP := LoadFromName("IP")
	idSYN := LoadFromName("SYN")

	StartSniffingAndWait()

	time.Sleep(1 * time.Second)
	Reset(idSYN)
	Reset(idIP)
	checkTitle("Checking IP counter reset...")
	ipCtrValue, _ := GetCounterValue(idIP)
	if ipCtrValue != 0 {
		testERROR()
		t.Error("Fail: resetting IP counter")
	} else {
		testOK()
	}

	checkTitle("Checking SYN counter reset...")
	synCtrValue, _ := GetCounterValue(idSYN)
	if synCtrValue != 0 {
		testERROR()
		t.Error("Fail: resetting SYN counter")
	} else {
		testOK()
	}

	checkTitle("Checking bad counter reset...")
	if Reset(-1) != -1 {
		testERROR()
		t.Error("Fail: resetting bad counter")
	} else {
		testOK()
	}
}

func TestResetAllCounter(t *testing.T) {
	title("Testing resetting all counters")
	Zero()
	setNormalConfig()
	// SetLogging(1)
	idIP := LoadFromName("IP")
	idSYN := LoadFromName("SYN")

	StartSniffingAndWait()
	time.Sleep(1 * time.Second)
	ResetAll()
	checkTitle("Checking IP counter reset...")
	ipCtrValue, _ := GetCounterValue(idIP)
	if ipCtrValue != 0 {
		testERROR()
		t.Error("Fail: resetting IP counter")
	} else {
		testOK()
	}

	checkTitle("Checking SYN counter reset...")
	synCtrValue, _ := GetCounterValue(idSYN)
	if synCtrValue != 0 {
		testERROR()
		t.Error("Fail: resetting SYN counter")
	} else {
		testOK()
	}
}

func TestStartingSingleCounter(t *testing.T) {
	title("Testing starting a single counter")
	Zero()
	id := LoadFromName("ACK")
	err := startCounter(id)

	checkTitle("Checking IP counter start...")
	if err != nil {
		testERROR()
		t.Error(err.Error())
	} else {
		testOK()
	}

	checkTitle("Checking second IP start...")
	if startCounter(id) == nil {
		testERROR()
		t.Error("Counter has wrongly started")
	} else {
		testOK()
	}

	checkTitle("Checking not loaded counter start...")
	if startCounter(-1) == nil {
		testERROR()
		t.Error("Counter has wrongly started")
	} else {
		testOK()
	}
	time.Sleep(200 * time.Millisecond)
	stopAllCounters()
}

func TestTick(t *testing.T) {
	title("Testing tick")
	Zero()

	setNormalConfig()
	LoadFromName("IP")
	LoadFromName("SYN")

	chant := Tick(500 * time.Microsecond)
	to := time.Tick(2 * time.Second)
	nticks := 0
	StartSniffing()
	checkTitle("Correct number of ticks...")
	for {
		select {
		case <-chant:
			nticks++
		case <-to:
			if nticks == 406 {
				testOK()
				return
			}
			t.Errorf("Expected 406 ticks, got %d", nticks)
			testERROR()
			return

		}
		// fmt.Println(nticks)
	}

}

func TestCAPI(t *testing.T) {
	title("Testing C API")
	checkTitle("Available devices...")
	if testCGetAvailableDevices() {
		testOK()
	} else {
		testERROR()
	}

	checkTitle("Get device...")
	if testCGetDevice() {
		testOK()
	} else {
		testERROR()
	}

	checkTitle("Set device...")
	if testCSetDevice() {
		testOK()
	} else {
		testERROR()
	}

	checkTitle("Set timeout...")
	if testCSetTimeout() {
		testOK()
	} else {
		testERROR()
	}

	checkTitle("Get loaded counter...")
	if testCGetLoadedCounters() {
		testOK()
	} else {
		testERROR()
	}

	checkTitle("Load counter...")
	if testCLoadFromName() {
		testOK()
	} else {
		testERROR()
	}

	checkTitle("Unload counter...")
	if testCUnloadFromName() {
		testOK()
	} else {
		testERROR()
	}

	Zero()
	i := LoadFromName("SYN")
	pcapSniffing()

	checkTitle("Geting counter value...")
	if testCGetCounterValue(i, 56) {
		testOK()
	} else {
		testERROR()
	}
}
