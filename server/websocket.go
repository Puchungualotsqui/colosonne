package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocketServer struct {
	Rooms *RoomManager
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// In production, restrict this to your domain.
		return true
	},
}

func (s *WebSocketServer) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "could not upgrade websocket", http.StatusBadRequest)
		return
	}

	client := &Client{
		Conn: conn,
		Name: "Guest",
		Send: make(chan ServerMessage, 32),
	}

	go writePump(client)
	readPump(s, client)
}

func writePump(c *Client) {
	defer c.Conn.Close()

	for msg := range c.Send {
		if err := c.Conn.WriteJSON(msg); err != nil {
			return
		}
	}
}

func readPump(s *WebSocketServer, c *Client) {
	defer c.Conn.Close()

	for {
		_, raw, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}

		var msg ClientMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			c.Send <- ServerMessage{Type: "error", Data: "invalid message"}
			continue
		}

		if err := s.handleMessage(c, msg); err != nil {
			c.Send <- ServerMessage{Type: "error", Data: err.Error()}
		}
	}
}
