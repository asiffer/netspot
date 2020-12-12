package miner

import (
	"encoding/hex"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func genTCPPacket() gopacket.Packet {
	// raw bytes of packet captured with
	// wireshark
	// TCP packet with SYN/ACK
	buffer, err := hex.DecodeString(
		"000000000000000000000000" +
			"08004500003c000040004006" +
			"3cba7f0000017f0000012710" +
			"8de208d40237eebf3bd1a012" +
			"ffcbfe3000000204ffd70402" +
			"080a7c9b61bc7c9b61bc0103" +
			"0307")
	if err != nil {
		panic(err)
	}
	return gopacket.NewPacket(
		buffer,
		layers.LayerTypeEthernet,
		gopacket.Default,
	)
}

func genICMPPacket() gopacket.Packet {
	// raw bytes of packet captured with
	// wireshark
	buffer, err := hex.DecodeString(
		"0000000000000000000000000" +
			"8004500005474d140004001c7" +
			"d57f0000017f0000010800a04" +
			"700010001bad23c5f00000000" +
			"97b10a0000000000101112131" +
			"415161718191a1b1c1d1e1f20" +
			"2122232425262728292a2b2c2" +
			"d2e2f3031323334353637")
	if err != nil {
		panic(err)
	}
	return gopacket.NewPacket(
		buffer,
		layers.LayerTypeEthernet,
		gopacket.Default,
	)
}

func genARPPacket() gopacket.Packet {
	// raw bytes of packet captured with
	// wireshark
	buffer, err := hex.DecodeString(
		"ffffffffffff309c23601fc808060001" +
			"080006040001309c23601fc8c0a8010e" +
			"000000000000c0a80001000000000000" +
			"00000000000000000000")
	if err != nil {
		panic(err)
	}

	return gopacket.NewPacket(
		buffer,
		layers.LayerTypeEthernet,
		gopacket.Default,
	)
}

func genUDPPacket() gopacket.Packet {
	// raw bytes of packet captured with
	// wireshark
	buffer, err := hex.DecodeString(
		"00000000000000000000000008004500" +
			"0020ab7e40004011914c7f0000017f00" +
			"0001cb882710000cfe1f5965730a")
	if err != nil {
		panic(err)
	}

	return gopacket.NewPacket(
		buffer,
		layers.LayerTypeEthernet,
		gopacket.Default,
	)
}

func TestLoadAndInit(t *testing.T) {
	d := NewDispatcher()
	// load
	ctrs := []string{"ARP", "IP", "ACK", "SYN", "PKTS", "ICMP", "UDP"}
	for _, c := range ctrs {
		if err := d.load(c); err != nil {
			t.Error(err)
		}
	}
	// init
	d.init()
	// pkt
	length := len(d.list.pkt)
	truth := 1
	if length != truth {
		t.Errorf("Bad size, expect %d got %d", truth, length)
	}
	// ip4
	length = len(d.list.ip4)
	truth = 1
	if length != truth {
		t.Errorf("Bad size, expect %d got %d", truth, length)
	}
	// tcp
	length = len(d.list.tcp)
	truth = 2
	if length != truth {
		t.Errorf("Bad size, expect %d got %d", truth, length)
	}
	// udp
	length = len(d.list.udp)
	truth = 1
	if length != truth {
		t.Errorf("Bad size, expect %d got %d", truth, length)
	}
	// icmp4
	length = len(d.list.icmp4)
	truth = 1
	if length != truth {
		t.Errorf("Bad size, expect %d got %d", truth, length)
	}
	// arp
	length = len(d.list.arp)
	truth = 1
	if length != truth {
		t.Errorf("Bad size, expect %d got %d", truth, length)
	}
}

func TestUnknownCounter(t *testing.T) {
	d := NewDispatcher()
	if err := d.load("WTF?"); err == nil {
		t.Errorf("Expect an error, got nil")
	}
}

func TestLoadUnload(t *testing.T) {
	d := NewDispatcher()
	ctrs := []string{"ARP", "IP", "ACK"}
	// load
	for _, c := range ctrs {
		if err := d.load(c); err != nil {
			t.Error(err)
		}
	}
	if len(d.counters) != len(ctrs) {
		t.Errorf("Wrong counters number, expect %d, got %d",
			len(ctrs), len(d.counters))
	}
	// loaded counters
	lc := d.loadedCounters()
	if len(lc) != len(ctrs) {
		t.Errorf("Wrong counters number, expect %d, got %d",
			len(ctrs), len(lc))
	}

	// unload
	for _, c := range ctrs {
		if err := d.unload(c); err != nil {
			t.Error(err)
		}
	}
	if len(d.counters) != 0 {
		t.Errorf("Wrong counters number, expect %d, got %d",
			0, len(d.counters))
	}
}

func TestDispatchARP(t *testing.T) {
	d := NewDispatcher()
	// load
	ctrs := []string{"ARP", "IP", "ACK", "SYN", "PKTS", "ICMP", "UDP"}
	for _, c := range ctrs {
		if err := d.load(c); err != nil {
			t.Error(err)
		}
	}
	// important!!
	d.init()
	// forge a packet
	a := genARPPacket()
	// dissect
	d.pool.Add(1)
	go d.dissect(a)
	d.terminate()
	// see result
	data := d.flushAll()
	for name, value := range data {
		if name == "ARP" || name == "PKTS" {
			if value != 1 {
				t.Errorf("[%s] Expecting %d, got %d", name, 1, value)
			}
		} else if value != 0 {
			t.Errorf("[%s] Expecting %d, got %d", name, 0, value)
		}
	}
}

func TestDispatchICMP(t *testing.T) {
	d := NewDispatcher()
	// load
	ctrs := []string{"ARP", "IP", "ACK", "SYN", "PKTS", "ICMP", "UDP"}
	for _, c := range ctrs {
		if err := d.load(c); err != nil {
			t.Error(err)
		}
	}
	// important!!
	d.init()
	// forge a packet
	a := genICMPPacket()
	// dissect
	d.pool.Add(1)
	go d.dissect(a)
	d.terminate()
	// see result
	data := d.flushAll()
	for name, value := range data {
		if name == "IP" || name == "ICMP" || name == "PKTS" {
			if value != 1 {
				t.Errorf("[%s] Expecting %d, got %d", name, 1, value)
			}
		} else if value != 0 {
			t.Errorf("[%s] Expecting %d, got %d", name, 0, value)
		}
	}

}

func TestDispatchTCP(t *testing.T) {
	d := NewDispatcher()
	// load
	ctrs := []string{"ARP", "IP", "ACK", "SYN", "PKTS", "ICMP", "UDP"}
	for _, c := range ctrs {
		if err := d.load(c); err != nil {
			t.Error(err)
		}
	}
	// important!!
	d.init()
	// forge a packet
	a := genTCPPacket()
	// dissect
	d.pool.Add(1)
	go d.dissect(a)
	d.terminate()
	// see result
	data := d.flushAll()
	for name, value := range data {
		if name == "IP" || name == "ACK" || name == "SYN" || name == "PKTS" {
			if value != 1 {
				t.Errorf("[%s] Expecting %d, got %d", name, 1, value)
			}
		} else if value != 0 {
			t.Errorf("[%s] Expecting %d, got %d", name, 0, value)
		}
	}
}

func TestDispatchUDP(t *testing.T) {
	d := NewDispatcher()
	// load
	ctrs := []string{"ARP", "IP", "ACK", "SYN", "PKTS", "ICMP", "UDP"}
	for _, c := range ctrs {
		if err := d.load(c); err != nil {
			t.Error(err)
		}
	}
	// important!!
	d.init()
	// forge a packet
	a := genUDPPacket()
	// dissect
	d.pool.Add(1)
	go d.dissect(a)
	d.terminate()
	// see result
	data := d.flushAll()
	for name, value := range data {
		if name == "IP" || name == "UDP" || name == "PKTS" {
			if value != 1 {
				t.Errorf("[%s] Expecting %d, got %d", name, 1, value)
			}
		} else if value != 0 {
			t.Errorf("[%s] Expecting %d, got %d", name, 0, value)
		}
	}

}
