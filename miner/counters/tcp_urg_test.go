package counters

import (
	"testing"

	"github.com/google/gopacket/layers"
)

func TestURGCounter(t *testing.T) {
	title("Testing URG counter")
	ctr := &URG{counter: 0}
	checkTitle("Check counter name...")
	if ctr.Name() != "URG" {
		testERROR()
		t.Errorf("Bad counter name (expected 'URG', got %s)", ctr.Name())
	}
	testOK()

	checkTitle("Check counter value...")
	if ctr.Value() != 0 {
		testERROR()
		t.Errorf("Bad counter value (expected 0, got %d)", ctr.Value())
	}
	testOK()

	checkTitle("Check layer processing...")
	ctr.Process(&layers.TCP{URG: true})
	ctr.Process(&layers.TCP{URG: false})
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
