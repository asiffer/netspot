// client_test.go

package client

import (
	"fmt"
	"testing"
	"time"

	"github.com/knadh/koanf/maps"
)

func find(elem string, slice []string) int {
	for i, v := range slice {
		if elem == v {
			return i
		}
	}
	return -1
}

func toStringSlice(slice []interface{}) []string {
	out := make([]string, len(slice))
	for i, s := range slice {
		out[i] = fmt.Sprintf("%v", s)
	}
	return out
}

func TestPing(t *testing.T) {
	ns := NewClient(defaultAddress)
	if err := ns.Ping(); err != nil {
		t.Error(err)
	}
}

func TestGetConfig(t *testing.T) {
	nc := NewClient(defaultAddress)
	conf, err := nc.GetConfig()
	if err != nil {
		t.Error(err)
	}
	t.Log(conf)
}

func TestPostConfig(t *testing.T) {
	nc := NewClient(defaultAddress)

	// post new config
	config0 := map[string]interface{}{
		"exporter.file.alarm": "/tmp/alarm.json",
		"exporter.file.data":  "/tmp/data.json",
		"analyzer.stats":      []string{"R_SYN", "PERF"},
	}
	if err := nc.PostConfig(maps.Unflatten(config0, ".")); err != nil {
		t.Fatal(err)
	}

	// check config
	newConfig, err := nc.GetConfig()
	if err != nil {
		t.Error(err)
	}
	newConfig, _ = maps.Flatten(newConfig, nil, ".")
	for k, v := range config0 {
		switch value := v.(type) {
		case []string:
			newArray := toStringSlice(newConfig[k].([]interface{}))
			for i := 0; i < len(value); i++ {
				if find(value[i], newArray) < 0 {
					t.Errorf("Bad config, expect %v, got %v", v, newConfig[k])
				}
			}
		default:
			if v != newConfig[k] {
				t.Errorf("Bad config, expect %v, got %v", v, newConfig[k])
			}
		}

	}
}

func TestGetStats(t *testing.T) {
	ns := NewClient(defaultAddress)
	stats, err := ns.GetStats()
	if err != nil {
		t.Error(err)
	}
	t.Log(stats)

	statsList := make([]string, 0)
	for s := range stats {
		statsList = append(statsList, s)
	}

	builtin := []string{
		"AVG_PKT_SIZE",
		"PERF",
		"R_ACK",
		"R_ARP",
		"R_DST_SRC",
		"R_DST_SRC_PORT",
		"R_ICMP",
		"R_IP",
		"R_SYN",
		"TRAFFIC",
	}
	for _, s := range builtin {
		if find(s, statsList) < 0 {
			t.Errorf("Stat %s not found", s)
		}
	}
}

func TestGetDevices(t *testing.T) {
	ns := NewClient(defaultAddress)
	devices, err := ns.GetDevices()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Devices: %v", devices)
	builtin := []string{
		"any",
		"lo",
	}
	for _, d := range builtin {
		if find(d, devices) < 0 {
			t.Errorf("Device %s not found", d)
		}
	}
}

func TestStartStop(t *testing.T) {
	nc := NewClient(defaultAddress)

	// post new config
	config := map[string]interface{}{
		"miner.device":           "lo",
		"analyzer.period":        "250ms",
		"exporter.console.data":  true,
		"exporter.console.alarm": false,
		"exporter.file.data":     nil,
		"exporter.file.alarm":    nil,
		"analyzer.stats":         []string{"TRAFFIC"},
	}
	if err := nc.PostConfig(maps.Unflatten(config, ".")); err != nil {
		t.Fatal(err)
	}

	// start
	if err := nc.Start(); err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	// stop
	if err := nc.Stop(); err != nil {
		t.Fatal(err)
	}
}

func TestSetDevice(t *testing.T) {
	nc := NewClient(defaultAddress)
	if err := nc.SetDevice("lo"); err != nil {
		t.Error(err)
	}

	if err := nc.SetDevice("xxxxx"); err == nil {
		t.Errorf("An error was expected")
	}
}
