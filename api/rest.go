// rest.go

// Package api aims to provide interfaces to control the
// NetSpot server (HTTP and RPC)
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"netspot/analyzer"
	"netspot/miner"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// HTTP Error codes for client
const (
	// BADREQUEST is used when the server cannot or will not process the
	// request due to something that is perceived to be a client error (e.g.
	// malformed request syntax, invalid request message framing, or
	// deceptive request routing).
	BADREQUEST = 400
	// NOTFOUND is used when the origin server did not find a current
	// representation for the target resource or is not willing to disclose
	// that one exists.
	NOTFOUND = 404
	// NOTALLOWED is used when the method received in the request-line is
	// known by the origin server but not supported by the target resource.
	NOTALLOWED = 405
	// UNPROCESSABLE is used when the server understands the content type of
	// the request entity (hence a 415 Unsupported Media Type status code is
	// inappropriate), and the syntax of the request entity is correct (thus
	// a 400 Bad Request status code is inappropriate) but was unable to
	// process the contained instructions.
	UNPROCESSABLE = 422
	// OK is used when the request has succeeded.
	OK = 200
)

// HTTP Error codes for server
const (
	// INTERNAL is used when the server encountered an unexpected condition
	// that prevented it from fulfilling the request.
	INTERNAL = 500
)

//------------------------------------------------------------------------------
// Side functions
//------------------------------------------------------------------------------

// str2bool convert a string to a boolean. By default it returns false.
// It returns true for the following inputs: "1", "true", "True",
// "TRUE", "TrUe", "yes", "YES", "Yes" etc
func str2bool(s string) bool {
	low := strings.ToLower(s)
	return (low == "1") || (low == "true") || (low == "yes")
}

//------------------------------------------------------------------------------
// REST API
//------------------------------------------------------------------------------

// CONFIG ----------------------------------------------------------------------

func config(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodPost:
		return setConfig(w, r)
	case http.MethodGet:
		return getConfig(w, r)
	default:
		return NewHTTPError(nil, NOTALLOWED, "Method not allowed.")
	}
}

func getConfig(w http.ResponseWriter, r *http.Request) error {
	confAnalyzer := analyzer.GenericStatus()
	confMiner := miner.GenericStatus()

	// build the response
	response := make(map[string]interface{})
	for k, v := range confAnalyzer {
		response[k] = v
	}
	for k, v := range confMiner {
		response[k] = v
	}

	// marshall the Map
	js, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("Response marshalling error : %v", err)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(OK)

	_, err = w.Write(js)
	if err != nil {
		return fmt.Errorf("JSON writing error : %v", err)
	}
	return nil
}

func setConfig(w http.ResponseWriter, r *http.Request) error {
	// check method
	if r.Method != http.MethodPost {
		return NewHTTPError(nil, NOTALLOWED, "Method not allowed.")
	}

	// read request
	body, err := ioutil.ReadAll(r.Body) // Read request body.
	if err != nil {
		return fmt.Errorf("Request body read error : %v", err)
	}

	// unmarshal data
	data := make(map[string]string)
	if err = json.Unmarshal(body, &data); err != nil {
		return NewHTTPError(err, BADREQUEST, "Bad request: invalid JSON")
	}

	// loop over the key/value tuples
	apiLogger.Debug().Msgf("Received data: %v", data)
	for key, value := range data {
		// fmt.Println(key)
		switch key {
		case "device", "dev":
			if err := setDevice(value); err != nil {
				return err
			}
		case "promiscuous", "promisc":
			if err := setPromiscuous(value); err != nil {
				return err
			}
		case "period":
			if err := setPeriod(value); err != nil {
				return err
			}
		case "dir", "output":
			if err := setOutputDir(value); err != nil {
				return err
			}
		case "file":
			if err := setFileLogging(value); err != nil {
				return err
			}
		case "influx", "influxdb":
			if err := setInfluxDBLogging(value); err != nil {
				return err
			}
		default:
			// if a key is incorrect an error is returned
			err = fmt.Errorf("Invalid key: %s", key)
			return NewHTTPError(err, BADREQUEST, "Invalid JSON")
		}
	}
	// everything is ok
	w.WriteHeader(OK)
	return nil
}

