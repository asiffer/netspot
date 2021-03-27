// api.go

// Package api must be documented
package api

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/asiffer/netspot/analyzer"
	"github.com/asiffer/netspot/config"
	"github.com/asiffer/netspot/exporter"
	"github.com/asiffer/netspot/miner"

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
func init() {}

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

	// init router
	router := mux.NewRouter()
	router.Path("/").Methods("GET").HandlerFunc(DashboardHandler)
	router.Path("/api/run").Methods("POST").Headers("Content-Type", "application/json").HandlerFunc(RunHandler)
	router.Path("/api/config").Methods("GET", "POST").HandlerFunc(ConfigHandler)
	router.Path("/api/ping").Methods("GET").HandlerFunc(PingHandler)
	router.Path("/api/devices").Methods("GET").HandlerFunc(DevicesHandler)
	router.Path("/api/stats").Methods("GET").HandlerFunc(StatsHandler)

	// logging middleware
	router.Use(LoggingMiddleware)
	// reset the 404 handler
	router.MethodNotAllowedHandler = BadMethodHandler{}
	// r := router.NewRoute()
	// router.MethodNotAllowedHandler = r.HandlerFunc(BadMethodHandler).GetHandler()
	router.NotFoundHandler = router.NewRoute().HandlerFunc(http.NotFound).GetHandler()

	// static files
	http.Handle("/static/", http.FileServer(http.FS(root)))
	http.Handle("/", router)

	// logs
	apiLogger.Info().Msg("API package configured")
	return nil
}

// Utils ==================================================================== //
// ========================================================================== //
// ========================================================================== //

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

func initSubpackages() error {
	if err := miner.InitConfig(); err != nil {
		return fmt.Errorf("Error while initializing the Miner: %v", err)
	}
	if err := analyzer.InitConfig(); err != nil {
		return fmt.Errorf("Error while initializing the Analyzer: %v", err)
	}
	if err := exporter.InitConfig(); err != nil {
		return fmt.Errorf("Error while initializing the Exporter: %v", err)
	}
	return nil
}

func extractProtoAndAddr(endpoint string) (string, string, error) {
	s := strings.Split(endpoint, "://")
	if len(s) < 2 {
		return "", "", fmt.Errorf("Bad format, expect PROTO://ADDR, got %s", endpoint)
	}
	return s[0], s[1], nil
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

	return http.Serve(lis, nil)
}
