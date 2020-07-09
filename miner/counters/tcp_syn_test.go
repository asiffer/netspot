package counters

import (
	"testing"

	"github.com/google/gopacket/layers"
)

func TestSYNCounter(t *testing.T) {
	title("Testing SYN counter")
	ctr := &SYN{counter: 0}
	checkTitle("Check counter name...")
	if ctr.Name() != "SYN" {
		testERROR()
		t.Errorf("Bad counter name (expected 'SYN', got %s)", ctr.Name())
	}
	testOK()

	checkTitle("Check counter value...")
	if ctr.Value() != 0 {
		testERROR()
		t.Errorf("Bad counter value (expected 0, got %d)", ctr.Value())
	}
	testOK()

	checkTitle("Check layer processing...")
	ctr.Process(&layers.TCP{SYN: true})
	ctr.Process(&layers.TCP{SYN: false})
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
