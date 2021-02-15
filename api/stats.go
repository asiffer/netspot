// stats.go

package api

import (
	"encoding/json"
	"net/http"
	"netspot/analyzer"
)

// StatsHandler returns the list of the available stats along with
// their description
func StatsHandler(w http.ResponseWriter, r *http.Request) {
	// accept only GET
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only GET method is allowed"))
		return
	}
	// return the stats

	bytes, err := json.Marshal(analyzer.GetAvailableStatsWithDesc())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}
