// bad_test.go
//
// It aims to deeply test the analyzer in weird cases

package analyzer

import (
	"netspot/miner"
	"testing"
	"time"
)

func prepareSniff() error {
	if err := Zero(); err != nil {
		return err
	}

	for _, stat := range []string{"R_SYN", "R_ACK", "R_IP"} {
		if err := LoadFromName(stat); err != nil {
			return err
		}
	}
	return miner.SetDevice(testFiles[1])
}

func TestBadStart(t *testing.T) {
	title(t.Name())
	if err := prepareSniff(); err != nil {
		t.Fatalf(err.Error())
	}

	// sending a bad ack message
	go func() { ackChannel <- STOPPED }()

	if err := Start(); err == nil {
		t.Errorf("An error was expected")
	}

	// receiveing the message of the loop
	<-ackChannel
	// sending stop
	if err := Stop(); err != nil {
		t.Errorf(err.Error())
	}
}

func TestBadStop(t *testing.T) {
	title(t.Name())
	if err := prepareSniff(); err != nil {
		t.Fatalf(err.Error())
	}

	if err := Start(); err != nil {
		t.Errorf("An error was expected")
	}

	// catch STOP message
	go func() { <-ackChannel }()
	time.Sleep(100 * time.Millisecond)
	// sending stop
	if err := Stop(); err == nil {
		t.Errorf("An error was expected")
	}

	// re-stopping
	if err := Stop(); err == nil {
		t.Errorf("An error was expected")
	}

}

func TestBadLoad(t *testing.T) {
	title(t.Name())

	Zero()
	if err := miner.Load("SYN"); err != nil {
		t.Errorf(err.Error())
	}

	LoadFromName("R_SYN")
}

func TestBadStartAndWait(t *testing.T) {
	title(t.Name())
	running.Begin()
	defer running.End()

	if err := StartAndWait(); err == nil {
		t.Errorf("An error was expected")
	}
}

func TestBadStartAndWait2(t *testing.T) {
	title(t.Name())
	Zero()

	if err := StartAndWait(); err == nil {
		t.Errorf("An error was expected")
	}
}
