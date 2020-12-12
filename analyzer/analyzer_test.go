// analyezr_test.go

package analyzer

import (
	"fmt"
	"io/ioutil"
	"netspot/config"
	"netspot/exporter"
	"netspot/miner"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

var (
	HeaderWidth = 100
	HeaderSym   = "-"
	pcapFile1   = "/data/pcap/snort.log.1425823194.pcap"
	pcapFile2   = "/data/pcap/202002071400.pcap"
	pcapFile3   = "/data/kitsune/Mirai/Mirai_pcap.pcap"
)

var (
	testFiles []string
	wd        string
)

func title(s string) {
	l := len(s)
	var border int
	var left string
	var right string
	remaining := HeaderWidth - l - 2
	if remaining%2 == 0 {
		border = remaining / 2
		left = strings.Repeat(HeaderSym, border) + " "
		right = " " + strings.Repeat(HeaderSym, border)
	} else {
		border = (remaining - 1) / 2
		left = strings.Repeat(HeaderSym, border+1) + " "
		right = " " + strings.Repeat(HeaderSym, border)
	}

	fmt.Println(left + s + right)
}

func checkTitle(s string) {
	format := "%-" + fmt.Sprint(HeaderWidth-7) + "s"
	fmt.Printf(format, s)
}

func findTestFiles() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	testDir := path.Join(wd, "../test")
	infos, err := ioutil.ReadDir(testDir)
	if err != nil {
		return err
	}
	testFiles = make([]string, len(infos))
	for i, f := range infos {
		testFiles[i] = path.Join(testDir, f.Name())
	}
	return nil
}

func init() {
	// load test files
	if err := findTestFiles(); err != nil {
		panic(err)
	}
	config.LoadDefaults()

	miner.InitLogger()
	if err := miner.InitConfig(); err != nil {
		panic(err)
	}
	InitLogger()

}

func testOK() {
	fmt.Println("[\033[32mOK\033[0m]")
}

func testERROR() {
	fmt.Println("[\033[31mERROR\033[0m]")
}

func TestLoadStat(t *testing.T) {
	title(t.Name())

	dir, _ := os.Getwd()
	fmt.Println("WD:", dir)

	UnloadAll()

	checkTitle("Checking available stats...")
	if len(GetAvailableStats()) >= 5 {
		testOK()
	} else {
		testERROR()
		t.Errorf("Expected at least 5 stats, got %d", len(GetAvailableStats()))
	}

	checkTitle("Loading R_SYN...")
	if err := LoadFromName("R_SYN"); err != nil {
		t.Errorf("Error while loading R_SYN: %v", err)
		testERROR()
	} else {
		testOK()
	}

	checkTitle("Reloading R_SYN...")
	if err := LoadFromName("R_SYN"); err == nil {
		t.Errorf("Error while re-loading R_SYN: %v", err)
		testERROR()
	} else {
		testOK()
	}

	checkTitle("Checking number of loaded stats...")
	if len(GetLoadedStats()) != 1 {
		t.Errorf("Bad number of loaded counters (expected 1, got %d)", len(GetLoadedStats()))
		testERROR()
	} else {
		testOK()
	}
	UnloadAll()
}

func TestUnloadAll(t *testing.T) {
	title(t.Name())
	LoadFromName("R_SYN")
	LoadFromName("R_ICMP")
	LoadFromName("R_DST_SRC")
	fmt.Println("Loaded stats:", GetLoadedStats())
	// fmt.Println("Loaded counters:", miner.GetLoadedCounters())
	// fmt.Println("Removing all stats ...")
	UnloadAll()
	if len(GetLoadedStats()) > 0 {
		t.Error("Error while removing all stats")
	}
	if len(miner.GetLoadedCounters()) > 0 {
		t.Error("Error while removing underlying counters")
	}
	fmt.Println("Loaded stats:", GetLoadedStats())
	fmt.Println("Loaded counters:", miner.GetLoadedCounters())
}

