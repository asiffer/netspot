package collector

import "time"

// CollectorConfig holds the configuration of a collector
// Source is either an interface name or a pcap file path
// Tick is the data collection period
type CollectorConfig struct {
	Source string        `json:"source"`
	Tick   time.Duration `json:"tick"`
}

// Collector is an interface that can be implemented for different
// source of data
type Collector interface {
	Hook
	Start(stop chan bool)
	Load() error
	Unload() error
	Config() CollectorConfig
}
