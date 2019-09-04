// perf_test.go

package stats

import (
	"math"
	"testing"
)

func TestPERF(t *testing.T) {
	title("Testing PERF")

	stat := &Perf{}
	checkTitle("Checking name...")
	if stat.Name() != "PERF" {
		testERROR()
		t.Errorf("Expected PERF, got %s", stat.Name())
	} else {
		testOK()
	}

	checkTitle("Checking requirements...")
	if !isEqual(stat.Requirement(), []string{"PKTS", "REAL_TIME"}) {
		testERROR()
		t.Errorf("Expected [PKTS, REAL_TIME], got %s", stat.Requirement())
	} else {
		testOK()
	}

	checkTitle("Checking computation 1/3...")
	ctrvalues := []uint64{2, 0}
	statVal := stat.Compute(ctrvalues)
	if !math.IsNaN(statVal) {
		testERROR()
		t.Errorf("Expected NaN, got %f", statVal)
	} else {
		testOK()
	}

	checkTitle("Checking computation 2/3...")
	ctrvalues = []uint64{17, 100000000}
	statVal = stat.Compute(ctrvalues)
	if statVal != 170. {
		testERROR()
		t.Errorf("Expected 170., got %f", statVal)
	} else {
		testOK()
	}

	checkTitle("Checking computation 3/3...")
	ctrvalues = []uint64{25, 200000000}
	statVal = stat.Compute(ctrvalues)
	if statVal != 250. {
		testERROR()
		t.Errorf("Expected 250., got %f", statVal)
	} else {
		testOK()
	}

}
