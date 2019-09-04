// arp_arp_test.go

package counters

import (
	"testing"

	"github.com/google/gopacket/layers"
)

func TestARPCounter(t *testing.T) {
	title("Testing ARP counter")
	ctr := &ARP{ARPCtr: NewARPCtr(), Counter: 0}
	checkTitle("Check counter name...")
	if ctr.Name() != "ARP" {
		testERROR()
		t.Errorf("Bad counter name (expected 'ARP', got %s)", ctr.Name())
	}
	testOK()

	checkTitle("Check counter value...")
	if ctr.Value() != 0 {
		testERROR()
		t.Errorf("Bad counter value (expected 0, got %d)", ctr.Value())
	}
	testOK()

	checkTitle("Check layer processing...")
	ctr.Process(&layers.ARP{Protocol: 2048})
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
