package analyzer

import (
	"fmt"
	"netspot/config"
	"netspot/exporter"
	"netspot/miner"
	"testing"
	"time"
)

func TestLivePcapSmall(t *testing.T) {
	title(t.Name())

	config.Clean()
	config.LoadDefaults()
	exporter.Load("console")
	Zero()

	LoadFromName("R_SYN")
	LoadFromName("AVG_PKT_SIZE")
	LoadFromName("R_ACK")
	LoadFromName("R_DST_SRC")
	LoadFromName("R_ICMP")

	// small
	// pcapFile1 : 1282,198295430 seconds 15407 pkts
	miner.SetDevice(pcapFile1)
	SetPeriod(5 * time.Second)

	if err := StartAndWait(); err != nil {
		t.Errorf("Error while running: %v", err)
	}
}

func TestLivePcapHuge(t *testing.T) {
	title(t.Name())

	config.Clean()
	config.LoadDefaults()
	exporter.InitConfig()
	// exporter.Load("console")
	Zero()

	extra := map[string]interface{}{
		"spot.R_ACK.n_init": 800,
		"spot.R_ACK.level":  0.99,
		"spot.R_ACK.q":      1e-3,
	}
	config.LoadForTest(extra)

	LoadFromName("R_ACK")

	// bigger
	// pcapFile2 : 22,486727 seconds 142202 pkts
	miner.SetDevice(pcapFile2)
	SetPeriod(10 * time.Millisecond)

	if err := Start(); err != nil {
		t.Error(err)
	}
	time.Sleep(1000 * time.Millisecond)

	checkTitle("Getting stat status...")
	ss, err := StatStatus("R_ACK")
	if err != nil {
		t.Error(err.Error())
	}
	if ss.N > 0 {
		testOK()
	} else {
		testERROR()
	}

	// stop only if running
	if IsRunning() {
		if err := Stop(); err != nil {
			t.Error(err)
		}
	}
}

func TestLiveAll(t *testing.T) {
	title(t.Name())

	config.Clean()
	config.LoadDefaults()
	config.LoadForTest(map[string]interface{}{
		"analyzer.stats": []string{"R_ARP", "R_ACK", "R_IP"},
	})

	if err := miner.InitConfig(); err != nil {
		t.Error(err)
	}
	if err := exporter.InitConfig(); err != nil {
		t.Error(err)
	}
	if err := InitConfig(); err != nil {
		t.Error(err)
	}

	// pcap files      | duration           | nb packets
	// ----------------|--------------------|-------------
	// empire.pcapng   | 1282,198 seconds   | 15407
	// geeklounge.pcap | 25197,67 seconds   | 246137
	// mirai.pcap      | 7136,90 seconds    | 764137
	// openflow.pcapng | 378,895 seconds    | 1933
	// snort.pcap      | 22,48672 seconds   | 142202
	// toolsmith.pcap  | 10,868 seconds     | 392
	// wifi.pcap       | 88,089739 seconds  | 3957
	periods := []time.Duration{
		1 * time.Second,
		25 * time.Second,
		7 * time.Second,
		500 * time.Millisecond,
		20 * time.Millisecond,
		1 * time.Second,
		1 * time.Second,
	}
	for i, f := range testFiles {
		checkTitle(f)
		if err := miner.SetDevice(f); err != nil {
			t.Error(err)
		}
		SetPeriod(periods[i])
		if err := StartAndWait(); err != nil {
			t.Error(err)
			testERROR()
		} else {
			testOK()
		}
	}

}

func TestSnapshot(t *testing.T) {
	title(t.Name())

	config.Clean()
	config.LoadDefaults()
	config.LoadForTest(map[string]interface{}{
		"miner.device":    testFiles[2],
		"analyzer.period": "1s",
		"analyzer.stats":  []string{"R_ARP", "R_ACK", "R_IP"},
	})

	if err := miner.InitConfig(); err != nil {
		t.Error(err)
	}
	if err := exporter.InitConfig(); err != nil {
		t.Error(err)
	}
	if err := InitConfig(); err != nil {
		t.Error(err)
	}

	checkTitle("Starting the analyzer")
	if err := Start(); err != nil {
		testERROR()
		t.Fatal(err)
	}
	testOK()

	time.Sleep(50 * time.Millisecond)
	if values := StatValues(); values == nil {
		t.Log("StatValues return nil")
	} else {
		t.Logf("Values: %v", values)
		for k, v := range values {
			fmt.Println(k, v)
		}
	}

	Stop()

}
