// run.go

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"netspot/analyzer"
)

// RunHandler manages start/stop actions
func RunHandler(w http.ResponseWriter, r *http.Request) {
	// read content
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		apiLogger.Error().Msg(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	data := make(map[string]string)
	if err := json.Unmarshal(raw, &data); err != nil {
		apiLogger.Error().Msg(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	action, ok := data["action"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No 'action' key found"))
		return
	}

	switch action {
	case "start":
		if err := analyzer.Start(); err != nil {
			apiLogger.Error().Msg(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Analyzer has started"))
	case "stop":
		if err := analyzer.Stop(); err != nil {
			apiLogger.Error().Msg(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Analyzer has stopped"))
	default:
		msg := fmt.Sprintf("Action %s is not supported", action)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
	}
}
