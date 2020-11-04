// api.go

// Package api must be documented
package api

import (
	"fmt"
	"net"
	"net/http"
	"netspot/config"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	apiLogger zerolog.Logger
	router    *mux.Router
	network   string
	address   string
)

// init does nothin here
func init() {
	// for socket, see https://golang.org/pkg/net/#Listen
}

// removePort removes port from a source address if present
// "[::1]:58292" => "[::1]"
func removePort(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}

// requestSource returns ip address of the client making the request,
// taking into account http proxies
func requestSource(r *http.Request) string {
	hdrRealIP := r.Header.Get("X-Real-Ip")
	hdrForwardedFor := r.Header.Get("X-Forwarded-For")
	if hdrRealIP == "" && hdrForwardedFor == "" {
		return removePort(r.RemoteAddr)
	}
	if hdrForwardedFor != "" {
		// X-Forwarded-For is potentially a list of addresses separated with ","
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		if len(parts) == 1 {
			return parts[0]
		}
		return fmt.Sprintf("%s (%s)", parts[0], strings.Join(parts[1:], ", "))
	}
	return hdrRealIP
}

// this function wraps an handler to add basic logging
func logRequestHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// print logs
		apiLogger.Info().Msgf("%s %s", r.Method, r.RequestURI)
		apiLogger.Debug().Msgf("body: %s, from: %s, referer: %s, user_agent: %s",
			r.Body,
			requestSource(r),
			r.Header.Get("Referer"),
			r.Header.Get("User-Agent"))
		// call the original http.Handler we're wrapping
		h.ServeHTTP(w, r)
	}

	// http.HandlerFunc wraps a function so that it
	// implements http.Handler interface
	return http.HandlerFunc(fn)
}

// RunHandler returns the current config
func RunHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

// ConfigHandler returns the current config
func ConfigHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if data, err := config.JSON(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}

	default:
		msg := fmt.Sprintf("Method %s not supported", r.Method)
		apiLogger.Warn().Msg(msg)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(msg))

	}

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
	// server = grpc.NewServer()
	router := mux.NewRouter()
	router.HandleFunc("/api/run", RunHandler)
	router.HandleFunc("/api/config", ConfigHandler)
	http.Handle("/", logRequestHandler(router))
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
// type Miner struct{}

// GetAvailableDevices return a list of all the interfaces
// netspot can listen on
// func (m *Miner) GetAvailableDevices(ctx context.Context, in *Empty) (*Devices, error) {
// 	available := miner.GetAvailableDevices()
// 	devices := Devices{Value: make([]*Device, len(available))}
// 	for i, d := range available {
// 		devices.Value[i] = &Device{Value: d}
// 	}
// 	return &devices, nil
// }

// SetDevice defines the interface or capture file netspot will sniff
// func (m *Miner) SetDevice(ctx context.Context, in *Device) (*Empty, error) {
// 	if miner.SetDevice(in.Value) != nil {
// 		return emptyResponse, fmt.Errorf("The device %s is neither a valid interface nor an existing network capture file", in.Value)
// 	}
// 	return emptyResponse, nil
// }

// GetNbParsedPackets returns the current number of parsed packets
// if the counter PKTS is loaded
// func (m *Miner) GetNbParsedPackets(ctx context.Context, in *Empty) (*ParsedPackets, error) {
// 	pp, err := miner.GetNbParsedPackets()
// 	return &ParsedPackets{Value: pp}, err
// }

// GetStatus returns the current status of the miner
// func (m *Miner) GetStatus(ctx context.Context, in *Empty) (*Status, error) {
// 	return &Status{Value: miner.RawStatus()}, nil
// }

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

	return http.Serve(lis, nil)
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
