// stats.go

package api

import (
	"encoding/json"
	"net/http"

	"github.com/asiffer/netspot/analyzer"
)

// StatsHandler returns the list of the available stats along with
// their description
//
// @Summary List the available statistics
// @Description This returns the list of the statistics than can be loaded
// @Accept  json
// @Produce json
// @Header 200 {string} Content-Type "application/json"
// @Success 200 {object} map[string]string "Available statistics along with their description"
// @Failure 500 {object} apiError "error message"
// @Router /stats [get]
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