func TestUnloadSpecific(t *testing.T) {
	title("Testing specific unloading")
	LoadFromName("R_SYN")
	LoadFromName("R_ICMP")
	fmt.Println("Loaded stats:", GetLoadedStats())
	// fmt.Println("Loaded counters:", miner.GetLoadedCounters())
	// fmt.Println("Removing R_SYN")
	UnloadFromName("R_SYN")
	if find(GetLoadedStats(), "R_SYN") > 0 {
		t.Error("Error while removing R_SYN")
	}
	if find(miner.GetLoadedCounters(), "SYN") > 0 {
		t.Error("Error while removing SYN counter")
	}
	fmt.Println("Loaded stats:", GetLoadedStats())
	// fmt.Println("Loaded counters:", miner.GetLoadedCounters())
}

func TestZero(t *testing.T) {
	title(t.Name())
	// SetLogging(1)
	UnloadAll()
	LoadFromName("R_SYN")
	LoadFromName("AVG_PKT_SIZE")
	LoadFromName("R_ACK")
	LoadFromName("R_DST_SRC")
	LoadFromName("R_ICMP")

	// small
	// pcapFile1 : ~420min

	if err := miner.SetDevice(testFiles[0]); err != nil {
		t.Fatal(err)
	}
	fmt.Println("AFTER:", miner.GetDevice())
	period = 5 * time.Minute

	// huge
	// pcapFile2 : 900s
	// miner.SetDevice(pcapFile2)
	// period = 200 * time.Millisecond

	time.Sleep(1 * time.Second)
	// miner.StartSniffing()
	a := time.Now()
	if err := StartAndWait(); err != nil {
		t.Error(err)
	}
	b := time.Since(a)
	fmt.Printf("Timing: %f\n", b.Seconds())
	// fmt.Println("Reset")
	// miner.Zero()
	Zero()
}
func TestLivePcapSmall(t *testing.T) {
	// SetLogging(0)
	// zerolog.SetGlobalLevel(zerolog.DebugLevel)
	config.Clean()
	config.LoadDefaults()
	exporter.Load("console")
	title(t.Name())
	// reset()
	// miner.Zero()
	UnloadAll()
	LoadFromName("R_SYN")
	LoadFromName("AVG_PKT_SIZE")
	LoadFromName("R_ACK")
	LoadFromName("R_DST_SRC")
	LoadFromName("R_ICMP")

	// small
	// pcapFile1 : ~420min
	// logDataToFile = true
	miner.SetDevice(testFiles[1])
	SetPeriod(5 * time.Minute)

	// miner.SetTickPeriod(period)
	// huge
	// pcapFile2 : 900s
	// miner.SetDevice(pcapFile2)
	// period = 200 * time.Millisecond

	// miner.StartSniffing()
	// if !miner.IsSniffing() {
	// 	t.Error("Error: no sniffing")
	// }
	time.Sleep(1 * time.Second)
	// a := time.Now()
	if err := StartAndWait(); err != nil {
		t.Errorf("Error while running: %v", err)
	}
	// b := time.Since(a)
	// log.Printf("Timing: %f", b.Seconds())
}

func TestLivePcapHuge(t *testing.T) {
	// DisableLogging()
	title(t.Name())
	UnloadAll()
	// LoadFromName("R_SYN")
	// LoadFromName("AVG_PKT_SIZE")
	LoadFromName("R_ACK")
	// LoadFromName("R_DST_SRC")
	// LoadFromName("R_ICMP")

	extra := map[string]interface{}{
		"spot.AVG_PKT_SIZE.n_init": 800,
		"spot.AVG_PKT_SIZE.level":  0.99,
		"spot.AVG_PKT_SIZE.q":      1e-3,
	}
	config.LoadForTest(extra)
	// LoadFromNameWithCustomConfig("AVG_PKT_SIZE", extra)

	// small
	// pcapFile1 : ~420min
	// logDataToFile = true
	// miner.SetDevice(pcapFile1)
	// period = 5 * time.Minute

	// huge
	// pcapFile2 : 900s
	miner.SetDevice(pcapFile2)
	period = 200 * time.Millisecond
	SetPeriod(period)
	// miner.SetTickPeriod(period)
	// miner.StartSniffing()
	// if !miner.IsSniffing() {
	// 	t.Error("Error: no sniffing")
	// }
	if err := Start(); err != nil {
		t.Error(err)
	}
	time.Sleep(4 * time.Second)
	if err := Stop(); err != nil {
		t.Error(err)
	}
	time.Sleep(1 * time.Second)
}

