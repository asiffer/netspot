// api.go

// Package api must be documented
package api

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"netspot/analyzer"
	"netspot/config"
	"netspot/exporter"
	"netspot/miner"
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
	router.HandleFunc("/api/run", RunHandler).Methods("POST")
	router.HandleFunc("/api/config", ConfigHandler).Methods("GET", "POST")
	http.Handle("/", logRequestHandler(router))

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

func getBody(r *http.Request) string {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err.Error()
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	return string(bodyBytes)
}

// this function wraps an handler to add basic logging
func logRequestHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// print logs
		apiLogger.Info().Msgf("%s %s", r.Method, r.RequestURI)
		apiLogger.Debug().Msgf("body: %s, from: %s, referer: %s, user_agent: %s",
			getBody(r),
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

func jsonOnly(header http.Header) error {
	contentType := header.Values("Content-Type")
	if len(contentType) == 0 {
		return fmt.Errorf("Content-Type not given, accept 'application/json'")
	}
	if contentType[0] != "application/json" {
		return fmt.Errorf("Content-Type %s is not accepted, use 'application/json'",
			contentType[0])
	}
	return nil
}

func initSubpackages() error {
	if err := analyzer.InitConfig(); err != nil {
		return fmt.Errorf("Error while initializing the Analyzer: %v", err)
	}
	if err := miner.InitConfig(); err != nil {
		return fmt.Errorf("Error while initializing the Miner: %v", err)
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
