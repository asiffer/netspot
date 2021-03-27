package api

import (
	"github.com/asiffer/netspot/config"
	"io/ioutil"
	"net/http"
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
		msg := "Config has been updated"
		apiLogger.Info().Msg(msg)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(msg))

	}

}
