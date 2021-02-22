// client.go

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

const defaultAddress = "http://localhost:11000/"

// helpers
func checkResponse(path string, expectedStatusCode int, resp *http.Response, err error) error {
	if err != nil {
		return fmt.Errorf("Error while hitting %s endpoint: %v", path, err)
	}
	if resp.StatusCode != expectedStatusCode {
		var msg string
		if bodyBytes, err := ioutil.ReadAll(resp.Body); err != nil {
			msg = fmt.Sprintf("cannot decode body: %v", err)
		} else {
			msg = string(bodyBytes)
		}
		return fmt.Errorf("Error while hitting %s endpoint (%d): %s",
			path, resp.StatusCode, msg)
	}
	return nil
}

// NetspotClient is a basic HTTP client
type NetspotClient struct {
	url *url.URL
}

// NewClient creates a new netspot client
func NewClient(addr string) *NetspotClient {
	u, err := url.Parse(addr)
	if err != nil {
		return nil
	}
	return &NetspotClient{url: u}
}

// GetAddress returns the url of the endpoint
func (ns *NetspotClient) GetAddress() string {
	return fmt.Sprintf("http://%s:%s", ns.url.Hostname(), ns.url.Port())
}

// route formats API endpoint
func (ns *NetspotClient) route(endpoint string) string {
	return ns.GetAddress() + path.Join(ns.url.Path, endpoint)
}

// Ping checks the connection
func (ns *NetspotClient) Ping() error {
	_, err := http.Get(ns.route("/api/ping"))
	if err != nil {
		return fmt.Errorf("Error while hitting /api/ping endpoint: %v", err)
	}
	return nil
}

// GetConfig returns the netspot config
func (ns *NetspotClient) GetConfig() (map[string]interface{}, error) {
	// action
	resp, err := http.Get(ns.route("/api/config"))
	// check error
	if e := checkResponse("/api/config", http.StatusOK, resp, err); e != nil {
		return nil, e
	}
	// treat response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error while reading response body: %v", err)
	}
	out := make(map[string]interface{})
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("Error while unmarshalling response body: %v", err)
	}
	return out, nil
}

// PostConfig modifies the netspot config
func (ns *NetspotClient) PostConfig(m map[string]interface{}) error {
	// prepare payload
	buffer, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("Error while marshalling config: %v", err)
	}
	body := bytes.NewReader(buffer)
	resp, err := http.Post(ns.route("api/config"), "application/json", body)
	// check errors
	return checkResponse("/api/config", http.StatusCreated, resp, err)
}

// GetStats returns the available stats along with their description
func (ns *NetspotClient) GetStats() (map[string]string, error) {
	resp, err := http.Get(ns.route("/api/stats"))
	// check errors
	if e := checkResponse("/api/stats", http.StatusOK, resp, err); e != nil {
		return nil, e
	}
	// treat response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error while reading response body: %v", err)
	}
	out := make(map[string]string)
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("Error while unmarshalling response body: %v", err)
	}
	return out, nil
}

// GetDevices returns the available network interfaces
func (ns *NetspotClient) GetDevices() ([]string, error) {
	resp, err := http.Get(ns.route("/api/devices"))
	// check errors
	if e := checkResponse("/api/devices", http.StatusOK, resp, err); e != nil {
		return nil, e
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error while reading response body: %v", err)
	}
	out := make([]string, 0)
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("Error while unmarshalling response body: %v", err)
	}
	return out, nil
}

// Start starts netspot
func (ns *NetspotClient) Start() error {
	resp, err := http.PostForm(ns.route("/api/run"), url.Values{"action": {"start"}})
	// check errors
	return checkResponse("/api/run", http.StatusOK, resp, err)
}

// Stop stops netspot
func (ns *NetspotClient) Stop() error {
	resp, err := http.PostForm(ns.route("/api/run"), url.Values{"action": {"stop"}})
	// check errors
	return checkResponse("/api/run", http.StatusOK, resp, err)
}

// ========================================================================= //
// Extra helpers =========================================================== //
// ========================================================================= //

// SetDevice configures the device to analyze
func (ns *NetspotClient) SetDevice(device string) error {
	return ns.PostConfig(map[string]interface{}{"miner.device": device})
}
