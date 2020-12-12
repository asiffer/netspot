// run.go

package api

import (
	"fmt"
	"net/http"
	"netspot/analyzer"
)

// RunHandler manages start/stop actions
func RunHandler(w http.ResponseWriter, r *http.Request) {
	// accept only POST
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only POST method is allowed"))
		return
	}
	// parse form data
	if err := r.ParseForm(); err != nil {
		apiLogger.Error().Msg(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	// check the requested action
	switch action := r.PostFormValue("action"); action {
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
		msg := fmt.Sprintf("Operation %s is not supported", action)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
	}
}
