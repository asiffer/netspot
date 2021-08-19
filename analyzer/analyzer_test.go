// analyzer_test.go

package analyzer

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/asiffer/netspot/config"
	"github.com/asiffer/netspot/exporter"
	"github.com/asiffer/netspot/miner"

	"github.com/rs/zerolog"
)

var (
	HeaderWidth = 100
	HeaderSym   = "-"
	pcapFile1   string
	pcapFile2   string
	pcapFile3   string
)

var (
	testFiles []string
)

func TestMain(m *testing.M) {
	// Do stuff BEFORE the tests
	zerolog.SetGlobalLevel(zerolog.Disabled)
	testDir, err := findTestDir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Test dir:", testDir)

	infos, err := ioutil.ReadDir(testDir)
	if err != nil {
		fmt.Println(err)
	}
	testFiles = make([]string, len(infos))
	for i, f := range infos {
		testFiles[i] = path.Join(testDir, f.Name())
	}
	sort.Strings(testFiles)

	fmt.Println("Test files:", testFiles)

	pcapFile1 = testFiles[0] // empire.pcapng
	pcapFile2 = testFiles[4] // snort.pcap
	pcapFile3 = testFiles[2] // mirai.pcap

	// Netspot loads
	config.LoadDefaults()
	miner.InitLogger()
	if err := miner.InitConfig(); err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
	InitLogger()

	// run tests
	exitVal := m.Run()

	// Do stuff AFTER the tests!
	// nothing

	os.Exit(exitVal)
}

func dirExists(dir string) bool {
	_, err := os.Stat(dir)
	return !os.IsNotExist(err)
}

func findTestDir() (string, error) {
	wd, _ := os.Getwd()
	base := "/"
	for i := 0; i < 3; i++ {
		d := filepath.Join(wd, base, "test")
		if dirExists(d) {
			return d, nil
		}
		base += "../"
	}
	return "", fmt.Errorf("test/ directory not found")
}

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

	checkTitle("Checking stats description...")
	stats := GetAvailableStatsWithDesc()
	available := GetAvailableStats()
	for name, desc := range stats {
		if find(available, name) < 0 || len(desc) == 0 {
			testERROR()
			t.Fatalf("Stats '%s' has no description", name)
		}
	}
	testOK()

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
	checkTitle("Loading 3 stats...")
	LoadFromName("R_SYN")
	LoadFromName("R_ICMP")
	LoadFromName("R_DST_SRC")
	if len(GetLoadedStats()) != 3 {
		t.Errorf("Error while loading stats: %v", GetLoadedStats())
		testERROR()
	} else {
		testOK()
	}

	checkTitle("Unloading all the stats...")
	UnloadAll()
	if len(GetLoadedStats()) > 0 {
		t.Error("Error while removing all stats")
		testERROR()
	} else {
		testOK()
	}

	checkTitle("Checking loaded counters...")
	if len(miner.GetLoadedCounters()) > 0 {
		testERROR()
		t.Error("Error while removing underlying counters")
	} else {
		testOK()
	}
}

func TestUnloadSpecific(t *testing.T) {
	title("Testing specific unloading")

	checkTitle("Loading 2 stats...")
	LoadFromName("R_SYN")
	LoadFromName("R_ICMP")
	if len(GetLoadedStats()) != 2 {
		t.Errorf("Error while loading stats: %v", GetLoadedStats())
		testERROR()
	} else {
		testOK()
	}

	checkTitle("Unloading a single stat...")
	UnloadFromName("R_SYN")
	if find(GetLoadedStats(), "R_SYN") > 0 {
		testERROR()
		t.Error("Error while removing R_SYN")
	} else {
		testOK()
	}

	checkTitle("Checking loaded counters...")
	if find(miner.GetLoadedCounters(), "SYN") > 0 {
		testERROR()
		t.Error("Error while removing SYN counter")
	} else {
		testOK()
	}
}

func TestZero(t *testing.T) {
	title(t.Name())
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	defer zerolog.SetGlobalLevel(zerolog.Disabled)
	UnloadAll()
	LoadFromName("R_SYN")
	LoadFromName("AVG_PKT_SIZE")
	LoadFromName("R_ACK")
	LoadFromName("R_DST_SRC")
	LoadFromName("R_ICMP")

	if err := miner.SetDevice(testFiles[1]); err != nil {
		t.Fatal(err)
	}
	period = 5 * time.Minute

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

func TestLivePerfs(t *testing.T) {
	title(t.Name())
	// UnloadAll()
	Zero()
	// zerolog.SetGlobalLevel(zerolog.DebugLevel)
	miner.SetDevice(pcapFile1)
	SetPeriod(5 * time.Second)
	config.SetValue("exporter.console.data", true)
	exporter.InitConfig()
	// exporter.Load("console")
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

	// miner.SetDevice(pcapFile2)
	miner.SetDevice(testFiles[1])
	SetPeriod(1 * time.Second)

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
