// devices.go

package api

import (
	"encoding/json"
	"net/http"

	"github.com/asiffer/netspot/miner"
)

// DevicesHandler returns the list of available interfaces
//
// @Summary List the available devices
// @Description This returns the list of the network interfaces that can be monitored
// @Accept  json
// @Produce json
// @Success 200 {array} string "list of the available devices"
// @Failure 500 {object} apiError "error message"
// @Router /api/devices [get]
func DevicesHandler(w http.ResponseWriter, r *http.Request) {
	// accept only GET
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(APIError("Only GET method is allowed").JSON())
		return
	}
	// return the devices

	bytes, err := json.Marshal(miner.GetAvailableDevices())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(APIErrorFromError(err).JSON())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}
