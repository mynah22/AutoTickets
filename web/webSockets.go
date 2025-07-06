package web

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// used to thread-safely manage several websocket connections
type wsClients struct {
	sync.Mutex
	clients map[*websocket.Conn]bool
}

// WebSocket handler for new connections
func (w *WebApp) handleWsTickets(c echo.Context) error {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	w.wsClients.Lock()
	w.wsClients.clients[conn] = true
	w.wsClients.Unlock()

	sm := statusMessage{
		Type:         "status",
		LastApiCheck: w.lastGoodApi.getTime().Format(time.RFC3339),
		IsActive:     w.serverParams.getActive(),
	}

	// send ticket and status message. If both succeed, listen for incoming messages
	// if incoming message has error, delete client from list and close connection
	if w.sendTicketMessage(conn) && w.sendStatusMessage(conn, sm) {
		go func() {
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					w.wsClients.Lock()
					delete(w.wsClients.clients, conn)
					w.wsClients.Unlock()
					conn.Close()
					break
				}
			}
		}()
	}

	return nil
}

// Broadcast tickets to all WebSocket clients
func (w *WebApp) broadcastTickets() {
	w.wsClients.Lock()
	defer w.wsClients.Unlock()
	for c := range w.wsClients.clients {
		err := c.WriteJSON(w.Tc.GetUnassignedTickets())
		if err != nil {
			c.Close()
			delete(w.wsClients.clients, c)
		}
	}
}

// send tickets to a single web socket client
func (w *WebApp) sendTicketMessage(conn *websocket.Conn) bool {
	err := conn.WriteJSON(w.Tc.GetUnassignedTickets())
	if err != nil {
		conn.Close()
		delete(w.wsClients.clients, conn)
		return false
	}
	return true
}

// status messages

// Status message struct for websocket broadcast
type statusMessage struct {
	Type         string `json:"type"`
	LastApiCheck string `json:"lastApiCheck"`
	IsActive     bool   `json:"isActive"`
}

// Broadcast status to all WebSocket clients every 10 minutes
func (w *WebApp) periodicallyBroadcastStatus() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		<-ticker.C
		w.broadcastStatus()
	}
}

// broadcasts status to a single websocket clients
func (w *WebApp) broadcastStatus() {
	sm := statusMessage{
		Type:         "status",
		LastApiCheck: w.lastGoodApi.getTime().Format(time.RFC3339),
		IsActive:     w.serverParams.getActive(),
	}
	w.wsClients.Lock()
	defer w.wsClients.Unlock()
	for conn := range w.wsClients.clients {
		w.sendStatusMessage(conn, sm)
	}
}

// sends status to a single websocket client
func (w *WebApp) sendStatusMessage(conn *websocket.Conn, sm statusMessage) bool {
	if err := conn.WriteJSON(sm); err != nil {
		conn.Close()
		delete(w.wsClients.clients, conn)
		return false
	}
	return true
}