func setDevice(dev string) error {
	// Perform the action
	if ret := miner.SetDevice(dev); ret != 0 {
		return NewHTTPError(nil, UNPROCESSABLE, "Device not available")
	}
	return nil
}

func setPromiscuous(p string) error {
	promisc := str2bool(p)
	// Perform the action
	if ret := miner.SetPromiscuous(promisc); ret != 0 {
		return NewHTTPError(nil, INTERNAL, "Unexpected error")
	}
	return nil
}

func setPeriod(p string) error {
	// parse
	d, err := time.ParseDuration(p)
	if err != nil {
		return NewHTTPError(err, BADREQUEST, "Period cannot be parsed")
	}
	analyzer.SetPeriod(d)
	return nil
}

func setOutputDir(p string) error {
	// parse
	err := analyzer.SetOutputDir(p)
	if err != nil {
		return NewHTTPError(err, BADREQUEST, "Directory error")
	}
	return nil
}

func setFileLogging(f string) error {
	err := analyzer.SetFileLogging(str2bool(f))
	if err != nil {
		return NewHTTPError(err, BADREQUEST, "File logging error")
	}
	return nil
}

func setInfluxDBLogging(i string) error {
	err := analyzer.SetInfluxDBLogging(str2bool(i))
	if err != nil {
		return NewHTTPError(err, BADREQUEST, "InfluxDB logging error")
	}
	return nil
}

// STATS -----------------------------------------------------------------------

func getLoadedStats(w http.ResponseWriter, r *http.Request) error {
	// check method
	if r.Method != http.MethodGet {
		return NewHTTPError(nil, NOTALLOWED, "Method not allowed.")
	}

	// build response
	response := map[string][]string{
		"loaded": analyzer.GetLoadedStats(),
	}
	js, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("Response marshalling error : %v", err)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(OK)

	_, err = w.Write(js)
	if err != nil {
		return fmt.Errorf("JSON writing error : %v", err)
	}
	return nil
}

func getAvailableStats(w http.ResponseWriter, r *http.Request) error {
	// check method
	if r.Method != http.MethodGet {
		return NewHTTPError(nil, NOTALLOWED, "Method not allowed.")
	}

	// build response
	response := map[string][]string{
		"available": analyzer.GetAvailableStats(),
	}
	js, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("Response marshalling error : %v", err)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(OK)

	_, err = w.Write(js)
	if err != nil {
		return fmt.Errorf("JSON writing error : %v", err)
	}
	return nil
}

func loadStat(w http.ResponseWriter, r *http.Request) error {
	// check method
	if r.Method != http.MethodPost {
		return NewHTTPError(nil, NOTALLOWED, "Method not allowed.")
	}

	// read request
	body, err := ioutil.ReadAll(r.Body) // Read request body.
	if err != nil {
		return fmt.Errorf("Request body read error : %v", err)
	}

	// unmarshal data
	data := make(map[string][]string)
	if err = json.Unmarshal(body, &data); err != nil {
		// error can occurs if an array is not passed but
		// only a single stat (string)
		datastr := make(map[string]string)
		if err = json.Unmarshal(body, &datastr); err != nil {
			return NewHTTPError(err, BADREQUEST, "Bad request: invalid JSON")
		}

		// try to access to the "stats" key
		stat, exists := datastr["stats"]
		if !exists {
			return NewHTTPError(nil, BADREQUEST, "Bad request: invalid JSON key")
		}
		data["stats"] = []string{stat}
	}

	// loop over the key/value tuples
	apiLogger.Debug().Msgf("Received data: %v", data)
	for _, stat := range data["stats"] {
		_, err := analyzer.LoadFromName(stat)
		if err != nil {
			return NewHTTPError(err, UNPROCESSABLE,
				"Cannot load the statistic (unknown or already loaded)")
		}
	}

	// everything is ok
	w.WriteHeader(OK)
	return nil
}

