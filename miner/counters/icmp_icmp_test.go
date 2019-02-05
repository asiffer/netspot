package counters

import (
	"testing"

	"github.com/google/gopacket/layers"
)

func TestICMPCounter(t *testing.T) {
	title("Testing ICMP counter")
	ctr := &ICMP{ICMPCtr: NewICMPCtr(), Counter: 0}
	checkTitle("Check counter name...")
	if ctr.Name() != "ICMP" {
		testERROR()
		t.Error("Bad counter name")
	}
	testOK()

	checkTitle("Check counter value...")
	if ctr.Value() != 0 {
		testERROR()
		t.Errorf("Bad counter value (expected 0, got %d)", ctr.Value())
	}
	testOK()

	checkTitle("Check layer processing...")
	ctr.Process(&layers.ICMPv4{})
	if ctr.Value() != 1 {
		testERROR()
		t.Errorf("Bad counter value (expected 1, got %d)", ctr.Value())
	}
	testOK()

	checkTitle("Check counter reset...")
	ctr.Reset()
	if ctr.Value() != 0 {
		testERROR()
		t.Errorf("Bad counter reset (expected 0, got %d)", ctr.Value())
	}
	testOK()
}
