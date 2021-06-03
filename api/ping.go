// ping.go

package api

import (
	"io/ioutil"
	"net/http"
)

// PingHandler responds to ping requests
//
// @Summary Server healthcheck
// @Description This endpoints basically aims to check if the server is up
// @Accept  json
// @Produce  json
// @Success 200
// @Failure 405 {object} apiError "Error message"
// @Router /api/ping [get]
func PingHandler(w http.ResponseWriter, r *http.Request) {
	// accept only GET
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(APIError("Only GET method is allowed").JSON())
		return
	}

	// read content
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		apiLogger.Error().Msg(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(APIErrorFromError(err).JSON())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}
