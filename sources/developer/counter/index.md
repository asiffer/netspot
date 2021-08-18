---
title: Counters
weight: -12
summary: The counter is the more atomic object within netspot.
---

Counters are elements of the `miner`. In the Go project layout, they must be located in the `miner/counter` sub-package.

Every counter is related to a network layer (or directly to a packet). It means that it will access _only_ on its layer specific field (we do not recommend to parse further layers in counters logic).
The following layers are currently available:

- PKT (raw packet)
- ARP
- IPv4
- ICMPv4
- TCP
- UDP

A counter must implement 3 simple functions given by the interface below.

```go
type BaseCtrInterface interface {
	Name() string  // the name of the counter
	Value() uint64 // get back the value of the counter
	Reset()        // method to reset the counter
}
```

In addition, it must implement a `Process` function related to
the layers it depends on. For instance, an IPv4 counter must have
a method with the following signature

```go
Process(ip *layers.IPv4)
```

You can have a look to the counters already implemented within netspot. As a general example, a counter is a simple file with the
following content.

```go
// <layer>_<counter_name>.go

package counters

import (
	"github.com/google/gopacket/layers"
)

func init() {
	// The counter is registered at the very beginning
	Register(&COUNTER{Counter: 0})
}

// COUNTER is a custom Counter I made
type COUNTER struct {
	BaseCtr
	Counter uint64
}

// Name returns the name of the counter
func (c *COUNTER) Name() string {

}

// Value returns the current value of the counter
func (c *COUNTER) Value() uint64 {

}

// Reset resets the counter
func (c *COUNTER) Reset() {

}

// Process update the counter according to data it receives
func (c *COUNTER) Process(*layers.IPv4) {

}

```

Obviously, you can create more complex counters with additonal structures but you should keep in mind that its operation
should be rather atomic.

<!-- prettier-ignore -->
!!! Atomic
	We recall that several goroutines are likely to call the `Process` function. It means that the internal counter value access is under high concurrency. For this purpose, you may have a look to the [`atomic/sync`](https://golang.org/pkg/sync/atomic/) package. You can also have a look to [this example](https://gobyexample.com/atomic-counters).
