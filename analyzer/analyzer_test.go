// analyezr_test.go
package analyzer

import (
	"fmt"
	"netspot/miner"
	"strings"
	"testing"
	"time"
)

var (
	HEADER_WIDTH int    = 100
	HEADER_SYM   string = "*"
	PCAP_FILE_1  string = "/home/asr/Documents/Work/Python/netspot/test/resources/4SICS-GeekLounge-151020.pcap"
	PCAP_FILE_2  string = "/data/201111111400.dump"
)

func title(s string) {
	var l int = len(s)
	var border int
	var left string
	var right string
	remaining := HEADER_WIDTH - l - 2
	if remaining%2 == 0 {
		border = remaining / 2
		left = strings.Repeat(HEADER_SYM, border) + " "
		right = " " + strings.Repeat(HEADER_SYM, border)
	} else {
		border = (remaining - 1) / 2
		left = strings.Repeat(HEADER_SYM, border+1) + " "
		right = " " + strings.Repeat(HEADER_SYM, border)
	}

	fmt.Println(left + s + right)
}

func TestFirstConfig(t *testing.T) {
	title("Testing config loading")
	fmt.Println(GetLoadedStats())
	UnloadAll()
}

func TestAvailableStats(t *testing.T) {
	title("Testing stat availability")
	fmt.Println(GetAvailableStats())
}

func TestLoadStat(t *testing.T) {
	title("Testing stat loading")
	id1, _ := LoadFromName("R_SYN")
	if id1 <= 0 {
		t.Error("Error while loading R_SYN")
	}
	id2, _ := LoadFromName("R_ACK")
	if id2 != (id1 + 1) {
		t.Error("Error while loading R_ACK")
	}
	id3, _ := LoadFromName("AVG_PKT_SIZE")
	if id3 != (id2 + 1) {
		t.Error("Error while loading AVG_PKT_SIZE")
	}

	id4, _ := LoadFromName("R_SYN")
	if id4 > 0 {
		t.Error("Error while re-loading R_SYN")
	}
	fmt.Println(GetLoadedStats())
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
	// PCAP_FILE_1 : ~420min
	miner.SetDevice(PCAP_FILE_1)
	period = 5 * time.Minute

	// huge
	// PCAP_FILE_2 : 900s
	// miner.SetDevice(PCAP_FILE_2)
	// period = 200 * time.Millisecond

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
	// PCAP_FILE_1 : ~420min
	// miner.SetDevice(PCAP_FILE_1)
	// period = 5 * time.Minute

	// huge
	// PCAP_FILE_2 : 900s
	// miner.SetDevice(PCAP_FILE_2)
	// period = 200 * time.Millisecond

	// miner.StartSniffing()
	// if !miner.IsSniffing() {
	// 	t.Error("Error: no sniffing")
	// }

	// start := time.Now()
	// StartStatsAndWait()
	// elapsed := time.Since(start)
	// log.Printf("Timing: %f", elapsed.Seconds())
}
