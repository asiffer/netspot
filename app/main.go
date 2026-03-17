package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asiffer/netspot/analyzer"
	"github.com/asiffer/netspot/collector"
	"github.com/asiffer/puzzle"
	"github.com/asiffer/puzzle/flagset"
)

var (
	collectorType string        = "gopacket"
	tick          time.Duration = 5 * time.Second
	source        string
	stats         []string = make([]string, 0)
	allStats      bool     = false
	output        string   = "netspot.jsonl"

	statsMap = map[string]analyzer.Stat{
		"RACK":   &analyzer.RACK{},
		"SYNFIN": &analyzer.SYNFIN{},
		"BPS":    &analyzer.BPS{},
		"PPS":    &analyzer.PPS{},
		"APS":    &analyzer.APS{},
		"DPE":    &analyzer.DPE{},
	}
)

var config = puzzle.NewConfig()

func init() {
	if err := errors.Join(
		puzzle.DefineVar(config, "collector", &collectorType, puzzle.WithDescription("Collector type (xdp or gopacket)")),
		puzzle.DefineVar(config, "source", &source, puzzle.WithDescription("Packet source (interface name or .pcap file)")),
		puzzle.DefineVar(config, "tick", &tick, puzzle.WithDescription("Stats computation period")),
		puzzle.DefineVar(config, "stats", &stats, puzzle.WithDescription("Stats to compute (RACK, SYNFIN, BPS, PPS, APS)")),
		puzzle.DefineVar(config, "all-stats", &allStats, puzzle.WithDescription("Compute all stats available")),
		puzzle.DefineVar(config, "output", &output, puzzle.WithShortFlagName("o"), puzzle.WithDescription("Output file for stats")),
	); err != nil {
		panic(err)
	}
}

func setupFromCLI() (collector.Collector, error) {
	fs, err := flagset.Build(config, "netspot", flag.ExitOnError)
	if err != nil {
		return nil, err
	}

	// parse cli args
	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	// populate stats manually (and validate)
	if allStats {
		// reset the array in case some stats were already added
		stats = make([]string, 0, len(statsMap))
		for stat := range statsMap {
			stats = append(stats, stat)
		}
	} else if len(stats) == 0 {
		return nil, fmt.Errorf("No stats specified, use --stats or --all-stats")
	} else {
		for _, s := range stats {
			if _, ok := statsMap[s]; !ok {
				return nil, fmt.Errorf("Unknown stat: %s", s)
			}
		}
	}

	// create the right collector
	switch collectorType {
	case "xdp":
		iface, err := net.InterfaceByName(source)
		if err != nil {
			return nil, err
		}
		return collector.NewXDPCollector(iface, tick), nil
	case "gopacket":
		if _, err := os.Stat(source); err != nil {
			return nil, err
		}
		return collector.NewGoPacketCollector(source, tick), nil
	default:
		return nil, fmt.Errorf("unknown collector type: %s", collectorType)
	}
}

func main() {
	co, err := setupFromCLI()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to setup")
	}

	logger.Info().
		Str("collector", collectorType).
		Str("source", co.Config().Source).
		Msg("Source defined")

	statsList := analyzer.NewMonitoredStatsList()
	for _, s := range stats {
		// already validated in setupFromCLI, no need to check existence
		stat, _ := statsMap[s]
		statsList.Add(analyzer.Monitor(stat))
	}
	logger.Info().Strs("stats", stats).Msg("Stats monitored")

	jsonl := JSONLHook{filename: output}
	if err := jsonl.Open(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to open JSONL output file")
	}
	defer func() {
		if err := jsonl.Close(); err != nil {
			logger.Error().Err(err).Msg("Failed to close JSONL output file")
		}
		logger.Info().Str("file", output).Msg("Data saved")
	}()

	co.
		OnLog(logger.Info().Msg).
		OnData(statsList.Hook).       // update the stats with new data from the collector
		OnData(jsonl.HandleCounters). // store also the counters
		OnError(func(err error) {
			logger.Error().Err(err).Msg("collector error")
		})

	statsList.
		OnData(jsonl.HandleRecord). // handle record + previously saved counters
		OnData(func(record *analyzer.Record) {
			info := logger.Info().Time("time", record.Time)
			for key, value := range record.Stats {
				info.Float64(key, value.Value)
			}
			info.Send()
		}).
		OnSpotError(func(s string, se *analyzer.SpotError) {
			logger.Error().
				Err(se.Err).
				Str("stat", s).
				Msg("Spot error")
		}).
		OnAlert(func(s string, v *analyzer.StatValue) {
			logger.Warn().
				Str("stat", s).
				Float64("value", v.Value).
				Float64("threshold", v.AnomalyThreshold).
				Float64("probability", v.Alert.Probability).
				Msg("Anomaly detected")
		})

	logger.Info().Msg("Loading collector")
	if err := co.Load(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to load collector")
	}

	stop := make(chan bool)
	go func() {
		sigChan := make(chan os.Signal, 1)
		// Notify channel on Interrupt (Ctrl+C) and SIGTERM (optional)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		// Block until a signal is received
		<-sigChan
		logger.Warn().Msg("\nReceived interrupt signal")
		stop <- true
	}()

	logger.Info().Msg("Starting")
	co.Start(stop) // blocking
	if err := co.Unload(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to unload collector")
	}
}
