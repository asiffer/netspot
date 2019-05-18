package counters

import (
	"testing"

	"github.com/google/gopacket/layers"
)

func TestUDPCounter(t *testing.T) {
	title("Testing UDP counter")
	ctr := &UDP{UDPCtr: NewUDPCtr(), Counter: 0}
	checkTitle("Check counter name...")
	if ctr.Name() != "UDP" {
		testERROR()
		t.Errorf("Bad counter name (expected 'UDP', got %s)", ctr.Name())
	}
	testOK()

	checkTitle("Check counter value...")
	if ctr.Value() != 0 {
		testERROR()
		t.Errorf("Bad counter value (expected 0, got %d)", ctr.Value())
	}
	testOK()

	checkTitle("Check layer processing...")
	ctr.Process(&layers.UDP{Length: 17})
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
