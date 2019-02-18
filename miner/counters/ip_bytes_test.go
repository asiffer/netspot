package counters

import (
	"testing"

	"github.com/google/gopacket/layers"
)

func TestIPBYTESCounter(t *testing.T) {
	title("Testing IP_BYTES counter")
	ctr := &IPBytes{IPCtr: NewIPCtr(), Counter: 0}
	checkTitle("Check counter name...")
	if ctr.Name() != "IP_BYTES" {
		testERROR()
		t.Errorf("Bad counter name (expected 'IP_BYTES', got %s)", ctr.Name())
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
	if ctr.Value() != 17 {
		testERROR()
		t.Errorf("Bad counter value (expected 17, got %d)", ctr.Value())
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