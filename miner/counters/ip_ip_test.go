package counters

import (
	"testing"

	"github.com/google/gopacket/layers"
)

func TestIPCounter(t *testing.T) {
	title("Testing IP counter")
	ctr := &IP{Counter: 0}
	checkTitle("Check counter name...")
	if ctr.Name() != "IP" {
		testERROR()
		t.Errorf("Bad counter name (expected 'IP', got %s)", ctr.Name())
	}
	testOK()

	checkTitle("Check counter value...")
	if ctr.Value() != 0 {
		testERROR()
		t.Errorf("Bad counter value (expected 0, got %d)", ctr.Value())
	}
	testOK()

	checkTitle("Check layer processing...")
	ctr.Process(&layers.IPv4{Length: 17})
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
