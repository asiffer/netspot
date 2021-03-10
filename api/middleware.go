// logger.go
// Logging middleware

package api

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
)

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true

	return
}

// LoggingMiddleware logs the incoming HTTP request & its duration.
func LoggingMiddleware(h http.Handler) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {

		// turn the incoming response writer into
		// a pointer
		wrapped := wrapResponseWriter(w)

		// call the original http.Handler we're wrapping
		h.ServeHTTP(wrapped, r)

		// print logs
		var event *zerolog.Event
		switch code := wrapped.Status(); { // missing expression means "true"
		case code < 200:
			event = apiLogger.Warn()
			break
		case code < 300:
			event = apiLogger.Info()
			break
		case code < 400:
			event = apiLogger.Warn()
			break
		default:
			event = apiLogger.Error()
		}

		event.Msgf("%s %s [%d]", r.Method, r.RequestURI, wrapped.Status())
		apiLogger.Debug().Msgf("from:%s, referer:%s, user_agent:%s",
			requestSource(r),
			r.Header.Get("Referer"),
			r.Header.Get("User-Agent"))
	}

	return http.HandlerFunc(fn)

}

// BadMethodHandler catches requests with Bad methods
type BadMethodHandler struct{}

func (b BadMethodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("Method %s is not allowed", r.Method)
	apiLogger.Error().Msg(msg)
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte(msg))
}

// BadMethodHandler catch requests with Bad methods
// func BadMethodHandler(w http.ResponseWriter, r *http.Request) {
// 	msg := fmt.Sprintf("Method %s is not allowed", r.Method)
// 	apiLogger.Error().Msg(msg)
// 	w.WriteHeader(http.StatusMethodNotAllowed)
// 	w.Write([]byte(msg))
// }
