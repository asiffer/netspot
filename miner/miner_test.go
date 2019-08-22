// test_miner.go

package miner

import (
	"encoding/gob"
	"fmt"
	"net"
	"netspot/miner/counters"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

var (
	headerWidth    = 80
	pcapTestFile   = "test/test.pcap"
	pcapTestFile2  = "test/wifi.pcap"
	pcapTestFile3  = "/data/pcap/4SICS-GeekLounge-151020.pcap"
	pcapTestFile4  = "/data/pcap/201111111400.dump"
	pcapTestFile5  = "/data/kitsune/Mirai/Mirai_pcap.pcap"
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

func testWARNING() {
	fmt.Println("[\033[32mWARNING\033[0m]")
}

func init() {
	InitLogger()
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

func TestGetAvailableCounters(t *testing.T) {
	title("Available counters")
	fmt.Println(GetAvailableCounters())
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

	//dev := AvailableDevices[0]
	dev := "any"
	checkTitle(fmt.Sprintf("Setting device to %s... ", dev))
	r = SetDevice(dev)
	if r != 0 || GetDevice() != dev {
		testERROR()
		t.Error("Fail: setting device (interface)")
		// }
		// if !IsDeviceInterface() {
		// 	testERROR()
		// 	t.Error("Fail: setting device (interface)")
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
		t.Errorf("Fail: setting device (expected test.pcap, got %s)", GetDevice())
	} else {
		testOK()
	}

	checkTitle("Checking snapshot length...")
	if snapshotLen != 9999 {
		testERROR()
		t.Errorf("Fail: setting snapshot length (expected 9999, got %d)", snapshotLen)
	} else {
		testOK()
	}

	checkTitle("Checking timeout...")
	if timeout != 7*time.Second {
		testERROR()
		t.Errorf("Fail: setting timeout (expected 7s, got %v)", timeout)
	} else {
		testOK()
	}

	checkTitle("Checking promiscuous...")
	if !IsPromiscuous() {
		testERROR()
		t.Errorf("Fail: setting promiscuous (expected true, got %t)", IsPromiscuous())
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
	if m["snapshot_length"] != "9999" {
		testERROR()
		t.Errorf("Fail: setting snapshot length (expected 9999, got %s)", m["snapshot length"])
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
	// SetLogging(0)
	setNormalConfig()
	LoadFromName("TIME")
	LoadFromName("PKTS")

	StartSniffingAndWait()
	// time.Sleep(1 * time.Second)

	checkTitle("Checking number of parsed packets... ")
	// fmt.Println("WTF")
	if n, err := GetNbParsedPackets(); n != 908 || err != nil {
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
	InitConfig()

	Zero()
	checkTapConfig(t)

	// checkTitle("Checking number of parsed packets...")
	// if n, err := GetNbParsedPackets(); n != 0 || err != nil {
	// 	testERROR()
	// 	t.Error("Fail: resetting number of parsed packets")
	// } else {
	// 	testOK()
	// }

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

	time.Sleep(250 * time.Millisecond)
	StartSniffingAndWait()
	time.Sleep(1 * time.Second)
	checkTitle("Checking IP counter value...")
	ipCtrValue, _ := GetCounterValue(idIP)
	if ipCtrValue < 880 {
		testERROR()
		t.Errorf("Fail: bad counter (get %d instead of 908)", ipCtrValue)
	} else if ipCtrValue != 908 {
		testWARNING()
	} else {
		testOK()
	}
	// if ipCtrValue != 908 {
	// 	testERROR()
	// 	t.Errorf("Fail: bad counter (get %d instead of 908)", ipCtrValue)
	// } else {
	// 	testOK()
	// }
	// DisableLogging()

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

	// chant := Tick(500 * time.Microsecond)
	// SetLogging(1)
	// SetTickPeriod(500 * time.Microsecond)
	// fmt.Println("SEND TICKS:", sendTicks)
	// to := time.Tick(2 * time.Second)
	nticks := 0
	// StartSniffing()
	// SetLogging(0)
	_, data := GoSniffAndYieldChannel(500 * time.Microsecond)
	checkTitle("Correct number of ticks...")
	for m := range data {
		if m != nil {
			nticks++
		}
	}

	if nticks == 405 {
		testOK()
		return
	}
	t.Errorf("Expected 405 ticks, got %d", nticks)
	testERROR()
	return

	// for {
	// 	select {
	// 	case <-tc:
	// 		nticks++
	// 	case <-to:
	// 		if nticks == 406 {
	// 			testOK()
	// 			return
	// 		}
	// 		t.Errorf("Expected 406 ticks, got %d", nticks)
	// 		testERROR()
	// 		return

	// 	}
	// fmt.Println(nticks)
	// }

}

// func TestCAPI(t *testing.T) {
// 	title("Testing C API")
// 	// checkTitle("Available devices...")
// 	// if testCGetAvailableDevices() {
// 	// 	testOK()
// 	// } else {
// 	// 	testERROR()
// 	// }

// 	checkTitle("Get device...")
// 	if testCGetDevice() {
// 		testOK()
// 	} else {
// 		testERROR()
// 	}

// 	// checkTitle("Set device...")
// 	// if testCSetDevice() {
// 	// 	testOK()
// 	// } else {
// 	// 	testERROR()
// 	// }

// 	checkTitle("Set timeout...")
// 	if testCSetTimeout() {
// 		testOK()
// 	} else {
// 		testERROR()
// 	}

// 	checkTitle("Get loaded counter...")
// 	if testCGetLoadedCounters() {
// 		testOK()
// 	} else {
// 		testERROR()
// 	}

// 	checkTitle("Load counter...")
// 	if testCLoadFromName() {
// 		testOK()
// 	} else {
// 		testERROR()
// 	}

// 	checkTitle("Unload counter...")
// 	if testCUnloadFromName() {
// 		testOK()
// 	} else {
// 		testERROR()
// 	}

// 	Zero()
// 	i := LoadFromName("SYN")
// 	pcapSniffing()

// 	checkTitle("Geting counter value...")
// 	if testCGetCounterValue(i, 56) {
// 		testOK()
// 	} else {
// 		testERROR()
// 	}
// }

func TestLoadPattern(t *testing.T) {
	title("Checking pattern loading")
	p := "192.168.200.166: -> :"

	checkTitle("Loading pattern '" + p + " ...")
	id, err := LoadPattern(p, "P0")
	if err != nil {
		testERROR()
		t.Fatal(err.Error())
	} else {
		testOK()
	}

	checkTitle("Checking counter value...")
	pcapSniffing()
	if val := counterMap[id].Value(); val != 441 {
		testERROR()
		t.Errorf("Bad counter value, expected 441, got %d", val)
	} else {
		testOK()
	}

	Zero()
	p = "192.168.200.166: -> :443"
	checkTitle("Loading pattern '" + p + " ...")
	id, err = LoadPattern(p, "P0")
	if err != nil {
		testERROR()
		t.Fatal(err.Error())
	} else {
		testOK()
	}

	checkTitle("Checking counter value...")
	pcapSniffing()
	if val := counterMap[id].Value(); val != 366 {
		testERROR()
		t.Errorf("Bad counter value, expected 366, got %d", val)
	} else {
		testOK()
	}

	Zero()
	p = ": -> :80"
	checkTitle("Loading pattern '" + p + " ...")
	id, err = LoadPattern(p, "P0")
	if err != nil {
		testERROR()
		t.Fatal(err.Error())
	} else {
		testOK()
	}

	checkTitle("Checking counter value...")
	pcapSniffing()
	if val := counterMap[id].Value(); val != 20 {
		testERROR()
		t.Errorf("Bad counter value, expected 20, got %d", val)
	} else {
		testOK()
	}

	Zero()
	p = "192.168.200.0/24: -> :"
	checkTitle("Loading pattern '" + p + " ...")
	id, err = LoadPattern(p, "P0")
	if err != nil {
		testERROR()
		t.Fatal(err.Error())
	} else {
		testOK()
	}

	checkTitle("Checking counter value...")
	pcapSniffing()
	if val := counterMap[id].Value(); val != 478 {
		testERROR()
		t.Errorf("Bad counter value, expected 20, got %d", val)
	} else {
		testOK()
	}

	p = ": -> :80"
	checkTitle("Loading pattern with a name already in used")
	id, err = LoadPattern(p, "P0")
	if err == nil {
		testERROR()
		t.Fatal(err.Error())
	} else {
		testOK()
	}

	p = "!µ*:%%+->é:?&"
	checkTitle("Loading a ugly pattern like '" + p + "'...")
	id, err = LoadPattern(p, "Perr")
	if err != nil {
		testERROR()
		t.Fatal(err.Error())
	} else {
		testOK()
	}
}

func TestSnapshot(t *testing.T) {
	title("Checking snapshots")
	Zero()
	SetDevice(pcapTestFile2)
	LoadFromName("ACK")
	LoadFromName("UDP")
	LoadFromName("IP")
	LoadFromName("SYN")

	os.Remove("test/miner_test.socket")
	addr := net.UnixAddr{
		Name: "test/miner_test.socket",
		Net:  "unixgram",
	}

	conn, err := net.ListenUnixgram("unixgram", &addr)
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	checkTitle("Decoding the incoming map...")
	decoder := gob.NewDecoder(conn)
	go SniffAndSendUnix(5*time.Millisecond, false, addr.Name)
	time.Sleep(100 * time.Millisecond)
	var a map[int]uint64
	err = decoder.Decode(&a)
	if err != nil {
		t.Error(err)
	} else {
		testOK()
	}
	// wait to end the sniffing
	time.Sleep(1 * time.Second)
	// fmt.Println("Sniffing:", IsSniffing())
}

func TestNbParsedPackets(t *testing.T) {
	title("Checking live perf")
	// SetLogging(0)
	Zero()
	SetDevice("/data/pcap/201111111400.dump")
	LoadFromName("ACK")
	LoadFromName("UDP")
	LoadFromName("IP")
	LoadFromName("SYN")
	time.Sleep(50 * time.Millisecond)
	StartSniffing()
	start := time.Now()
	time.Sleep(3 * time.Second)
	end := time.Now()
	duration := end.Sub(start).Seconds()

	if pp, err := GetNbParsedPackets(); err == nil {
		fmt.Printf("#packets: %d\n", pp)
		fmt.Printf("Packet rate: %d packets/s\n", int(float64(pp)/duration))
	}

	StopSniffing()
}

func TestLiveFlush(t *testing.T) {
	title("Checking live flush")
	// SetLogging(0)
	Zero()
	SetDevice("/data/pcap/201111111400.dump")
	LoadFromName("ACK")
	LoadFromName("UDP")
	LoadFromName("IP")
	LoadFromName("SYN")
	time.Sleep(50 * time.Millisecond)
	event, data := StartSniffing()
	// for i := 0; i < 5; i++ {
	time.Sleep(50 * time.Millisecond)
	event <- FLUSH
	fmt.Println(<-data)
	event <- GET
	fmt.Println(<-data)
	StopSniffing()
	time.Sleep(100 * time.Millisecond)
}

func TestBasicIOSniff(t *testing.T) {
	title("Checking io while sniffing")
	// SetLogging(0)
	Zero()
	SetDevice("/data/pcap/201111111400.dump")
	LoadFromName("ACK")
	LoadFromName("UDP")
	LoadFromName("IP")
	LoadFromName("SYN")
	time.Sleep(50 * time.Millisecond)
	go Sniff()

	time.Sleep(500 * time.Millisecond)
	defaultEventChannel <- GET
	fmt.Println("Data:", <-defaultDataChannel)

	time.Sleep(500 * time.Millisecond)
	defaultEventChannel <- FLUSH
	fmt.Println("Data:", <-defaultDataChannel)

	// time.Sleep(500 * time.Millisecond)
	// defaultEventChannel <- PERF
	// fmt.Println("Perf:", <-defaultDataChannel)

	// time.Sleep(500 * time.Millisecond)
	// defaultEventChannel <- TIME
	// fmt.Println("Time:", <-defaultDataChannel)

	time.Sleep(500 * time.Millisecond)
	defaultEventChannel <- STOP
	time.Sleep(1 * time.Second)
	// fmt.Println("Perf:", <-defaultDataChannel)
	// StopSniffing()
}

func TestBasicIOSniff2(t *testing.T) {
	title("Checking io while sniffing (GoSniff)")
	// SetLogging(0)
	Zero()
	SetDevice(pcapTestFile4)
	LoadFromName("ACK")
	LoadFromName("UDP")
	LoadFromName("IP")
	LoadFromName("SYN")
	time.Sleep(50 * time.Millisecond)
	event, data := GoSniff()

	time.Sleep(500 * time.Millisecond)
	event <- GET
	fmt.Println("Data:", <-data)

	time.Sleep(500 * time.Millisecond)
	event <- FLUSH
	fmt.Println("Data:", <-data)

	// time.Sleep(500 * time.Millisecond)
	// event <- PERF
	// fmt.Println("Perf:", <-data)

	// time.Sleep(500 * time.Millisecond)
	// event <- TIME
	// fmt.Println("Time:", <-data)

	time.Sleep(500 * time.Millisecond)
	event <- STOP
	time.Sleep(1 * time.Second)
	// fmt.Println("Perf:", <-defaultDataChannel)
	// StopSniffing()
}

func TestBasicIOSniff3(t *testing.T) {
	title("Checking io while sniffing (GoSniffAndYieldChannel)")
	// SetLogging(0)
	Zero()
	SetDevice(pcapTestFile4)
	LoadFromName("ACK")
	LoadFromName("UDP")
	LoadFromName("TCP")
	LoadFromName("SYN")
	time.Sleep(50 * time.Millisecond)
	event, data := GoSniffAndYieldChannel(1 * time.Hour)

	time.Sleep(500 * time.Millisecond)
	event <- GET
	fmt.Println("Data:", <-data)

	time.Sleep(500 * time.Millisecond)
	event <- FLUSH
	fmt.Println("Data:", <-data)

	// time.Sleep(500 * time.Millisecond)
	// event <- PERF
	// fmt.Println("Perf:", <-data)

	// time.Sleep(500 * time.Millisecond)
	// event <- TIME
	// fmt.Println("Time:", <-data)

	time.Sleep(500 * time.Millisecond)
	event <- STOP
	time.Sleep(1 * time.Second)
}
