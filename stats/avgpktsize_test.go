// avgpktsize_test.go

package stats

import (
	"math"
	"testing"
)

func TestAVGPKTSIZE(t *testing.T) {
	title("Testing AVG_PKT_SIZE")

	stat := &AvgPktSize{}
	checkTitle("Checking name...")
	if stat.Name() != "AVG_PKT_SIZE" {
		testERROR()
		t.Errorf("Expcted AVG_PKT_SIZE, got %s", stat.Name())
	} else {
		testOK()
	}

	checkTitle("Checking requirements...")
	if !isEqual(stat.Requirement(), []string{"IP_BYTES", "IP"}) {
		testERROR()
		t.Errorf("Expected [IP_BYTES, IP], got %s", stat.Requirement())
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
