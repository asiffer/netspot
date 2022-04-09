package exporter

import (
	"fmt"
	"net/http"
	"time"

	"github.com/asiffer/netspot/config"
	"github.com/gorilla/websocket"
)

type WebsocketClient struct{}

// WebsocketEndpoint manages connections of a single endpoint
type WebsocketEndpoint struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	control    chan uint8
}

// NewWebsocketEndpoint inits a new structure
func NewWebsocketEndpoint() *WebsocketEndpoint {
	return &WebsocketEndpoint{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		control:    make(chan uint8),
	}
}

func (wse *WebsocketEndpoint) clear() {
	for k := range wse.clients {
		delete(wse.clients, k)
	}
}

func (wse *WebsocketEndpoint) run() {
	for {
		select {
		case <-wse.control:
			return
		case conn := <-wse.register:
			// fmt.Println("REGISTER:", conn)
			wse.clients[conn] = true
		case conn := <-wse.unregister:
			if _, ok := wse.clients[conn]; ok {
				conn.Close()
				delete(wse.clients, conn)
			}
		case bytes := <-wse.broadcast:
			for conn := range wse.clients {
				// fmt.Println("CONN:", conn)
				if conn == nil {
					continue
				}

				err := conn.WriteMessage(websocket.TextMessage, bytes)
				// if err := conn.WriteMessage(websocket.TextMessage, bytes); err != nil {
				if err != nil {
					fmt.Println(err)
					conn.Close()
					delete(wse.clients, conn)
				}
			}
		}
	}
}

// Websocket is the exporting module
type Websocket struct {
	router   *http.ServeMux
	upgrader *websocket.Upgrader
	data     *WebsocketEndpoint
	alarms   *WebsocketEndpoint
	server   *http.Server
}

var upgrader = websocket.Upgrader{
	// this function always return true to avoid 403 error when Origin header is bad
	CheckOrigin: func(r *http.Request) bool { return true },
}

func init() {
	Register(&Websocket{
		router:   nil,
		upgrader: nil,
		data:     NewWebsocketEndpoint(),
		alarms:   NewWebsocketEndpoint(),
		server:   nil,
	})
	RegisterParameter("websocket.data", false, "Activate data endpoint")
	RegisterParameter("websocket.alarm", false, "Activate alarm endpoint")
	RegisterParameter("websocket.endpoint", "localhost:11001", "Websocket server")
}

// Main functions =========================================================== //
// ========================================================================== //
// ========================================================================== //

// Name returns the name of the exporter
func (ws *Websocket) Name() string {
	return "websocket"
}

// registerDataClient is called at the /data enspoint
func (ws *Websocket) registerDataClient(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err == nil {
		ws.data.register <- conn
	}
}

// registerAlarmClient is called at the /alarm endpoint
func (ws *Websocket) registerAlarmClient(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err == nil {
		ws.alarms.register <- conn
	}
}

// Init prepares the module from the config
func (ws *Websocket) Init() error {

	withDataEndpoint := config.MustBool("exporter.websocket.data")
	withAlarmEndpoint := config.MustBool("exporter.websocket.alarm")

	// do not load in this case
	if !(withDataEndpoint || withAlarmEndpoint) {
		return nil
	}

	ws.router = http.NewServeMux()
	ws.upgrader = &upgrader

	if withDataEndpoint {
		ws.router.HandleFunc("/data", ws.registerDataClient)
	}

	if withAlarmEndpoint {
		ws.router.HandleFunc("/alarm", ws.registerAlarmClient)
	}

	// compute endpoint from config
	addr, err := config.GetString("exporter.websocket.endpoint")
	if err != nil {
		return err
	}

	ws.server = &http.Server{
		Addr:    addr,
		Handler: ws.router,
	}

	return Load(ws.Name())
}

// Start starts the server that will wait for incoming connections
func (ws *Websocket) Start(series string) error {
	// start bith endpoints
	go ws.data.run()
	go ws.alarms.run()
	// start the server
	go ws.server.ListenAndServe()
	return nil
}

// Write sends data to the clients
func (ws *Websocket) Write(t time.Time, data map[string]float64) error {
	bytes := []byte(jsonifyWithTime(t, data))
	ws.data.broadcast <- bytes
	return nil
}

// Warn sends alarms to the clients
func (ws *Websocket) Warn(t time.Time, s *SpotAlert) error {
	bytes := []byte(s.toJSONwithTime(t))
	ws.alarms.broadcast <- bytes
	return nil
}

// Close stops the server and its connections
func (ws *Websocket) Close() error {
	// close the server
	if err := ws.server.Close(); err != nil {
		return err
	}
	// stop endpoints (goroutines)
	ws.data.control <- 0
	ws.alarms.control <- 0
	// flush endpoint clients
	ws.data.clear()
	ws.data.clear()
	// nullify other attrs
	ws.router = nil
	ws.upgrader = nil
	ws.server = nil
	return nil
}
