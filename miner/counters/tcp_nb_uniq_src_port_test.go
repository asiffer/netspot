package counters

import (
	"testing"

	"github.com/google/gopacket/layers"
)

func TestNbUniqSrcPortCounter(t *testing.T) {
	title("Testing NB_UNIQ_SRC_PORT counter")
	ctr := &NbUniqSrcPort{TCPCtr: NewTCPCtr(), Port: make(map[uint16]bool)}
	checkTitle("Check counter name...")
	if ctr.Name() != "NB_UNIQ_SRC_PORT" {
		testERROR()
		t.Errorf("Bad counter name (expected 'NB_UNIQ_SRC_PORT', got %s)", ctr.Name())
	}
	testOK()

	checkTitle("Check counter value...")
	if ctr.Value() != 0 {
		testERROR()
		t.Errorf("Bad counter value (expected 0, got %d)", ctr.Value())
	}
	testOK()

	checkTitle("Check layer processing...")
	ctr.Process(&layers.TCP{SrcPort: 22, DstPort: 100})
	ctr.Process(&layers.TCP{SrcPort: 80, DstPort: 100})
	ctr.Process(&layers.TCP{SrcPort: 22, DstPort: 100})
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
