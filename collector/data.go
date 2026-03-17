// Package collector provides structures and functions to collect and manage network data counters.
package collector

import "fmt"

// Key represents a counter identifier.
type Key uint32

// Enumeration of all possible counter keys.
const (
	PKT   Key = iota // Packet counter
	BYTES            // Byte counter
	IP               // IPv4 packet counter
	IP6              // IPv6 packet counter
	TCP              // TCP packet counter
	UDP              // UDP packet counter
	ICMP             // ICMP packet counter
	ICMP6            // ICMPv6 packet counter
	ARP              // ARP packet counter
	ACK              // ACK packet counter
	SYN              // SYN packet counter
	FIN              // FIN packet counter
	RST              // RST packet counter
	FRAG             // Fragmented IP packet counter
	TIME             // Not a real counter, used to loop over the counters (must be the last item)
)

// Keys returns a slice of all counter keys (excluding TIME).
func Keys() []Key {
	keys := make([]Key, TIME-PKT)
	for index, key := 0, PKT; key < TIME; index, key = index+1, key+1 {
		keys[index] = key
	}
	return keys
}

// String returns the string representation of the Key.
func (k Key) String() string {
	return [...]string{
		"PKT",
		"BYTES",
		"IP",
		"IP6",
		"TCP",
		"UDP",
		"ICMP",
		"ICMP6",
		"ARP",
		"ACK",
		"SYN",
		"FIN",
		"RST",
		"FRAG",
		"TIME",
	}[k]
}

// Data represents the main structure sent by the collectors, containing various network counters.
type Data struct {
	PKT           uint64        `json:"PKT"`           // Packet counter
	BYTES         uint64        `json:"BYTES"`         // Byte counter
	IP            uint64        `json:"IP"`            // IPv4 packet counter
	IP6           uint64        `json:"IP6"`           // IPv6 packet counter
	TCP           uint64        `json:"TCP"`           // TCP packet counter
	UDP           uint64        `json:"UDP"`           // UDP packet counter
	ICMP          uint64        `json:"ICMP"`          // ICMP packet counter
	ICMP6         uint64        `json:"ICMP6"`         // ICMPv6 packet counter
	ARP           uint64        `json:"ARP"`           // ARP packet counter
	ACK           uint64        `json:"ACK"`           // ACK packet counter
	SYN           uint64        `json:"SYN"`           // SYN packet counter
	FIN           uint64        `json:"FIN"`           // FIN packet counter
	RST           uint64        `json:"RST"`           // RST packet counter
	FRAG          uint64        `json:"FRAG"`          // Fragmented IP packet counter
	TIME          uint64        `json:"TIME"`          // Time counter
	TCP_DST_PORTS [65536]uint64 `json:"TCP_DST_PORTS"` // TCP destination ports counter
}

// addr returns a pointer to the structure attribute corresponding to the given counter key.
func (d *Data) Addr(key Key) *uint64 {
	switch key {
	case PKT:
		return &d.PKT
	case BYTES:
		return &d.BYTES
	case IP:
		return &d.IP
	case IP6:
		return &d.IP6
	case TCP:
		return &d.TCP
	case UDP:
		return &d.UDP
	case ICMP:
		return &d.ICMP
	case ICMP6:
		return &d.ICMP6
	case ARP:
		return &d.ARP
	case ACK:
		return &d.ACK
	case SYN:
		return &d.SYN
	case FIN:
		return &d.FIN
	case RST:
		return &d.RST
	case FRAG:
		return &d.FRAG
	case TIME:
		return &d.TIME
	}
	panic(fmt.Errorf("trying to access unavailable key: %d", key))
}
