package counters

import (
	"testing"

	"github.com/google/gopacket/layers"
)

func TestRSTCounter(t *testing.T) {
	title("Testing RST counter")
	ctr := &RST{TCPCtr: NewTCPCtr(), Counter: 0}
	checkTitle("Check counter name...")
	if ctr.Name() != "RST" {
		testERROR()
		t.Errorf("Bad counter name (expected 'RST', got %s)", ctr.Name())
	}
	testOK()

	checkTitle("Check counter value...")
	if ctr.Value() != 0 {
		testERROR()
		t.Errorf("Bad counter value (expected 0, got %d)", ctr.Value())
	}
	testOK()

	checkTitle("Check layer processing...")
	ctr.Process(&layers.TCP{RST: true})
	ctr.Process(&layers.TCP{RST: false})
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
