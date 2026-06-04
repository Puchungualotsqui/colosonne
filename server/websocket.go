package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocketServer struct {
	App   *App
	Rooms *RoomManager
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// For development.
		// In production, restrict to your frontend domain.
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
		ID:      secureClientID(),
		Conn:    conn,
		Name:    "Guest",
		IsGuest: true,
		Role:    "",
		Send:    make(chan ServerMessage, 32),
	}

	session, err := s.App.currentSession(r)
	if err == nil {
		me, err := s.App.meFromSession(session)
		if err == nil {
			client.SessionID = session.Token
			client.Name = me.DisplayName
			client.UserID = me.UserID
			client.IsGuest = me.IsGuest
		}
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
	defer func() {
		if c.RoomID != "" {
			if room, ok := s.Rooms.GetRoom(c.RoomID); ok {
				room.RemoveClient(c)
				room.BroadcastState()
			}
		}

		c.CloseSend()
		_ = c.Conn.Close()
	}()

	for {
		_, raw, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}

		var msg ClientMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			select {
			case c.Send <- ServerMessage{Type: "error", Data: "invalid message"}:
			default:
			}
			continue
		}

		if err := s.handleMessage(c, msg); err != nil {
			select {
			case c.Send <- ServerMessage{Type: "error", Data: err.Error()}:
			default:
			}
		}
	}
}

func secureClientID() string {
	token, err := secureToken(16)
	if err != nil {
		return "client"
	}
	return token
}
