// api.go

// Package api must be documented
package api

import (
	context "context"
	"fmt"
	"net"
	"netspot/config"
	"netspot/miner"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	grpc "google.golang.org/grpc"
)

var (
	apiLogger zerolog.Logger
	server    *grpc.Server
	network   string
	address   string
)

var (
	emptyResponse = &Empty{}
)

// init sets the default configuration
func init() {
	// for socket, see https://golang.org/pkg/net/#Listen
	// // default config
	// viper.SetDefault("api.network", "tcp")
	// viper.SetDefault("api.address", "localhost:11000")
}

// InitLogger initialize the sublogger for API
func InitLogger() {
	apiLogger = log.With().Str("module", "API").Logger()
}

// InitConfig prepare the API according to the config file
// In particular, it initializes the server but does not
// start to listen
func InitConfig() error {
	// parse endpoint
	var err error
	network, address, err = config.GetSocket("api.endpoint")
	if err != nil {
		return fmt.Errorf("Error while parsing endpoint: %v", err)
	}

	// init server
	server = grpc.NewServer()
	// register sub-API
	// RegisterManagerServer(server, &Manager{})
	// RegisterMinerServer(server, &Miner{})

	// logs
	apiLogger.Info().Msg("API package configured")
	return nil
}

// Manager API ============================================================== //
// ========================================================================== //
// ========================================================================== //

// Manager API
// type Manager struct{}

// // Start the analysis
// func (m *Manager) Start(context.Context, *Empty) (*Empty, error) {
// 	return emptyResponse, manager.Start()
// }

// // Stop the analysis
// func (m *Manager) Stop(context.Context, *Empty) (*Empty, error) {
// 	return emptyResponse, manager.Stop()
// }

// // Zero resets the internal state of netspot
// func (m *Manager) Zero(context.Context, *Empty) (*Empty, error) {
// 	return emptyResponse, manager.Zero()
// }

// Miner API ================================================================ //
// ========================================================================== //
// ========================================================================== //

// Miner API
type Miner struct{}

// GetAvailableDevices return a list of all the interfaces
// netspot can listen on
func (m *Miner) GetAvailableDevices(ctx context.Context, in *Empty) (*Devices, error) {
	available := miner.GetAvailableDevices()
	devices := Devices{Value: make([]*Device, len(available))}
	for i, d := range available {
		devices.Value[i] = &Device{Value: d}
	}
	return &devices, nil
}

// SetDevice defines the interface or capture file netspot will sniff
func (m *Miner) SetDevice(ctx context.Context, in *Device) (*Empty, error) {
	if miner.SetDevice(in.Value) != nil {
		return emptyResponse, fmt.Errorf("The device %s is neither a valid interface nor an existing network capture file", in.Value)
	}
	return emptyResponse, nil
}

// GetNbParsedPackets returns the current number of parsed packets
// if the counter PKTS is loaded
// func (m *Miner) GetNbParsedPackets(ctx context.Context, in *Empty) (*ParsedPackets, error) {
// 	pp, err := miner.GetNbParsedPackets()
// 	return &ParsedPackets{Value: pp}, err
// }

// GetStatus returns the current status of the miner
func (m *Miner) GetStatus(ctx context.Context, in *Empty) (*Status, error) {
	return &Status{Value: miner.RawStatus()}, nil
}

// Server =================================================================== //
// ========================================================================== //
// ========================================================================== //

// Serve set up the netspot server, the API is now
// available
func Serve() error {
	// open socket
	lis, err := net.Listen(network, address)
	if err != nil {
		return err
	}
	apiLogger.Info().Msgf("Start listening on %s://%s",
		network, address)
	return server.Serve(lis)
}

// Utils ==================================================================== //
// ========================================================================== //
// ========================================================================== //

func extractProtoAndAddr(endpoint string) (string, string, error) {
	s := strings.Split(endpoint, "://")
	if len(s) < 2 {
		return "", "", fmt.Errorf("Bad format, expect PROTO://ADDR, got %s", endpoint)
	}
	return s[0], s[1], nil
}