func unloadStat(w http.ResponseWriter, r *http.Request) error {
	// check method
	if r.Method != http.MethodPost {
		return NewHTTPError(nil, NOTALLOWED, "Method not allowed.")
	}

	// read request
	body, err := ioutil.ReadAll(r.Body) // Read request body.
	if err != nil {
		return fmt.Errorf("Request body read error : %v", err)
	}

	// unmarshal data
	data := make(map[string][]string)
	if err = json.Unmarshal(body, &data); err != nil {
		// error can occurs if an array is not passed but
		// only a single stat (string)
		datastr := make(map[string]string)
		if err = json.Unmarshal(body, &datastr); err != nil {
			return NewHTTPError(err, BADREQUEST, "Bad request: invalid JSON")
		}

		// try to access to the "stats" key
		stat, exists := datastr["stats"]
		if !exists {
			return NewHTTPError(nil, BADREQUEST, "Bad request: invalid JSON key")
		}

		// (specific to "unload") check if everything has to be unloaded
		if stat == "all" {
			analyzer.UnloadAll()
			// everything is ok
			w.WriteHeader(OK)
			return nil
		}
		data["stats"] = []string{stat}
	}

	statList, exists := data["stats"]
	if !exists {
		return NewHTTPError(nil, BADREQUEST, "Bad request: invalid JSON key")
	}
	// loop over the key/value tuples
	apiLogger.Debug().Msgf("Received data: %v", data)
	for _, stat := range statList {
		_, err := analyzer.UnloadFromName(stat)
		if err != nil {
			return NewHTTPError(err, UNPROCESSABLE,
				"Cannot unload the statistic (unknown or not already loaded)")
		}
	}

	// everything is ok
	w.WriteHeader(OK)
	return nil
}

