// // pattern_test.go

// package counters

// import (
// 	"fmt"
// 	"net"
// 	"testing"
// 	"time"

// 	"github.com/google/gopacket/layers"
// )

// func TestSubnetContains(t *testing.T) {
// 	title("Testing subnet contains")
// 	subnet := &net.IPNet{
// 		IP:   net.IP{192, 168, 1, 15},
// 		Mask: net.IPMask{255, 255, 255, 255},
// 	}

// 	ip := &net.IPNet{
// 		IP:   net.IP{192, 168, 1, 100},
// 		Mask: net.IPMask{255, 255, 255, 255},
// 	}

// 	// fmt.Println(subnet.Equal(ip))

// 	msg := fmt.Sprintf("%s includes %s", subnet, ip)
// 	checkTitle("Checking if " + msg + " ...")
// 	if subnet.Contains(ip.IP) {
// 		testERROR()
// 		t.Errorf("First address must not contain second")
// 	} else {
// 		testOK()
// 	}

// 	subnet = &net.IPNet{
// 		IP:   net.IP{192, 168, 1, 15},
// 		Mask: net.IPMask{255, 255, 255, 0},
// 	}
// 	msg = fmt.Sprintf("%s includes %s", subnet, ip)
// 	checkTitle("Checking if " + msg + " ...")
// 	if !subnet.Contains(ip.IP) {
// 		testERROR()
// 		t.Errorf("First address must not contain second")
// 	} else {
// 		testOK()
// 	}

// 	subnet = &net.IPNet{
// 		IP:   net.IP{0, 0, 0, 0},
// 		Mask: net.IPMask{0, 0, 0, 0},
// 	}
// 	msg = fmt.Sprintf("%s includes %s", subnet, ip)
// 	checkTitle("Checking if " + msg + " ...")
// 	if !subnet.Contains(ip.IP) {
// 		testERROR()
// 		t.Errorf("First address must not contain second")
// 	} else {
// 		testOK()
// 	}

// }

// func TestPatternContains(t *testing.T) {
// 	title("Testing pattern contains")
// 	src := &net.IPNet{
// 		IP:   net.IP{192, 168, 1, 100},
// 		Mask: net.IPMask{255, 255, 255, 0},
// 	}

// 	pattern := &Pattern{
// 		Src:     src,
// 		Dst:     NullIPNet,
// 		SrcPort: 80,
// 		DstPort: NullPort,
// 	}

// 	other := &Pattern{
// 		Src: &net.IPNet{
// 			IP:   net.IP{192, 168, 1, 10},
// 			Mask: net.IPMask{255, 255, 255, 255},
// 		},
// 		Dst: &net.IPNet{
// 			IP:   net.IP{192, 168, 1, 15},
// 			Mask: net.IPMask{255, 255, 255, 255},
// 		},
// 		SrcPort: 80,
// 		DstPort: 55937,
// 	}

// 	msg := fmt.Sprintf("%s includes %s", pattern, other)
// 	checkTitle("Checking if " + msg + " ...")
// 	if !pattern.Contains(other) {
// 		testERROR()
// 		t.Errorf("First address must not contain second")
// 	} else {
// 		testOK()
// 	}

// 	other.SrcPort = 22

// 	msg = fmt.Sprintf("%s includes %s", pattern, other)
// 	checkTitle("Checking if " + msg + " ...")
// 	if pattern.Contains(other) {
// 		testERROR()
// 		t.Errorf("First address must not contain second")
// 	} else {
// 		testOK()
// 	}

// 	other.SrcPort = 80
// 	other.Src = &net.IPNet{
// 		IP:   net.IP{192, 168, 0, 10},
// 		Mask: net.IPMask{255, 255, 255, 255},
// 	}
// 	other.Dst = &net.IPNet{
// 		IP:   net.IP{192, 168, 1, 25},
// 		Mask: net.IPMask{255, 255, 255, 255},
// 	}

// 	msg = fmt.Sprintf("%s includes %s", pattern, other)
// 	checkTitle("Checking if " + msg + " ...")
// 	if pattern.Contains(other) {
// 		testERROR()
// 		t.Errorf("First address must not contain second")
// 	} else {
// 		testOK()
// 	}
// }

// func TestParsePattern(t *testing.T) {
// 	title("Parsing patterns")
// 	var s string

