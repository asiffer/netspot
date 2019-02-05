package counters

import (
	"net"
	"testing"

	"github.com/google/gopacket/layers"
)

func TestNbUniqDstAddrCounter(t *testing.T) {
	title("Testing NB_UNIQ_DST_ADDR counter")
	ctr := &NbUniqDstAddr{IPCtr: NewIPCtr(), Addr: make(map[string]bool)}
	checkTitle("Check counter name...")
	if ctr.Name() != "NB_UNIQ_DST_ADDR" {
		testERROR()
		t.Errorf("Bad counter name (expected 'NB_UNIQ_DST_ADDR', got %s)", ctr.Name())
	}
	testOK()

	checkTitle("Check counter value...")
	if ctr.Value() != 0 {
		testERROR()
		t.Errorf("Bad counter value (expected 0, got %d)", ctr.Value())
	}
	testOK()

	checkTitle("Check layer processing...")
	ctr.Process(&layers.IPv4{DstIP: net.IP{1, 2, 3, 4}})
	ctr.Process(&layers.IPv4{DstIP: net.IP{5, 6, 7, 8}})
	ctr.Process(&layers.IPv4{DstIP: net.IP{1, 2, 3, 4}})
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
