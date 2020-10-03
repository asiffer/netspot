package miner

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var (
	headerWidth = 80
	testDir     string
	hugePcap    = "/data/pcap/202002071400.pcap"
)

func init() {
	setTestDir()
	hugePcap = filepath.Join(testDir, "snort.pcap")
}

func setTestDir() {
	wd, _ := os.Getwd()
	testDir = filepath.Join(wd, "../test")
	fmt.Println(testDir)
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

func genTraffic() {
	addr := "localhost:9000"
	_, err := net.Listen("tcp", addr)
	if err != nil {
		// handle error
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		// handle error
	}
	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)
		fmt.Fprintf(conn, "data")
	}
}

func TestGetAvailableDevices(t *testing.T) {
	title(t.Name())
	devs := GetAvailableDevices()
	if len(devs) < 2 {
		t.Errorf("Expecting more devices, found %d", len(devs))
	}
}

func TestGetAvailableCounters(t *testing.T) {
	title(t.Name())
	ctrs := GetAvailableCounters()
	if len(ctrs) < 17 {
		t.Errorf("Expecting more counters, found %d", len(ctrs))
	}
}

func TestLoading(t *testing.T) {
	title(t.Name())
	if err := Load("IP"); err != nil {
		t.Error(err)
	}
	if err := Load("ACK"); err != nil {
		t.Error(err)
	}
	if len(GetLoadedCounters()) != 2 {
		t.Errorf("Expecting %d counters, got %d",
			2, len(GetLoadedCounters()))
	}
	if err := Unload("IP"); err != nil {
		t.Error(err)
	}
	if len(GetLoadedCounters()) != 1 {
		t.Errorf("Expecting %d counters, got %d",
			1, len(GetLoadedCounters()))
	}

	if err := Load("SYN"); err != nil {
		t.Error(err)
	}
	if err := Load("ICMP"); err != nil {
		t.Error(err)
	}
	UnloadAll()
	if len(GetLoadedCounters()) != 0 {
		t.Errorf("Expecting %d counters, got %d",
			0, len(GetLoadedCounters()))
	}
}

func TestRun(t *testing.T) {
	title(t.Name())
	if err := SetDevice(hugePcap); err != nil {
		t.Error(err)
	}

	if err := Load("IP"); err != nil {
		t.Error(err)
	}

	_, err := Start()
	if err != nil {
		t.Error(err)
	}
	if !IsSniffing() {
		t.Errorf("Should sniff but it does not")
	}
	time.Sleep(1 * time.Second)

	GetSourceTime()
	// Maybe it has finished
	if !sniffing {
		return
	}

	if err := Stop(); err != nil {
		t.Error(err)
	}
}

func TestRunWithPeriod(t *testing.T) {
	title(t.Name())
	if err := SetDevice(hugePcap); err != nil {
		t.Error(err)
	}

	if err := Load("IP"); err != nil {
		t.Error(err)
	}

	data, err := StartAndYield(1 * time.Second)

	if err != nil {
		t.Error(err)
	}
	if !IsSniffing() {
		t.Errorf("Should sniff but it does not")
	}

	to := time.After(2 * time.Second)
	for {
		select {
		case <-data:
			// pass
		case <-to:
			if err := Stop(); err != nil {
				t.Error(err)
			}
			return
		default:
			// Maybe it has finished
			if !sniffing {
				return
			}
			// pass
		}
	}

}

func TestRunInterface(t *testing.T) {
	title(t.Name())
	if err := SetDevice("lo"); err != nil {
		t.Error(err)
	}

	if err := Load("IP"); err != nil {
		t.Error(err)
	}

	// generate traffic
	go genTraffic()

	// run
	data, err := StartAndYield(1 * time.Second)
	if err != nil {
		t.Error(err)
	}

	if !IsSniffing() {
		t.Errorf("Should sniff but it does not")
	}

	to := time.After(2 * time.Second)
	for {
		select {
		case <-data:
			// pass
		case <-to:
			if err := Stop(); err != nil {
				t.Error(err)
			}
			return
		default:
			// pass
		}
	}
	// time.Sleep(2 * time.Second)
	// <-data
	// fmt.Println("Stopping")
	// if err := Stop(); err != nil {
	// 	t.Error(err)
	// }

	// close(data)
}