// 	s = "192.168.0.1/24: -> :22"
// 	checkTitle("Parsing '" + s + "' ...")
// 	pattern, err := ParsePattern(s)
// 	if err != nil {
// 		testERROR()
// 		t.Error(err)
// 	} else if !pattern.Src.IP.Equal(net.IP{192, 168, 0, 1}) {
// 		testERROR()
// 		t.Errorf("bad src ip (expected 192.168.0.1, got %s)", pattern.Src.IP)
// 	} else if !pattern.Dst.IP.Equal(net.IP{0, 0, 0, 0}) {
// 		testERROR()
// 		t.Errorf("bad dst ip (expected 0.0.0.0, got %s)", pattern.Dst.IP)
// 	} else if a, b := pattern.Src.Mask.Size(); a != 24 || b != 32 {
// 		testERROR()
// 		t.Errorf("bad src mask (expected 24, got %d)", a)
// 	} else if a, b := pattern.Dst.Mask.Size(); a != 0 || b != 32 {
// 		testERROR()
// 		t.Errorf("bad dst mask (expected 0, got %d)", a)
// 	} else if pattern.SrcPort != 0 {
// 		testERROR()
// 		t.Errorf("bad src port (expected 0, got %d)", pattern.SrcPort)
// 	} else if pattern.DstPort != 22 {
// 		testERROR()
// 		t.Errorf("bad dst port (expected 22, got %d)", pattern.DstPort)
// 	} else {
// 		testOK()
// 	}

// 	s = ":80 -> 10.10.10.10:22"
// 	checkTitle("Parsing '" + s + "' ...")
// 	pattern, err = ParsePattern(s)

// 	if err != nil {
// 		testERROR()
// 		t.Error(err)
// 	} else if !pattern.Src.IP.Equal(net.IP{0, 0, 0, 0}) {
// 		testERROR()
// 		t.Errorf("bad src ip (expected 0.0.0.0, got %s)", pattern.Src.IP)
// 	} else if !pattern.Dst.IP.Equal(net.IP{10, 10, 10, 10}) {
// 		testERROR()
// 		t.Errorf("bad dst ip (expected 10.10.10.10, got %s)", pattern.Dst.IP)
// 	} else if a, b := pattern.Src.Mask.Size(); a != 0 || b != 32 {
// 		testERROR()
// 		t.Errorf("bad src mask (expected 0, got %d)", a)
// 	} else if a, b := pattern.Dst.Mask.Size(); a != 32 || b != 32 {
// 		testERROR()
// 		t.Errorf("bad dst mask (expected 32, got %d)", a)
// 	} else if pattern.SrcPort != 80 {
// 		testERROR()
// 		t.Errorf("bad src port (expected 80, got %d)", pattern.SrcPort)
// 	} else if pattern.DstPort != 22 {
// 		testERROR()
// 		t.Errorf("bad dst port (expected 22, got %d)", pattern.DstPort)
// 	} else {
// 		testOK()
// 	}

// 	s = "192.168.0.1/24: -> 192.168.1.1:"
// 	checkTitle("Parsing '" + s + "' ...")
// 	pattern, err = ParsePattern(s)

// 	if err != nil {
// 		testERROR()
// 		t.Error(err)
// 	} else if !pattern.Src.IP.Equal(net.IP{192, 168, 0, 1}) {
// 		testERROR()
// 		t.Errorf("bad src ip (expected 192.168.0.1, got %s)", pattern.Src.IP)
// 	} else if !pattern.Dst.IP.Equal(net.IP{192, 168, 1, 1}) {
// 		testERROR()
// 		t.Errorf("bad dst ip (expected 192.168.1.1, got %s)", pattern.Dst.IP)
// 	} else if a, b := pattern.Src.Mask.Size(); a != 24 || b != 32 {
// 		testERROR()
// 		t.Errorf("bad src mask (expected 24, got %d)", a)
// 	} else if a, b := pattern.Dst.Mask.Size(); a != 32 || b != 32 {
// 		testERROR()
// 		t.Errorf("bad dst mask (expected 32, got %d)", a)
// 	} else if pattern.SrcPort != 0 {
// 		testERROR()
// 		t.Errorf("bad src port (expected 0, got %d)", pattern.SrcPort)
// 	} else if pattern.DstPort != 0 {
// 		testERROR()
// 		t.Errorf("bad dst port (expected 0, got %d)", pattern.DstPort)
// 	} else {
// 		testOK()
// 	}
// }

// func TestBPF(t *testing.T) {
// 	title("Checking BPF conversion")
// 	s := ":80 -> 10.10.10.10:22"
// 	bpf := "dst net 10.10.10.10/32 and src port 80 and dst port 22"
// 	checkTitle("Parsing '" + s + "' ...")
// 	pattern, _ := ParsePattern(s)

// 	if pattern.ToBPF() != bpf {
// 		t.Errorf("Expected %s, got %s", bpf, pattern.ToBPF())
// 		testERROR()
// 	} else {
// 		testOK()
// 	}
// }

// func TestPatternUpdate(t *testing.T) {
// 	title("Checking pattern update")
// 	pattern, err := ParsePattern("192.168.0.1:44000 -> 192.168.1.0/24:")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	layer := layers.IPv4{SrcIP: net.IP{192, 168, 0, 2},
// 		DstIP: net.IP{192, 168, 2, 1}}

