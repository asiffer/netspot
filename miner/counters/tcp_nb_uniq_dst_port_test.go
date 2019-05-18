package counters

import (
	"testing"

	"github.com/google/gopacket/layers"
)

func TestNbUniqDstPortCounter(t *testing.T) {
	title("Testing NB_UNIQ_DST_PORT counter")
	ctr := &NbUniqDstPort{TCPCtr: NewTCPCtr(), Port: make(map[uint16]bool)}
	checkTitle("Check counter name...")
	if ctr.Name() != "NB_UNIQ_DST_PORT" {
		testERROR()
		t.Errorf("Bad counter name (expected 'NB_UNIQ_DST_PORT', got %s)", ctr.Name())
	}
	testOK()

	checkTitle("Check counter value...")
	if ctr.Value() != 0 {
		testERROR()
		t.Errorf("Bad counter value (expected 0, got %d)", ctr.Value())
	}
	testOK()

	checkTitle("Check layer processing...")
	ctr.Process(&layers.TCP{DstPort: 22})
	ctr.Process(&layers.TCP{DstPort: 80})
	ctr.Process(&layers.TCP{DstPort: 22})
	if ctr.Value() != 2 {
		testERROR()
		t.Errorf("Bad counter value (expected 2, got %d)", ctr.Value())
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
