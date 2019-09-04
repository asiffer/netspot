// traffic_test.go

package stats

import (
	"fmt"
	"math"
	"testing"
)

func TestTRAFFIC(t *testing.T) {
	title("Testing TRAFFIC")

	stat := &Traffic{}
	checkTitle("Checking name...")
	if stat.Name() != "TRAFFIC" {
		testERROR()
		t.Errorf("Expected TRAFFIC, got %s", stat.Name())
	} else {
		testOK()
	}

	checkTitle("Checking requirements...")
	if !isEqual(stat.Requirement(), []string{"IP", "SOURCE_TIME"}) {
		testERROR()
		t.Errorf("Expected [IP, SOURCE_TIME], got %s", stat.Requirement())
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
	ctrvalues = []uint64{17, 100000}
	statVal = stat.Compute(ctrvalues)
	if statVal != 170. {
		testERROR()
		t.Errorf("Expected 170., got %f", statVal)
	} else {
		testOK()
	}

	checkTitle("Checking computation 3/3...")
	ctrvalues = []uint64{25, 300000}
	statVal = stat.Compute(ctrvalues)
	if statVal != 125.0 {
		testERROR()
		t.Errorf("Expected 125.0, got %f", statVal)
		fmt.Println(statVal)
	} else {
		testOK()
	}

}
