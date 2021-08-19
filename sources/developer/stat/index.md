---
title: Statistics
weight: -11
summary: Based on available counters, you can implement your own statistics
---

The statistics are the real values `netspot` monitors. In the same manner as the counters, you can implement your
own statistic according to your needs. In particular, they must be located
in the `netspot/stats` subpackage.

<!-- prettier-ignore -->
!!! note
    The analyzer is in charge of managing statistics. When you implement a 
    new statistic you don't have to take care of neither load/unload operations nor Spot algorithm configuration.

Statistics are built on top of the counters. They use some counters to
compute their own value. For example, the statistic `R_SYN` is computed
as the ratio `SYN`/`IP` where `SYN` counts the number of SYN packets and
`IP` counts the number of IP packets.

The general interface of a statistic (`StatInterface`) is quite rich but
only few functions must be implemented.

```go
// StatInterface gathers the common behavior of the statistics
type StatInterface interface {
	Name() string
	Configure() error
	Requirement() []string
	Compute(ctrvalues []uint64) float64
	Update(val float64) int
	UpProbability(quantile float64) float64
	DownProbability(quantile float64) float64
	GetThresholds() (float64, float64)
	Status() gospot.DSpotStatus
}
```

Actually, statistics must inherit from the the `BaseStat` object which
already implements some functions of the `StatInterface` (especially
all related to the Spot algorithm).

```go
// BaseStat is the basic structure which defines a statistic. It
// embeds a string (its unique name) and a DSpot instance which monitors
// itself.
type BaseStat struct {
	name  string
	dspot *gospot.DSpot // the spot instance
}
```

To define a new statistic you only need to inehrit from the previous
structure and code the following 3 functions:

```go
type StatInterface interface {
    // ...
    Name() string
    Requirement() []string
    Compute(ctrvalues []uint64) float64
    // ...
}
```

<!-- prettier-ignore -->
!!! warning
    You have to register your statistic by calling the 
    `Register(stat StatInterface)` in the `init()`
    function.

Finally, your new statistic should be defined in a single
file and may look like below.

```go
// my_stat.go

package stats

func init() {
    // The statistic is registered at the very beginning
    Register(&STAT{BaseStat{name: "..."}})
}

// STAT computes something to define
type STAT struct {
    BaseStat // compose with BaseStat
    // you can add some other attributes
}

// Name returns the unique name of the stat
func (stat *STAT) Name() string {
    return "..."
}

// Requirement returns the requested counters to compute the stat.
// The order is important for the Compute() function
func (stat *STAT) Requirement() []string {

}

// Compute implements the way to compute the stat from the counters.
// Counters are given in the same order as the Requirement() function
func (stat *STAT) Compute(ctrvalues []uint64) float64 {

}
```
