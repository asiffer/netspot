package exporter

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/asiffer/netspot/config"
	"golang.org/x/net/websocket"
)

var websocketTestConfig = map[string]interface{}{
	"exporter.websocket.data":     true,
	"exporter.websocket.alarm":    true,
	"exporter.websocket.endpoint": "localhost:11001",
}

var websocketTestSpotAlert = SpotAlert{
	Status:      "UP_ALERT",
	Stat:        "STAT",
	Value:       rand.Float64(),
	Code:        1,
	Probability: 3.14e-8,
}

func getInitializedWebsocket() ExportingModule {
	s := available["websocket"]
	config.LoadForTest(websocketTestConfig)
	if err := s.Init(); err != nil {
		panic(err)
	}
	return s
}

func TestInitWebsocket(t *testing.T) {
	name := "websocket"
	s, exists := available[name]
	if !exists {
		t.Fatalf("Exporter with name '%s' does not exist", name)
	}

	if s.Name() != name {
		t.Errorf("Expected name '%s' but got '%s'", name, s.Name())
	}

	config.LoadForTest(websocketTestConfig)
	if err := s.Init(); err != nil {
		t.Error(err)
	}
	defer Unload(name)

	if !isLoaded(name) {
		t.Errorf("The exporter must be loaded, but is not... List of loaded modules: %v", loadedModules())
	}
}

func TestStartCloseWebsocket(t *testing.T) {
	s := getInitializedWebsocket()
	defer Unload(s.Name())
	if err := s.Start("test-websocket"); err != nil {
		t.Error(err)
	}
	time.Sleep(1 * time.Second)
	if err := s.Close(); err != nil {
		t.Error(err)
	}
}

func TestWebsocketWithDataClient(t *testing.T) {
	s := getInitializedWebsocket()
	defer s.Close()
	defer Unload(s.Name())

	if err := s.Start("test-websocket-with-data-client"); err != nil {
		t.Error(err)
	}

	time.Sleep(200 * time.Millisecond)
	url, ok := websocketTestConfig["exporter.websocket.endpoint"].(string)
	if !ok {
		t.Fatalf("Cannot convert 'exporter.websocket.endpoint' config to string")
	}
	conn, err := websocket.Dial(fmt.Sprintf("ws://%s/data", url), "", "http://localhost/")
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	key := "CTR"
	value := rand.Float64()
	data := map[string]float64{key: value}
	s.Write(time.Now(), data)

	time.Sleep(200 * time.Millisecond)
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		t.Error(err)
	}
	result := make(map[string]float64)
	if err := json.Unmarshal(buffer[:n], &result); err != nil {
		t.Error(err)
	}
	if math.Abs(result[key]-data[key]) > 1e-6 {
		t.Errorf("Values mismatch (err > 1e-6): %v != %v", result[key], data[key])
	}
	time.Sleep(50 * time.Millisecond)
}

func TestWebsocketWithAlarmClient(t *testing.T) {
	s := getInitializedWebsocket()
	defer s.Close()
	defer Unload(s.Name())

	if err := s.Start("test-websocket-with-alarm-client"); err != nil {
		t.Error(err)
	}

	time.Sleep(200 * time.Millisecond)
	url, ok := websocketTestConfig["exporter.websocket.endpoint"].(string)
	if !ok {
		t.Fatalf("Cannot convert 'exporter.websocket.endpoint' config to string")
	}
	conn, err := websocket.Dial(fmt.Sprintf("ws://%s/alarm", url), "", "http://localhost/")
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	s.Warn(time.Now(), &websocketTestSpotAlert)

	time.Sleep(200 * time.Millisecond)
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		t.Error(err)
	}
	result := SpotAlert{}
	if err := json.Unmarshal(buffer[:n], &result); err != nil {
		t.Error(err)
	}

	if result.Status != websocketTestSpotAlert.Status {
		t.Errorf("Bad status, got %v instead of %v",
			result.Status, websocketTestSpotAlert.Status)
	}

	if result.Stat != websocketTestSpotAlert.Stat {
		t.Errorf("Bad stat, got %v instead of %v",
			result.Stat, websocketTestSpotAlert.Stat)
	}

	if result.Code != websocketTestSpotAlert.Code {
		t.Errorf("Bad code, got %v instead of %v",
			result.Code, websocketTestSpotAlert.Code)
	}

	if result.Probability != websocketTestSpotAlert.Probability {
		t.Errorf("Bad probability, got %v instead of %v",
			result.Probability, websocketTestSpotAlert.Probability)
	}
	fmt.Println(result.Probability)
}
