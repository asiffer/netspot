// rip_test.go

package stats

import (
	"math"
	"testing"
)

func TestRIP(t *testing.T) {
	title("Testing R_IP")

	stat := AvailableStats["R_IP"]
	checkTitle("Checking name...")
	if stat.Name() != "R_IP" {
		testERROR()
		t.Errorf("Expcted R_IP, got %s", stat.Name())
	} else {
		testOK()
	}

	checkTitle("Checking requirements...")
	if !isEqual(stat.Requirement(), []string{"IP", "PKTS"}) {
		testERROR()
		t.Errorf("Expected [IP, PKTS], got %s", stat.Requirement())
	} else {
		testOK()
	}

	checkTitle("Checking computation 1/3...")
	ctrvalues := []uint64{2, 5}
	if stat.Compute(ctrvalues) != 0.4 {
		testERROR()
		t.Errorf("Expected O.4, got %f", stat.Compute(ctrvalues))
	} else {
		testOK()
	}

	checkTitle("Checking computation 2/3...")
	ctrvalues = []uint64{0, 5}
	if stat.Compute(ctrvalues) != 0. {
		testERROR()
		t.Errorf("Expected O., got %f", stat.Compute(ctrvalues))
	} else {
		testOK()
	}

	checkTitle("Checking computation 3/3...")
	ctrvalues = []uint64{7, 0}
	if !math.IsNaN(stat.Compute(ctrvalues)) {
		testERROR()
		t.Errorf("Expected NaN, got %f", stat.Compute(ctrvalues))
	} else {
		testOK()
	}
}
