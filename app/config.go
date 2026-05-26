package app

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/asiffer/netspot/analyzer"
	"github.com/asiffer/puzzle"
)

var config = puzzle.NewConfig()

var (
	collectorType string        = "gopacket"
	tick          time.Duration = 5 * time.Second
	source        string
	stats         []string = make([]string, 0, len(statsMap))
	allStats      bool     = false
	output        string   = ""
	logRecords    bool     = false

	statsMap = map[string]analyzer.Stat{
		"RACK":   &analyzer.RACK{},
		"SYNFIN": &analyzer.SYNFIN{},
		"BPS":    &analyzer.BPS{},
		"PPS":    &analyzer.PPS{},
		"APS":    &analyzer.APS{},
		"DPE":    &analyzer.DPE{},
	}
)

func init() {
	if err := errors.Join(
		puzzle.DefineVar(config, "collector", &collectorType, puzzle.WithDescription("Collector type (xdp or gopacket)")),
		puzzle.DefineVar(config, "source", &source, puzzle.WithDescription("Packet source (interface name or .pcap file)"), puzzle.WithShortFlagName("s")),
		puzzle.DefineVar(config, "tick", &tick, puzzle.WithDescription("Stats computation period"), puzzle.WithShortFlagName("t")),
		puzzle.DefineVar(config, "stats", &stats, puzzle.WithDescription("Stats to compute")),
		puzzle.DefineVar(config, "all-stats", &allStats, puzzle.WithDescription("Compute all stats available"), puzzle.WithShortFlagName("a")),
		puzzle.DefineVar(config, "output", &output, puzzle.WithShortFlagName("o"), puzzle.WithDescription("Output file for all records")),
		puzzle.DefineVar(config, "log-records", &logRecords, puzzle.WithDescription("Enable logging of records"), puzzle.WithShortFlagName("v")),
	); err != nil {
		panic(err)
	}

	for stat := range statsMap {
		base := strings.ToLower(stat)
		if err := errors.Join(
			puzzle.Define(config, fmt.Sprintf("spot.%s.q", base), 5e-4, puzzle.WithDescription(fmt.Sprintf("%s detection probability", stat))),
			puzzle.Define(config, fmt.Sprintf("spot.%s.level", base), 0.98, puzzle.WithDescription(fmt.Sprintf("%s tail level", stat))),
			puzzle.Define(config, fmt.Sprintf("spot.%s.low", base), false, puzzle.WithDescription(fmt.Sprintf("%s lower tail monitoring", stat))),
			puzzle.Define(config, fmt.Sprintf("spot.%s.max-excess", base), uint64(1000), puzzle.WithDescription(fmt.Sprintf("%s max tail data", stat))),
		); err != nil {
			panic(err)
		}
	}

	config.SortFunc(func(keys []string) []string {
		nonSpotKeys := make([]string, 0)
		spotKeys := make([]string, 0)
		for _, key := range keys {
			if strings.HasPrefix(key, "spot.") {
				spotKeys = append(spotKeys, key)
			} else {
				nonSpotKeys = append(nonSpotKeys, key)
			}
		}
		sort.Strings(nonSpotKeys)
		sort.Strings(spotKeys)
		return append(nonSpotKeys, spotKeys...)
	})

}

// getSpotOptions returns the options for a given stat based on the config values
// (default values are provided otherwise)
func getSpotOptions(stat string) ([]analyzer.SpotOption, error) {
	base := strings.ToLower(stat)
	q, err := puzzle.Get[float64](config, fmt.Sprintf("spot.%s.q", base))
	if err != nil {
		return nil, err
	}
	level, err := puzzle.Get[float64](config, fmt.Sprintf("spot.%s.level", base))
	if err != nil {
		return nil, err
	}
	low, err := puzzle.Get[bool](config, fmt.Sprintf("spot.%s.low", base))
	if err != nil {
		return nil, err
	}
	maxExcess, err := puzzle.Get[uint64](config, fmt.Sprintf("spot.%s.max-excess", base))
	if err != nil {
		return nil, err
	}
	return []analyzer.SpotOption{
		analyzer.WithQ(q),
		analyzer.WithLevel(level),
		analyzer.WithLow(low),
		analyzer.WithMaxExcess(maxExcess),
	}, nil
}
