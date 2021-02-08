package counters

import (
	"log"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

func TestTimeCounter(t *testing.T) {
	title("Testing Time counter")
	ctr := &SOURCE_TIME{Counter: 0}
	checkTitle("Check counter name...")
	if ctr.Name() != "SOURCE_TIME" {
		testERROR()
		t.Errorf("Bad counter name (expected 'SOURCE_TIME', got %s)", ctr.Name())
	}
	testOK()

	checkTitle("Check counter value...")
	if ctr.Value() != 0 {
		testERROR()
		t.Errorf("Bad counter value (expected 0, got %d)", ctr.Value())
	}
	testOK()

	checkTitle("Check layer processing...")

	handle, err := pcap.OpenOffline(pcapTestFile)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Loop through packets in file
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	ts := []uint64{1558195813635381000, 1558195813732127000, 1558195813732181000}

	for i := 0; i < 3; i++ {
		pkt, err := packetSource.NextPacket()
		if err != nil {
			testERROR()
			t.Error(err)
		}
		ctr.Process(pkt)
		if ctr.Value() != ts[i] {
			testERROR()
			t.Errorf("Bad counter value (expected %d, got %d)", ts[i], ctr.Value())
		}
	}
	testOK()

	checkTitle("Check counter reset...")
	ctr.Reset()
	// the reset of this counter is special: it does nothing
	if ctr.Value() != ts[2] {
		testERROR()
		t.Errorf("Bad counter reset (expected 1445353834525737000, got %d)", ctr.Value())
	}
	testOK()
}
