package counters

import (
	"log"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

func TestTimeCounter(t *testing.T) {
	title("Testing Time counter")
	ctr := &SOURCE_TIME{PktCtr: NewPktCtr(), Counter: 0}
	checkTitle("Check counter name...")
	if ctr.Name() != "TIME" {
		testERROR()
		t.Errorf("Bad counter name (expected 'Time', got %s)", ctr.Name())
	}
	testOK()

	checkTitle("Check counter value...")
	if ctr.Value() != 0 {
		testERROR()
		t.Errorf("Bad counter value (expected 0, got %d)", ctr.Value())
	}
	testOK()

	checkTitle("Check layer processing...")

	handle, err := pcap.OpenOffline("test/small.pcap")
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Loop through packets in file
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	pkt0, err := packetSource.NextPacket()

	ctr.Process(pkt0)
	if ctr.Value() != 1445353834524710000 {
		testERROR()
		t.Errorf("Bad counter value (expected 1445353834524710000, got %d)", ctr.Value())
	}
	// testOK()

	pkt1, err := packetSource.NextPacket()
	ctr.Process(pkt1)
	if ctr.Value() != 1445353834525737000 {
		testERROR()
		t.Errorf("Bad counter value (expected 1445353834525737000, got %d)", ctr.Value())
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
