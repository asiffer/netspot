// ping.go

package api

import (
	"net/http"
)

// PingHandler responds to ping requests
func PingHandler(w http.ResponseWriter, r *http.Request) {
	// accept only GET
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only GET method is allowed"))
		return
	}
	// parse ping data
	w.WriteHeader(http.StatusOK)
	return
}
