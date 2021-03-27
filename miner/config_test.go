package miner

import (
	"testing"
	"time"

	"github.com/asiffer/netspot/config"
)

func TestInitConfig(t *testing.T) {
	title(t.Name())
	InitLogger()

	conf := map[string]interface{}{
		"miner.device":       hugePcap,
		"miner.snapshot_len": int32(65535),
		"miner.promiscuous":  false,
		"miner.timeout":      0 * time.Second,
	}
	if err := config.LoadForTest(conf); err != nil {
		t.Error(err)
	}
	if err := InitConfig(); err != nil {
		t.Errorf(err.Error())
	}

	if GetDevice() != hugePcap {
		t.Errorf("Bad device, expect %s, got %s", hugePcap, GetDevice())
	}

}

func TestInitBadDevice(t *testing.T) {
	title(t.Name())
	conf := map[string]interface{}{
		"miner.device":       "unknown",
		"miner.snapshot_len": int32(65535),
		"miner.promiscuous":  false,
		"miner.timeout":      10 * time.Second,
	}
	if err := config.LoadForTest(conf); err != nil {
		t.Error(err)
	}
	if err := InitConfig(); err == nil {
		t.Errorf(err.Error())
	}

}

func TestInitBadSnapshotLength(t *testing.T) {
	title(t.Name())
	conf := map[string]interface{}{
		"miner.device":       "any",
		"miner.snapshot_len": -1000,
	}
	if err := config.LoadForTest(conf); err != nil {
		t.Errorf(err.Error())
	}
	if err := InitConfig(); err == nil {
		t.Errorf("An error should occur")
	}

}

func TestInitBadTimeout(t *testing.T) {
	title(t.Name())
	config.Clean()
	conf := map[string]interface{}{
		"miner.device":       "any",
		"miner.snapshot_len": 1500,
		"miner.promiscuous":  "1",
	}
	if err := config.LoadForTest(conf); err != nil {
		t.Errorf(err.Error())
	}
	if err := InitConfig(); err == nil {
		t.Errorf("An error should occur")
	}

	if !promiscuous {
		t.Errorf("Promiscuous mode must be activated")
	}
}

func TestInitWithoutKey(t *testing.T) {
	title(t.Name())
	config.Clean()
	conf := map[string]interface{}{
		"miner.device":       "any",
		"miner.snapshot_len": 1500,
	}
	if err := config.LoadForTest(conf); err != nil {
		t.Errorf(err.Error())
	}
	if err := InitConfig(); err == nil {
		t.Errorf("An error should occur")
	}

}
