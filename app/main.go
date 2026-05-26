package app

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asiffer/netspot/analyzer"
	"github.com/asiffer/netspot/collector"
	"github.com/asiffer/puzzle/pflagset"
	"github.com/spf13/pflag"
)

func setupFromCLI() (collector.Collector, error) {
	fs, err := pflagset.Build(config, "netspot", pflag.ExitOnError)
	if err != nil {
		return nil, err
	}
	fs.SortFlags = false

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
		return nil, fmt.Errorf("No stats specified, use --stats=<stats> or --all-stats")
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

func Run() {
	co, err := setupFromCLI()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to setup")
	}

	logger.Info().
		Str("source", co.Config().Source).
		Msg("Source defined")

	statsList := analyzer.NewMonitoredStatsList()
	for _, s := range stats {
		// already validated in setupFromCLI, no need to check existence
		stat, _ := statsMap[s]
		options, err := getSpotOptions(s)
		if err != nil {
			logger.Fatal().Err(err).Str("stat", s).Msg("Failed to get SPOT options")
		}
		ms, err := analyzer.Monitor(stat, options...)
		if err != nil {
			logger.Fatal().Err(err).Str("stat", s).Msg("Failed to create monitored stat")
		}
		statsList.Add(ms)
	}
	logger.Info().Strs("stats", stats).Msg("Stats monitored")

	// ================================ HOOKS ================================
	co.
		OnLog(func(msg string) { // forward collector logs to app logger
			logger.Info().Msg(msg)
		}).
		OnError(func(err error) { // forward collector errors to app logger
			logger.Error().Err(err).Msg("collector error")
		}).
		OnData(statsList.Hook) // update the stats with new data from the collector

	// write jsonl logs
	if output != "" {
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
		co.OnData(jsonl.HandleCounters)      // store the counters
		statsList.OnData(jsonl.HandleRecord) // handle record + previously saved counters
	}

	statsList.
		// forward internal spot errors to app logger
		OnSpotError(func(s string, se *analyzer.SpotError) {
			logger.Error().
				Err(se.Err).
				Str("stat", s).
				Msg("Spot error")
		}).
		// forward stat alerts to app logger
		OnAlert(func(s string, t time.Time, v *analyzer.StatValue) {
			logger.Warn().
				Str("stat", s).
				Time("source_time_ns", t).
				TimeDiff("timesince_ns", t, co.FirstTimestamp()).
				Float64("value", v.Value).
				Float64("threshold", v.AnomalyThreshold).
				Float64("probability", v.Alert.Probability).
				Msg("Anomaly detected")
		}).
		// print on fit
		OnSpotFit(func(s string, t time.Time) {
			logger.Info().
				Str("stat", s).
				Time("source_time_ns", t).
				TimeDiff("timesince_ns", t, co.FirstTimestamp()).
				Msg("Stat fit")
		})

	if logRecords {
		// log records to stdout as well
		statsList.OnData(func(record *analyzer.Record) {
			info := logger.Info().Time("time", record.Time)
			for key, value := range record.Stats {
				info.Float64(key, value.Value)
			}
			info.Send()
		})
	}

	logger.Info().Str("collector", collectorType).Msg("Loading collector")
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
