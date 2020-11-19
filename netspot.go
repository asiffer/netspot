// netspot.go

//
// Netspot is a simple IDS with statistical learning.
// It works either directly on a device
// (network interface or .pcap file)
// or through a server exposing a basic REST API.
//
// Basically, netspot monitors network statistics and detect abnormal events.
// Its core mainly relies on the SPOT algorithm
// (https://asiffer.github.io/libspot/)
// which flags extreme events on high throughput streaming data.
//
// Get Started
//
// The following example runs netspot on the localhost interface.
//
//  netspot run -d lo -s R_SYN -p 2s -v
//
// In particular, it monitors the ratio of SYN packets every 2 seconds
// Raw data are printed to the console.
package main

import (
	"netspot/cmd"

	"github.com/rs/zerolog/log"
)

func main() {
	// run netspot
	if err := cmd.Run(); err != nil {
		log.Fatal().Msgf("%v", err)
	}
}
