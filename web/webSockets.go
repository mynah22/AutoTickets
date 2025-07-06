package web

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// used to safely iterate over websocket connections
type wsClients struct {
	sync.Mutex
	clients map[*websocket.Conn]bool
}

// WebSocket handler for tickets information
func (w *WebApp) handleWsTickets(c echo.Context) error {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	w.wsClients.Lock()
	defer w.wsClients.Unlock()
	w.wsClients.clients[conn] = true
	err = conn.WriteJSON(w.Tc.GetUnassignedTickets())
	if err != nil {
		conn.Close()
		delete(w.wsClients.clients, conn)
		return nil
	}

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
	return nil
}

// Websocket Message Sending

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