func getStatStatus(w http.ResponseWriter, r *http.Request) error {
	// check method
	if r.Method != http.MethodGet {
		return NewHTTPError(nil, NOTALLOWED, "Method not allowed.")
	}

	// params := mux.Vars(r)
	// apiLogger.Debug().Msgf("GET params: %v", params)
	// stat, exists := params["stat"]
	// if !exists {
	// 	return NewHTTPError(nil, BADREQUEST, "Bad request: invalid GET parameter")
	// }
	stat := r.URL.Query().Get("stat")
	if len(stat) <= 0 {
		return NewHTTPError(nil, BADREQUEST, "Bad request: parameter is missing")
	}

	// retrieve values response
	response, err := analyzer.StatStatus(stat)
	if err != nil {
		return NewHTTPError(err, BADREQUEST, "Bad request: invalid stat")
	}
	// build response
	js, err := json.Marshal(response)
	if err != nil {
		// if an error occurs it is probably due to
		// NaN values which are not supported by JSON
		// Therefore, it must be done manually.
		// Normally, this problem is solved...
		return fmt.Errorf("Response marshalling error : %v", err)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(OK)

	_, err = w.Write(js) // check method
	if r.Method != http.MethodGet {
		return NewHTTPError(nil, NOTALLOWED, "Method not allowed.")
	}
	if err != nil {
		return fmt.Errorf("JSON writing error : %v", err)
	}
	return nil
}

func getStatValues(w http.ResponseWriter, r *http.Request) error {
	// check method
	if r.Method != http.MethodGet {
		return NewHTTPError(nil, NOTALLOWED, "Method not allowed.")
	}

	// build response
	response := analyzer.StatValues()
	js, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("Response marshalling error : %v", err)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(OK)

	_, err = w.Write(js)
	if err != nil {
		return fmt.Errorf("JSON writing error : %v", err)
	}
	return nil
}

// ACTIONS ---------------------------------------------------------------------

func start() error {
	if analyzer.IsRunning() {
		return errors.New("The statistics are currently computed")
	}

	if miner.IsSniffing() {
		return errors.New("The sniffer is already running")
	}

	// start the analyzer (it also starts the miner)
	analyzer.StartStats()
	return nil
}

func stop() error {
	if !analyzer.IsRunning() {
		return errors.New("The statistics are not currently monitored")
	}
	// stop the analyzer. It also stops the miner.
	analyzer.StopStats()
	return nil
}

func reload() error {
	miner.InitConfig()
	analyzer.InitConfig()
	return nil
}

func zero() error {
	if err := analyzer.Zero(); err != nil {
		return err
	}
	if err := miner.Zero(); err != nil {
		return err
	}
	return nil
}

func act(w http.ResponseWriter, r *http.Request) error {
	// check method
	if r.Method != http.MethodPost {
		return NewHTTPError(nil, NOTALLOWED, "Method not allowed.")
	}

	// read request
	body, err := ioutil.ReadAll(r.Body) // Read request body.
	if err != nil {
		return fmt.Errorf("Request body read error : %v", err)
	}

	// unmarshal data
	data := make(map[string]string)
	if err = json.Unmarshal(body, &data); err != nil {
		return NewHTTPError(err, BADREQUEST, "Bad request: invalid JSON")
	}

	// try to access to the "action" key
	command, exists := data["command"]
	if !exists {
		return NewHTTPError(nil, BADREQUEST, "Bad request: invalid JSON key")
	}

	switch strings.ToLower(command) {
	case "start":
		return start()
	case "stop":
		return stop()
	case "reload":
		return reload()
	case "zero":
		return zero()
	default:
		// if a key is incorrect an error is returned
		err = fmt.Errorf("Invalid command: %s", command)
		return NewHTTPError(err, BADREQUEST, "Invalid JSON")
	}
}

// OTHER -----------------------------------------------------------------------

func getAvailableInterfaces(w http.ResponseWriter, r *http.Request) error {
	// check method
	if r.Method != http.MethodGet {
		return NewHTTPError(nil, NOTALLOWED, "Method not allowed.")
	}

	// build response
	response := map[string][]string{
		"ifaces": miner.GetAvailableDevices(),
	}
	js, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("Response marshalling error : %v", err)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(OK)

	_, err = w.Write(js)
	if err != nil {
		return fmt.Errorf("JSON writing error : %v", err)
	}
	return nil
}

// ROUTES AND SERVER -----------------------------------------------------------

// setRoutes define the routes of the API. It associate a function (handler)
// to every route
func setRoutes(r *mux.Router) []*mux.Route {
	return []*mux.Route{
		r.Handle("/api/config", basicHandler(config)).Methods("GET", "POST"),
		r.Handle("/api/stats/loaded", basicHandler(getLoadedStats)).Methods("GET"),
		r.Handle("/api/stats/available", basicHandler(getAvailableStats)).Methods("GET"),
		r.Handle("/api/stats/load", basicHandler(loadStat)).Methods("POST"),
		r.Handle("/api/stats/unload", basicHandler(unloadStat)).Methods("POST"),
		r.Handle("/api/stats/values", basicHandler(getStatValues)).Methods("GET"),
		r.Handle("/api/stats/status", basicHandler(getStatStatus)).Methods("GET"),
		r.Handle("/api/ifaces/available", basicHandler(getAvailableInterfaces)).Methods("GET"),
		r.Handle("/api/run", basicHandler(act)).Methods("POST"),
	}
}

// RunHTTP starts the HTTP server
func RunHTTP(addr string, com chan error) {
	router := mux.NewRouter()
	setRoutes(router)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		// log.Fatal().Msgf("listen error: %v", err)
		com <- err
	}
	log.Info().Msgf("HTTP listening on %s", addr)

	err = http.Serve(listener, router)
	// err := http.ListenAndServe(addr, router)
	if err != nil {
		// log.Fatal().Msgf("server error: %v", err)
		com <- err
	}
}

// RunHTTPS starts the HTTPS server. Paths to a certificate and the private
// key are naturally needed.
func RunHTTPS(addr string, cert string, key string, com chan error) {
	router := mux.NewRouter()
	setRoutes(router)

	log.Info().Msgf("HTTPS server about to listen on %s", addr)
	err := http.ListenAndServeTLS(addr, cert, key, router)
	if err != nil {
		// log.Fatal().Msgf("server error: %v", err)
		com <- err
	}
}
