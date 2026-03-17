package collector

import (
	"fmt"
	"testing"
	"time"
)

func TestGoPacketCollector(t *testing.T) {
	source := "200704011400.dump"
	tick := 250 * time.Millisecond
	stop := make(chan bool)
	c := NewGoPacketCollector(source, tick)

	calls := 0
	prev := uint64(0)
	hook := func(data *Data) {
		calls += 1
		if prev > 0 {
			fmt.Println(time.Duration(data.TIME - prev))
		}
		prev = data.TIME
	}
	c.OnData(hook)
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	// blocking
	c.Start(stop)
	fmt.Println("Total calls:", calls)
}
