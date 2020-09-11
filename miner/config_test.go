package miner

import (
	"fmt"
	"netspot/config"
	"strings"
	"testing"
	"time"
)

func TestInitConfig(t *testing.T) {
	InitLogger()

	conf := map[string]interface{}{
		"miner.device":       hugePcap,
		"miner.snapshot_len": int32(65535),
		"miner.promiscuous":  false,
		"miner.timeout":      30 * time.Second,
	}
	if err := config.LoadForTest(conf); err != nil {
		t.Error(err)
	}
	if err := InitConfig(); err != nil {
		t.Errorf(err.Error())
	}

	raw := RawStatus()
	for key, value := range conf {
		k := strings.Replace(key, "miner.", "", -1)
		truth := fmt.Sprintf("%v", value)
		if raw[k] != truth {
			t.Errorf("Expecting %s, got %s", truth, raw[k])
		}
	}

	generic := GenericStatus()
	for key, value := range conf {
		k := strings.Replace(key, "miner.", "", -1)
		// truth := fmt.Sprintf("%v", value)
		if generic[k] != value {
			t.Errorf("Expecting %v, got %v", value, generic[k])
		}
	}

}
