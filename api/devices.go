// devices.go

package api

import (
	"encoding/json"
	"net/http"
	"netspot/miner"
)

// DevicesHandler returns the list of available interfaces
func DevicesHandler(w http.ResponseWriter, r *http.Request) {
	// accept only GET
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only GET method is allowed"))
		return
	}
	// return the devices

	bytes, err := json.Marshal(miner.GetAvailableDevices())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}
