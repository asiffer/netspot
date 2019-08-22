// handler.go

package api

import "net/http"

// basicHandler is a basic object which will implement the
// Handle interface
type basicHandler func(http.ResponseWriter, *http.Request) error

// ServeHTTP is the generic function handling error
func (bh basicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	err := bh(w, r)
	if err == nil {
		return
	}

	// This is where our error handling logic starts.
	apiLogger.Error().Msgf("An error occured: %v", err) // Log the error.

	clientError, ok := err.(ClientError) // Check if it is a ClientError.
	if !ok {
		// If the error is not ClientError, assume that it is ServerError.
		w.WriteHeader(INTERNAL) // return 500 Internal Server Error.
		return
	}

	body, err := clientError.ResponseBody() // Try to get response body of ClientError.
	if err != nil {
		// problem when marshaling the error
		apiLogger.Error().Msgf("An error occured: %v", err)
		w.WriteHeader(INTERNAL)
		return
	}
	status, headers := clientError.ResponseHeaders() // Get http status code and headers.
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(status)
	w.Write(body)
}