func TestLivePerfs(t *testing.T) {
	title(t.Name())
	// UnloadAll()
	Zero()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	exporter.Load("console")
	if err := exporter.Start("testAnalyzer"); err != nil {
		t.Errorf("Error while starting exporter")
	}
	extra := map[string]interface{}{
		"spot.AVG_PKT_SIZE.n_init": 800,
		"spot.AVG_PKT_SIZE.level":  0.98,
		"spot.AVG_PKT_SIZE.q":      1e-3,
		"spot.R_SYN.n_init":        800,
		"spot.R_SYN.level":         0.98,
		"spot.R_SYN.q":             1e-3,
		"spot.R_ACK.n_init":        800,
		"spot.R_ACK.level":         0.98,
		"spot.R_ACK.q":             1e-3,
	}

	if err := config.LoadForTest(extra); err != nil {
		t.Errorf("Error while loading extra config")
	}

	LoadFromName("AVG_PKT_SIZE")
	LoadFromName("R_SYN")
	LoadFromName("R_ACK")
	LoadFromName("R_ICMP")
	LoadFromName("PERF")

	// logDataToFile = true
	// SetOutputDir("/home/asr/Documents/Work/go/src/netspot/analyzer/")
	time.Sleep(10 * time.Millisecond)
	// small
	// pcapFile1 : ~420min
	// logDataToFile = true
	// miner.SetDevice(pcapFile1)
	// period = 5 * time.Minute
	// huge
	// pcapFile2 : 900s
	miner.SetDevice(pcapFile2)
	SetPeriod(3 * time.Second)

	Start()
	time.Sleep(5 * time.Second)
	Stop()
	time.Sleep(2 * time.Second)
	exporter.Zero()
}

// func TestLivePcapMirai(t *testing.T) {
// 	title(t.Name())
// 	// SetLogging(1)
// 	Zero()

// 	extra := map[string]interface{}{
// 		"spot.PERF.depth":  100,
// 		"spot.PERF.n_init": 1000,
// 		"spot.PERF.level":  0.98,
// 		"spot.PERF.q":      1e-3,
// 		"spot.PERF.down":   false,
// 		"spot.PERF.up":     true,
// 	}
// 	if err := config.LoadForTest(extra); err != nil {
// 		t.Errorf("Error while loading extra config")
// 	}

// 	LoadFromName("PERF")

// 	// Mirai
// 	// pcapFile3 : 7137s
// 	miner.SetDevice(pcapFile3)
// 	period = 10 * time.Second
// 	SetPeriod(period)
// 	// logDataToFile = true
// 	// SetOutputDir("/tmp")

// 	// miner.SetTickPeriod(period)
// 	// miner.StartSniffing()
// 	// if !miner.IsSniffing() {
// 	// 	t.Error("Error: no sniffing")
// 	// }

// 	StartAndWait()
// 	// StartStats()
// 	// fmt.Println("START")
// 	// time.Sleep(10 * time.Second)
// 	// fmt.Println("STOP")
// 	// StopStats()
// }

func TestLive(t *testing.T) {
	title(t.Name())
	UnloadAll()
	config.Clean()
	config.LoadDefaults()
	config.LoadForTest(map[string]interface{}{
		"analyzer.stats": []string{"R_ARP", "R_ACK"},
	})

	if err := miner.InitConfig(); err != nil {
		t.Error(err)
	}
	if err := InitConfig(); err != nil {
		t.Error(err)
	}

	for _, f := range testFiles {
		if err := miner.SetDevice(f); err != nil {
			t.Error(err)
		}
		if err := StartAndWait(); err != nil {
			t.Error(err)
		}
	}

}
