// run.go

package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/asiffer/netspot/analyzer"
)

// param name,param type,data type,is mandatory?,comment attribute(optional)

// RunHandler manages start/stop actions
//
// @Summary Manage the IDS status
// @Description Use this path to start/stop the IDS
// @Accept  json
// @Produce  json
// @Param action body string false "the action to perform" Enums("start", "stop")
// @Success 200 {string} string "Comment about the action performed"
// @Failure 400 {object} apiError "Error message"
// @Router /api/run [post]
func RunHandler(w http.ResponseWriter, r *http.Request) {
	// read content
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		apiLogger.Error().Msg(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(APIErrorFromError(err).JSON())
		return
	}
	data := make(map[string]string)
	if err := json.Unmarshal(raw, &data); err != nil {
		apiLogger.Error().Msg(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write(APIErrorFromError(err).JSON())
		return
	}

	action, ok := data["action"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(APIError("No 'action' key found").JSON())
		return
	}

	switch action {
	case "start":
		if err := analyzer.Start(); err != nil {
			apiLogger.Error().Msg(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write(APIErrorFromError(err).JSON())
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Analyzer has started"))
	case "stop":
		if err := analyzer.Stop(); err != nil {
			apiLogger.Error().Msg(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write(APIErrorFromError(err).JSON())
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Analyzer has stopped"))
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write(APIErrorf("action %s is not supported", action).JSON())
	}
}
