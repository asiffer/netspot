// stats.go

package api

import (
	"encoding/json"
	"github.com/asiffer/netspot/analyzer"
	"net/http"
)

// StatsHandler returns the list of the available stats along with
// their description
func StatsHandler(w http.ResponseWriter, r *http.Request) {
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
