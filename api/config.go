package api

import (
	"io/ioutil"
	"net/http"

	"github.com/asiffer/netspot/config"
)

// ConfigPostHandler update the IDS config
//
// @Summary Update the config of the IDS
// @Description You can update the netspot config through this endpoint
// @Accept  json
// @Produce plain,json
// @Param config body map[string]string true "Input config"
// @Success 201 {string} string "Acknowledge message"
// @Failure 400 {object} apiError "Error message"
// @Failure 405 {object} apiError "Error message"
// @Failure 500 {object} apiError "Error message"
// @Router /config [post]
func ConfigPostHandler(w http.ResponseWriter, r *http.Request) {
	// accept only POST
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(APIError("Only POST method is allowed").JSON())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// read content
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		apiLogger.Error().Msg(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(APIErrorFromError(err).JSON())
		return
	}
	// make a backup
	config.Save()
	// load into config
	if err := config.LoadFromRawJSON(raw); err != nil {
		apiLogger.Error().Msg(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write(APIErrorFromError(err).JSON())
		config.Fallback()
		return
	}
	// reload packages
	if err := initSubpackages(); err != nil {
		apiLogger.Error().Msg(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write(APIErrorFromError(err).JSON())
		config.Fallback()
		return
	}
	msg := "Config has been updated"
	apiLogger.Info().Msg(msg)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(msg))

}

// ConfigGetHandler returns the current config
//
// @Summary Get the config of the IDS
// @Description You can fetch the netspot config through this endpoint
// @Accept  json
// @Produce json
// @Success 200 {object} string "Acknowledge message"
// @Failure 400 {object} apiError "Error message"
// @Failure 405 {object} apiError "Error message"
// @Failure 500 {object} apiError "Error message"
// @Router /config [get]
func ConfigGetHandler(w http.ResponseWriter, r *http.Request) {
	// accept only GET
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(APIError("Only GET method is allowed").JSON())
		return
	}

	if data, err := config.JSON(); err != nil {
		apiLogger.Error().Msg(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(APIErrorFromError(err).JSON())
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}

}
