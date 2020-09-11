package miner

import (
	"fmt"
	"testing"
	"time"
)

var hugePcap = "/data/pcap/202002071400.pcap"

func TestGetAvailableDevices(t *testing.T) {
	devs := GetAvailableDevices()
	if len(devs) < 2 {
		t.Errorf("Expecting more devices, found %d", len(devs))
	}
}

func TestGetAvailableCounters(t *testing.T) {
	ctrs := GetAvailableCounters()
	if len(ctrs) < 17 {
		t.Errorf("Expecting more counters, found %d", len(ctrs))
	}
}

func TestLoading(t *testing.T) {
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
	time.Sleep(2 * time.Second)

	GetSourceTime()

	if err := Stop(); err != nil {
		t.Error(err)
	}
}

func TestRunWithPeriod(t *testing.T) {
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
		}
	}

}

func TestRunInterface(t *testing.T) {
	if err := SetDevice("lo"); err != nil {
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

	fmt.Printf("Stopping")
	if err := Stop(); err != nil {
		t.Error(err)
	}

	close(data)
}
