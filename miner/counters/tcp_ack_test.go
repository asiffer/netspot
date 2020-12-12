package counters

import (
	"testing"

	"github.com/google/gopacket/layers"
)

func TestACKCounter(t *testing.T) {
	title("Testing ACK counter")
	ctr := &ACK{Counter: 0}
	checkTitle("Check counter name...")
	if ctr.Name() != "ACK" {
		testERROR()
		t.Errorf("Bad counter name (expected 'ACK', got %s)", ctr.Name())
	}
	testOK()

	checkTitle("Check counter value...")
	if ctr.Value() != 0 {
		testERROR()
		t.Errorf("Bad counter value (expected 0, got %d)", ctr.Value())
	}
	testOK()

	checkTitle("Check layer processing...")
	ctr.Process(&layers.TCP{ACK: true})
	ctr.Process(&layers.TCP{ACK: false})
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
