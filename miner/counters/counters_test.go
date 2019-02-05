// test_counters.go

package counters

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/gopacket/layers"
	"github.com/rs/zerolog"
)

var (
	headerWidth    = 80
	pcapTestFile   = "test/test.pcap"
	configTestFile = "test/miner.toml"
)

func init() {
	// disable logging
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

func checkTitle(s string) {
	format := "%-" + fmt.Sprint(headerWidth-7) + "s"
	fmt.Printf(format, s)
}

func testOK() {
	fmt.Println("[\033[32mOK\033[0m]")
}

func testERROR() {
	fmt.Println("[\033[31mERROR\033[0m]")
}

func title(s string) {
	var l = len(s)
	var border int
	var left string
	var right string
	remaining := headerWidth - l - 2
	if remaining%2 == 0 {
		border = remaining / 2
		left = strings.Repeat("-", border) + " "
		right = " " + strings.Repeat("-", border)
	} else {
		border = (remaining - 1) / 2
		left = strings.Repeat("-", border+1) + " "
		right = " " + strings.Repeat("-", border)
	}

	fmt.Println(left + s + right)

}

func TestGetAvailableCounters(t *testing.T) {
	title("Testing counter availability")
	ac := GetAvailableCounters()

	checkTitle("Checking constructors...")
	for _, ctr := range ac {
		constructor, exist := AvailableCounters[ctr]
		if !exist {
			testERROR()
			t.Errorf("The constructor of the counter '%s' does not exist", ctr)
		}
		constructor()
	}
	testOK()
}

func TestDoubleRegistering(t *testing.T) {
	title("Testing counter over-registering")
	err := Register("SYN", func() BaseCtrInterface {
		return &SYN{
			TCPCtr:  NewTCPCtr(),
			Counter: 0}
	})

	checkTitle("Checking register...")
	if err == nil {
		testERROR()
		t.Error("Double registering does not fail")
	}
	testOK()
}

func TestBaseCounter(t *testing.T) {
	title("Testing basic counter")
	base := NewBaseCtr()
	checkTitle("Checking value channel...")
	go func() { base.ValPipe() <- 5455 }()
	if v := <-base.ValPipe(); v != 5455 {
		testERROR()
		t.Error("Bad value")
	}
	testOK()

	checkTitle("Checking signal channel...")
	go func() { base.SigPipe() <- 3 }()
	if v := <-base.SigPipe(); v != 3 {
		testERROR()
		t.Error("Bad value")
	}
	testOK()

	checkTitle("Checking running state...")
	if base.IsRunning() {
		testERROR()
		t.Error("The counter must not be running")
	}
	testOK()

	checkTitle("Checking running state switching...")
	if base.SwitchRunningOn(); !base.IsRunning() {
		testERROR()
		t.Error("The counter must be running")
	}

	if base.SwitchRunningOff(); base.IsRunning() {
		testERROR()
		t.Error("The counter must not be running")
	}
	testOK()

}

// Run starts a counter, making it waiting for new incoming packets
func TestBaseRun(t *testing.T) {
	title("Testing base counter run")
	checkTitle("Running all kinds of counters")
	ip := &IPBytes{IPCtr: NewIPCtr(), Counter: 0}
	icmp := &ICMP{ICMPCtr: NewICMPCtr(), Counter: 0}
	tcp := &SYN{TCPCtr: NewTCPCtr(), Counter: 0}
	go Run(ip)
	go Run(tcp)
	go Run(icmp)
	go Run(nil)
	time.Sleep(1 * time.Second)
	if !ip.IsRunning() || !tcp.IsRunning() || !icmp.IsRunning() {
		testERROR()
	}
	testOK()
}

func TestIPRun(t *testing.T) {
	title("Testing IP counter run")
	ctr := &IP{IPCtr: NewIPCtr(), Counter: 0}
	go RunIPCtr(ctr)
	time.Sleep(5 * time.Millisecond)
	checkTitle("Check running status...")
	if !ctr.IsRunning() {
		testERROR()
		t.Error("The counter has not started")
	}
	testOK()

	checkTitle("Check value signal...")
	ctr.LayPipe() <- &layers.IPv4{Length: 10}
	ctr.SigPipe() <- 1
	time.Sleep(5 * time.Millisecond)
	if v := <-ctr.ValPipe(); v != 1 {
		testERROR()
		t.Errorf("Bad counter value (expected 1, got %d)", v)
	}
	testOK()

	checkTitle("Check reset signal...")
	ctr.LayPipe() <- &layers.IPv4{Length: 10}
	ctr.SigPipe() <- 2
	time.Sleep(5 * time.Millisecond)
	if ctr.Value() != 0 {
		testERROR()
		t.Errorf("Bad counter value (expected 0, got %d)", ctr.Value())
	}
	testOK()

	checkTitle("Check stop signal...")
	ctr.LayPipe() <- &layers.IPv4{Length: 10}
	ctr.SigPipe() <- 0
	time.Sleep(5 * time.Millisecond)
	if ctr.IsRunning() {
		testERROR()
		t.Error("Counter is still running")
	}
	testOK()
}

func TestICMPRun(t *testing.T) {
	title("Testing ICMP counter run")
	ctr := &ICMP{ICMPCtr: NewICMPCtr(), Counter: 0}
	go RunICMPCtr(ctr)
	time.Sleep(5 * time.Millisecond)
	checkTitle("Check running status...")
	if !ctr.IsRunning() {
		testERROR()
		t.Error("The counter has not started")
	}
	testOK()

	checkTitle("Check value signal...")
	ctr.LayPipe() <- &layers.ICMPv4{}
	ctr.SigPipe() <- 1
	time.Sleep(5 * time.Millisecond)
	if v := <-ctr.ValPipe(); v != 1 {
		testERROR()
		t.Errorf("Bad counter value (expected 1, got %d)", v)
	}
	testOK()

	checkTitle("Check reset signal...")
	ctr.LayPipe() <- &layers.ICMPv4{}
	ctr.SigPipe() <- 2
	time.Sleep(5 * time.Millisecond)
	if ctr.Value() != 0 {
		testERROR()
		t.Errorf("Bad counter value (expected 0, got %d)", ctr.Value())
	}
	testOK()

	checkTitle("Check stop signal...")
	ctr.LayPipe() <- &layers.ICMPv4{}
	ctr.SigPipe() <- 0
	time.Sleep(5 * time.Millisecond)
	if ctr.IsRunning() {
		testERROR()
		t.Error("Counter is still running")
	}
	testOK()
}

func TestTCPRun(t *testing.T) {
	title("Testing TCP counter run")
	ctr := &ACK{TCPCtr: NewTCPCtr(), Counter: 0}
	go RunTCPCtr(ctr)
	time.Sleep(5 * time.Millisecond)
	checkTitle("Check running status...")
	if !ctr.IsRunning() {
		testERROR()
		t.Error("The counter has not started")
	}
	testOK()

	checkTitle("Check value signal...")
	ctr.LayPipe() <- &layers.TCP{ACK: true}
	ctr.SigPipe() <- 1
	time.Sleep(5 * time.Millisecond)
	if v := <-ctr.ValPipe(); v != 1 {
		testERROR()
		t.Errorf("Bad counter value (expected 1, got %d)", v)
	}
	testOK()

	checkTitle("Check reset signal...")
	ctr.LayPipe() <- &layers.TCP{ACK: false}
	ctr.SigPipe() <- 2
	time.Sleep(5 * time.Millisecond)
	if ctr.Value() != 0 {
		testERROR()
		t.Errorf("Bad counter value (expected 0, got %d)", ctr.Value())
	}
	testOK()

	checkTitle("Check stop signal...")
	ctr.LayPipe() <- &layers.TCP{ACK: true}
	ctr.SigPipe() <- 0
	time.Sleep(5 * time.Millisecond)
	if ctr.IsRunning() {
		testERROR()
		t.Error("Counter is still running")
	}
	testOK()
}
