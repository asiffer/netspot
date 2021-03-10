// ping.go

package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// PingHandler responds to ping requests
func PingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("PING HANDLER")
	// accept only GET
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only GET method is allowed"))
		return
	}

	// read content
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		apiLogger.Error().Msg(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
	return
}