// 	checkTitle("Updating from IP layer...")
// 	pattern.SetFromIPv4Layer(&layer)
// 	if !pattern.Src.IP.Equal(layer.SrcIP) {
// 		testERROR()
// 		t.Errorf("Expected %s, got %s", layer.SrcIP, pattern.Src.IP)
// 	} else if !pattern.Dst.IP.Equal(layer.DstIP) {
// 		testERROR()
// 		t.Errorf("Expected %s, got %s", layer.DstIP, pattern.Dst.IP)
// 	} else {
// 		testOK()
// 	}

// 	layer1 := layers.TCP{SrcPort: 22,
// 		DstPort: 80}
// 	checkTitle("Updating from TCP layer...")
// 	pattern.SetFromTCPLayer(&layer1)
// 	if pattern.SrcPort != int(layer1.SrcPort) {
// 		testERROR()
// 		t.Errorf("Expected %d, got %d", layer1.SrcPort, pattern.SrcPort)
// 	} else if pattern.DstPort != int(layer1.DstPort) {
// 		testERROR()
// 		t.Errorf("Expected %d, got %d", layer1.DstPort, pattern.DstPort)
// 	} else {
// 		testOK()
// 	}

// 	layer2 := layers.UDP{SrcPort: 50,
// 		DstPort: 8000}
// 	checkTitle("Updating from UDP layer...")
// 	pattern.SetFromUDPLayer(&layer2)
// 	if pattern.SrcPort != int(layer2.SrcPort) {
// 		testERROR()
// 		t.Errorf("Expected %d, got %d", layer2.SrcPort, pattern.SrcPort)
// 	} else if pattern.DstPort != int(layer2.DstPort) {
// 		testERROR()
// 		t.Errorf("Expected %d, got %d", layer2.DstPort, pattern.DstPort)
// 	} else {
// 		testOK()
// 	}
// }

// func TestProcessing(t *testing.T) {
// 	title("Checking packet processing")
// 	pattern, err := ParsePattern("192.168.0.1:44000 -> 192.168.1.0/24:")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	ctr := NewPatternCtr(pattern, "test")

// 	p0 := &Pattern{
// 		Src: &net.IPNet{
// 			IP:   net.IP{192, 168, 0, 1},
// 			Mask: net.IPMask{255, 255, 255, 255},
// 		},
// 		Dst: &net.IPNet{
// 			IP:   net.IP{192, 168, 1, 10},
// 			Mask: net.IPMask{255, 255, 255, 255},
// 		},
// 		SrcPort: 44000,
// 		DstPort: 22,
// 	}
// 	checkTitle("Processing right patterns...")
// 	ctr.Process(p0)
// 	ctr.Process(p0)
// 	if ctr.Value() != 2 {
// 		t.Errorf("Expected %d, got %d", 2, ctr.Value())
// 		testERROR()
// 	} else {
// 		testOK()
// 	}

// 	p1 := &Pattern{
// 		Src: &net.IPNet{
// 			IP:   net.IP{192, 168, 0, 1},
// 			Mask: net.IPMask{255, 255, 255, 255},
// 		},
// 		Dst: &net.IPNet{
// 			IP:   net.IP{192, 168, 5, 10},
// 			Mask: net.IPMask{255, 255, 255, 255},
// 		},
// 		SrcPort: 44000,
// 		DstPort: 22,
// 	}

// 	checkTitle("Processing bad pattern...")
// 	ctr.Process(p1)
// 	if ctr.Value() != 2 {
// 		t.Errorf("Expected %d, got %d", 2, ctr.Value())
// 		testERROR()
// 	} else {
// 		testOK()
// 	}

// 	go RunPatternCtr(ctr)

// 	ctr.LayPipe() <- p0
// 	checkTitle("Checking GET signal...")
// 	ctr.SigPipe() <- GET
// 	if <-ctr.ValPipe() != 3 {
// 		testERROR()
// 	} else {
// 		testOK()
// 	}

// 	checkTitle("Checking RESET signal...")
// 	ctr.SigPipe() <- RESET
// 	time.Sleep(50 * time.Millisecond)
// 	if ctr.Value() != 0 {
// 		testERROR()
// 	} else {
// 		testOK()
// 	}

// 	ctr.LayPipe() <- p0
// 	ctr.LayPipe() <- p0
// 	checkTitle("Checking FLUSH signal...")
// 	ctr.SigPipe() <- FLUSH
// 	<-ctr.ValPipe()
// 	time.Sleep(50 * time.Millisecond)
// 	if ctr.Value() != 0 {
// 		testERROR()
// 	} else {
// 		testOK()
// 	}

// 	ctr.LayPipe() <- p0
// 	checkTitle("Checking STOP signal...")
// 	ctr.SigPipe() <- STOP
// 	time.Sleep(50 * time.Millisecond)
// 	if ctr.IsRunning() {
// 		testERROR()
// 	} else {
// 		testOK()
// 	}

// }
