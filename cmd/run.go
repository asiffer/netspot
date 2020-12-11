package cmd

import (
	"fmt"
	"netspot/analyzer"
	"netspot/api"
	"netspot/config"
	"os"
	"sort"
	"strings"

	cli "github.com/urfave/cli/v2"
)

// RunServer is the entrypoint of the cli
func RunServer(c *cli.Context) error {
	if err := initConfig(c); err != nil {
		return err
	}
	// run server
	return api.Serve()
}

// RunCli starts directly the analyzer
func RunCli(c *cli.Context) error {
	if err := initConfig(c); err != nil {
		return err
	}
	// run cli
	// return analyzer.Run()
	return analyzer.StartAndWait()
}

// RunListStats prints the available statistics and return
func RunListStats(c *cli.Context) error {
	stats := analyzer.GetAvailableStats()
	// sort in-place
	sort.Strings(stats)
	fmt.Println(strings.Join(stats, "\n"))
	return nil
}

// RunPrintDefaults prints the default configuration
func RunPrintDefaults(c *cli.Context) error {
	return config.PrintTOML()
}

// Run starts netspot
func Run() error {
	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	return app.Run(os.Args)
}
