// analyezr_test.go

package analyzer

import (
	"fmt"
	"log"
	"netspot/miner"
	"strings"
	"testing"
	"time"
)

var (
	HeaderWidth = 100
	HeaderSym   = "-"
	pcapFile1   = "/home/asr/Documents/Work/Python/netspot/test/resources/4SICS-GeekLounge-151020.pcap"
	pcapFile2   = "/data/201111111400.dump"
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

func init() {
	DisableLogging()
}

func testOK() {
	fmt.Println("[\033[32mOK\033[0m]")
}

func testERROR() {
	fmt.Println("[\033[31mERROR\033[0m]")
}

func TestLoadStat(t *testing.T) {
	title("Testing stat loading")
	UnloadAll()

	checkTitle("Checking available stats...")
	if len(GetAvailableStats()) >= 5 {
		testOK()
	} else {
		testERROR()
		t.Errorf("Expected at least 5 stats, got %d", len(GetAvailableStats()))
	}

	checkTitle("Loading R_SYN...")
	id1, _ := LoadFromName("R_SYN")
	if id1 <= 0 {
		t.Error("Error while loading R_SYN")
		testERROR()
	} else {
		testOK()
	}

	checkTitle("Loading R_ACK...")
	id2, _ := LoadFromName("R_ACK")
	if id2 != (id1 + 1) {
		t.Error("Error while loading R_ACK")
		testERROR()
	} else {
		testOK()
	}

	checkTitle("Loading AVG_PKT_SIZE...")
	id3, _ := LoadFromName("AVG_PKT_SIZE")
	if id3 != (id2 + 1) {
		t.Error("Error while loading AVG_PKT_SIZE")
		testERROR()
	} else {
		testOK()
	}

	checkTitle("Reloading R_SYN...")
	id4, _ := LoadFromName("R_SYN")
	if id4 > 0 {
		t.Error("Error while re-loading R_SYN")
		testERROR()
	} else {
		testOK()
	}

	checkTitle("Checking number of loaded stats...")
	if len(GetLoadedStats()) != 3 {
		t.Errorf("Bad number of loaded counters (expected 3, got %d)", len(GetLoadedStats()))
		testERROR()
	} else {
		testOK()
	}
	UnloadAll()
}

func TestUnloadAll(t *testing.T) {
	title("Testing full unloading")
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
	title("Testing reset")

	UnloadAll()
	LoadFromName("R_SYN")
	LoadFromName("AVG_PKT_SIZE")
	LoadFromName("R_ACK")
	LoadFromName("R_DST_SRC")
	LoadFromName("R_ICMP")

	// small
	// pcapFile1 : ~420min
	miner.SetDevice(pcapFile1)
	period = 5 * time.Minute

	// huge
	// pcapFile2 : 900s
	// miner.SetDevice(pcapFile2)
	// period = 200 * time.Millisecond

	time.Sleep(1 * time.Second)
	miner.StartSniffing()
	StartStatsAndWait()

	fmt.Println("Reset")
	miner.Zero()
	Zero()
}
func TestLivePcap(t *testing.T) {
	title("Testing on PCAP")
	UnloadAll()
	LoadFromName("R_SYN")
	LoadFromName("AVG_PKT_SIZE")
	LoadFromName("R_ACK")
	LoadFromName("R_DST_SRC")
	LoadFromName("R_ICMP")

	// small
	// pcapFile1 : ~420min
	logDataToFile = true
	miner.SetDevice(pcapFile1)
	period = 5 * time.Minute

	// huge
	// pcapFile2 : 900s
	// miner.SetDevice(pcapFile2)
	// period = 200 * time.Millisecond

	miner.StartSniffing()
	if !miner.IsSniffing() {
		t.Error("Error: no sniffing")
	}

	start := time.Now()
	StartStatsAndWait()
	elapsed := time.Since(start)
	log.Printf("Timing: %f", elapsed.Seconds())
}
