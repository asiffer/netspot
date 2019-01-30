// test_miner.go
package miner

import (
	"fmt"
	"netspot/miner/counters"
	"strings"
	"testing"
)

var HEADER_WIDTH int = 80

func title(s string) {
	var l int = len(s)
	var border int
	var left string
	var right string
	remaining := HEADER_WIDTH - l - 2
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

func TestAvailableCounter(t *testing.T) {
	title("Testing counter availability")
	fmt.Println(counters.AVAILABLE_COUNTERS)
}
func TestLogging(t *testing.T) {
	title("Testing logging")
	DisableLogging()
	SetLogging(1)
}

func TestGetAvailableDevices(t *testing.T) {
	title("Test Get Available Devices")
	devices := GetAvailableDevices()
	fmt.Println("Available devices: ", devices)
	if len(devices) < 2 {
		t.Error("Fail")
	}
}

func TestChangeDevice(t *testing.T) {
	title("Test Change Device")
	fmt.Printf("Current device: %s\n", GetDevice())
	SetDevice("lo")
	// fmt.Printf("Changing device to 'lo' ... (output: %d)\n", SetDevice("lo"))
	if GetDevice() != "lo" {
		t.Error("Fail")
	}
}
