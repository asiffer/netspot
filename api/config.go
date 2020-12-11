package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"netspot/config"
)

// ConfigHandler returns the current config
func ConfigHandler(w http.ResponseWriter, r *http.Request) {
	switch method := r.Method; method {
	case "GET":
		if data, err := config.JSON(); err != nil {
			apiLogger.Error().Msg(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}
	case "POST":
		// header
		if err := jsonOnly(r.Header); err != nil {
			apiLogger.Error().Msg(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
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
		// make a backup
		config.Save()
		// load into config
		if err := config.LoadFromRawJSON(raw); err != nil {
			apiLogger.Error().Msg(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			config.Fallback()
			return
		}
		// reload packages
		if err := initSubpackages(); err != nil {
			apiLogger.Error().Msg(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			config.Fallback()
			return
		}
		apiLogger.Info().Msg("Config has been updated")
		w.WriteHeader(http.StatusCreated)

	default:
		msg := fmt.Sprintf("Method %s not supported", method)
		apiLogger.Warn().Msg(msg)
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(msg))

	}

}
