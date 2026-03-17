package main

import (
	"encoding/json"
	"os"

	"github.com/asiffer/netspot/analyzer"
	"github.com/asiffer/netspot/collector"
)

type LoggerHook struct {
}

func (h *LoggerHook) HandleCounters(data *collector.Data) {
	info := logger.Info()
	for _, key := range collector.Keys() {
		value := data.Addr(key)
		info.Uint64(key.String(), *value)
	}
	info.Send()
}

func (h *LoggerHook) HandleRecord(record *analyzer.Record) {
	info := logger.Info()
	for key, value := range record.Stats {
		info.Float64(key, value.Value)
	}
	info.Send()
}

func (h *LoggerHook) HandleSpotError(name string, se *analyzer.SpotError) {
	logger.Error().
		Err(se.Err).
		Str("stat", name).
		Msg("Spot error")
}

// type JSONLRecord struct {
// 	analyzer.Record
// 	// Counters *collector.Data  `json:"counters,omitempty"`
// 	// Stats *analyzer.Record `json:"stats,omitempty"`
// }

type JSONLHook struct {
	filename string
	encoder  *json.Encoder
	file     *os.File
	data     *collector.Data
}

func (h *JSONLHook) Open() error {
	f, err := os.OpenFile(h.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	h.file = f
	h.encoder = json.NewEncoder(f)
	return nil
}

func (h *JSONLHook) Close() error {
	h.encoder = nil
	return h.file.Close()
}

func (h *JSONLHook) HandleCounters(data *collector.Data) {
	h.data = data
}

func (h *JSONLHook) HandleRecord(record *analyzer.Record) {
	// jsonRecord := JSONLRecord{Record: *record}
	// Counters: h.data,
	// 	Stats: record,
	// }
	if err := h.encoder.Encode(record); err != nil {
		logger.Error().Err(err).Msg("Failed to write JSONL record")
	}
	h.data = nil
}

// func StatAlertHook(alert analyzer.Alert) {
// 	logger.Warn().Str("stat", alert.Name).
// 		Float64("value", alert.Value).
// 		Float64("threshold", alert.Threshold).
// 		Float64("probability", alert.Probability).Send()
// }
