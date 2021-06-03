package api

import (
	"encoding/json"
	"fmt"
)

// apiError wraps all the error that may occur when hitting API
type apiError struct {
	Msg string `json:"error" example:"Oh my god! Something wrong happened"`
}

func (e apiError) Error() string {
	return e.Msg
}

func (e apiError) JSON() []byte {
	bytes, err := json.Marshal(e)
	if err != nil {
		return []byte(fmt.Sprintf("An error occured while marshalling a first error: %v", err))
	}
	return bytes
}

// APIError inits a new error from a string message
func APIError(msg string) apiError {
	return apiError{Msg: msg}
}

// APIErrorf inits a new error with string formatting support
func APIErrorf(format string, a ...interface{}) apiError {
	return apiError{Msg: fmt.Sprintf(format, a...)}
}

// APIErrorFromMsg generates the error from an input error message
func APIErrorFromError(e error) apiError {
	return apiError{e.Error()}
}
