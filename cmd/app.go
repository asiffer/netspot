// Package cmd manages the netspot CLI entrypoint
package cmd

import (
	"fmt"
	"strings"
	"time"

	cli "github.com/urfave/cli/v2"
)

// Version is the major netspot version
const Version = "2.1.2"

var (
	gitCommit string
)

var (
	desc = `
Netspot is a simple IDS powered by statistical learning. 
It actually monitors network statistics and detect abnormal events. 
Its core mainly relies on the SPOT algorithm (https://asiffer.github.io/libspot/) 
which flags extreme events on high throughput streaming data.
`
)

func getVersion() string {
	if len(gitCommit) == 0 {
		return Version
	}
	return fmt.Sprintf("%s %s", Version, gitCommit)
}

var (
	commonFlags = []cli.Flag{
		&cli.PathFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "Load configuration from `FILE`",
		},
		&cli.IntFlag{
			Name:    "log-level",
			Aliases: []string{"l"},
			Usage:   "Level of debug (0 is the most verbose)",
			Value:   1, // INFO
		},
	}

	minerFlags = []cli.Flag{
		&cli.StringFlag{
			Name:    "miner.device",
			Aliases: []string{"d"},
			Usage:   "Sniff `DEVICE` (pcap or interface)",
			Value:   "any",
		},
		&cli.BoolFlag{
			Name:  "miner.promiscuous",
			Value: false,
			Usage: "Activate promiscuous mode",
		},
		&cli.IntFlag{
			Name:  "miner.snapshot_len",
			Value: 65535,
			Usage: "Amount of captured bytes for each packet",
		},
		&cli.DurationFlag{
			Name:  "miner.timeout",
			Value: 30 * time.Second,
			Usage: "Time to wait before stopping if no packets is received",
		},
	}

	apiFLags = []cli.Flag{
		&cli.StringFlag{
			Name:    "api.endpoint",
			Aliases: []string{"e"},
			Usage:   "Listen to `ENDPOINT` (format: proto://addr)",
			Value:   "tcp://localhost:11000",
		},
	}

	analyzerFlags = []cli.Flag{
		&cli.DurationFlag{
			Name:    "analyzer.period",
			Aliases: []string{"p"},
			Value:   1 * time.Second,
			Usage:   "Time between two stats computations",
		},
		&cli.StringSliceFlag{
			Name:    "analyzer.stats",
			Aliases: []string{"s"},
			Usage:   "List of statistics to monitor",
		},
	}

	exporterFlags = []cli.Flag{
		&cli.BoolFlag{
			Name:    "exporter.console.data",
			Aliases: []string{"v"},
			Value:   true,
			Usage:   "Display statistics on the console",
		},
		&cli.StringFlag{
			Name:    "exporter.file.data",
			Aliases: []string{"f"},
			Value:   "netspot_%s_data.json",
			Usage:   "Log statistics to the given file",
		},
	}
)

var (
	app = &cli.App{
		Name:                 "netspot",
		Usage:                "A simple IDS with statistical learning",
		Authors:              []*cli.Author{{Name: "asr"}},
		Version:              getVersion(),
		Copyright:            "GPLv3",
		Description:          removeCharacters(desc, []string{"\n"}),
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:      "serve",
				Usage:     "Start netspot as a server",
				UsageText: "netspot serve [options]",
				Action:    RunServer,
				Flags:     concatFlags(commonFlags, minerFlags, analyzerFlags, apiFLags, exporterFlags),
			},
			{
				Name:   "run",
				Usage:  "Run netspot directly on the device",
				Action: RunCli,
				Flags:  concatFlags(commonFlags, minerFlags, analyzerFlags, exporterFlags),
			},
			{
				Name:    "list-stats",
				Usage:   "Print the available statistics",
				Action:  RunListStats,
				Aliases: []string{"ls"},
			},
			{
				Name:    "defaults",
				Usage:   "Print the default config",
				Action:  RunPrintDefaults,
				Aliases: []string{"def"},
			},
		},
	}
)

func concatFlags(flags ...[]cli.Flag) []cli.Flag {
	out := make([]cli.Flag, 0)
	for _, f := range flags {
		out = append(out, f...)
	}
	return out
}

func removeCharacters(s string, char []string) string {
	for _, c := range char {
		s = strings.Replace(s, c, "", -1)
	}
	return s
}

func Markdown() {
	fmt.Println("----------------------------------------")
	md, _ := app.ToMan()
	fmt.Println(md)
	fmt.Println("----------------------------------------")
}
