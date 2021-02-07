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
	hugePcap    string
	smallPcap   string
	// hugePcap    = "/data/pcap/202002071400.pcap"
)

func init() {
	setTestDir()
	hugePcap = filepath.Join(testDir, "snort.pcap")
	smallPcap = filepath.Join(testDir, "empire.pcap")
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
	Zero()
	if err := SetDevice(hugePcap); err != nil {
		t.Error(err)
	}

	if err := Load("IP"); err != nil {
		t.Error(err)
	}

	_, err := Start(30 * time.Second)
	if err != nil {
		t.Error(err)
	}
	if !IsSniffing() {
		t.Errorf("Should sniff but it does not")
	}
	time.Sleep(1 * time.Second)

	GetSourceTime()
	// Maybe it has finished
	if !IsSniffing() {
		return
	}

	if err := Stop(); err != nil {
		t.Error(err)
	}
}

func TestRunWithPeriod(t *testing.T) {
	title(t.Name())
	Zero()
	if err := SetDevice(hugePcap); err != nil {
		t.Error(err)
	}

	if err := SetTimeout(0); err != nil {
		t.Error(err)
	}

	if err := Load("IP"); err != nil {
		t.Error(err)
	}

	data, err := Start(1 * time.Second)

	if err != nil {
		t.Error(err)
	}
	if !IsSniffing() {
		t.Errorf("Should sniff but it does not")
	}

	to := time.After(2 * time.Second)
	i := 0
	for {
		select {
		case <-data:
			// pass
			i++
			fmt.Println(i)
		case <-to:
			if err := Stop(); err != nil {
				t.Error(err)
			}
			return
		default:
			// Maybe it has finished
			if !IsSniffing() {
				return
			}
			// pass
		}
	}

}

func TestRunSmallPcap(t *testing.T) {
	title(t.Name())
	Zero()
	if err := SetDevice(smallPcap); err != nil {
		t.Error(err)
	}
	t.Log(GetSeriesName())

	if err := SetTimeout(0 * time.Second); err != nil {
		t.Error(err)
	}

	for _, c := range []string{"IP", "SYN", "ACK", "ICMP"} {
		if err := Load(c); err != nil {
			t.Error(err)
		}
	}

	_, err := Start(1500 * time.Second)
	if err != nil {
		t.Error(err)
	}
	if !IsSniffing() {
		t.Errorf("Should sniff but it does not")
	}

	// tro to start again
	if _, err := Start(0); err == nil {
		t.Error("An error must occur: cannot start again")
	}

	// try zero but it must fail
	if err := Zero(); err == nil {
		t.Error("An error must occur: cannot reset while sniffing")
	}

	to := time.After(2 * time.Second)
	for {
		select {
		// case <-data:
		// 	// pass
		case <-to:
			if err := Stop(); err != nil {
				t.Error(err)
			}
			return
		default:
			// Maybe it has finished
			if !IsSniffing() {
				// verify
				all := dispatcher.getAll()
				// truth values for empire.pcapng
				truth := map[string]uint64{
					"ACK":  14233,
					"ICMP": 0,
					"IP":   15377,
					"SYN":  671}
				for k, v := range all {
					if v != truth[k] {
						t.Errorf("Bad value of %s, expect %d, got %d",
							k, truth[k], v)
					}
				}
			}
			// pass
		}
	}
}

func TestRunInterface(t *testing.T) {
	title(t.Name())
	Zero()
	if err := SetDevice("lo"); err != nil {
		t.Error(err)
	}
	t.Log(GetSeriesName())

	if err := SetTimeout(0); err != nil {
		t.Error(err)
	}

	if err := Load("IP"); err != nil {
		t.Error(err)
	}

	// generate traffic
	go genTraffic()

	// run
	data, err := Start(1 * time.Second)
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
}
