package server

import (
	"encoding/json"
	"errors"
	"math/rand"
	"sync"
	"time"

	"webGameGo/engine"

	"github.com/google/uuid"
)

type Client struct {
	Conn     WSConn
	RoomID   string
	PlayerID engine.PlayerId
	UserID   *int64
	Name     string
	Send     chan ServerMessage
}

type WSConn interface {
	WriteJSON(v any) error
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
}

type Room struct {
	ID      string
	Game    *engine.GameState
	Clients map[engine.PlayerId]*Client
	mu      sync.Mutex
}

type RoomManager struct {
	rooms map[string]*Room
	mu    sync.Mutex
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

func (rm *RoomManager) CreateRoom() *Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	id := shortRoomID()

	room := &Room{
		ID:      id,
		Game:    nil,
		Clients: make(map[engine.PlayerId]*Client),
	}

	rm.rooms[id] = room
	return room
}

func (rm *RoomManager) GetRoom(id string) (*Room, bool) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, ok := rm.rooms[id]
	return room, ok
}

func shortRoomID() string {
	return uuid.NewString()[:8]
}

func (r *Room) AddClient(c *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Clients) >= 2 {
		return errors.New("room is full")
	}

	var playerID engine.PlayerId = 1
	if _, exists := r.Clients[1]; exists {
		playerID = 2
	}

	c.PlayerID = playerID
	c.RoomID = r.ID
	r.Clients[playerID] = c

	if len(r.Clients) == 2 && r.Game == nil {
		players := []engine.Player{
			{Id: 1},
			{Id: 2},
		}

		// Use deterministic seed only for now. Later use crypto/rand or time.
		r.Game = engine.NewGameState(players, randSource())
	}

	return nil
}

func randSource() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

func (r *Room) Broadcast(msg ServerMessage) {
	for _, client := range r.Clients {
		select {
		case client.Send <- msg:
		default:
			close(client.Send)
			delete(r.Clients, client.PlayerID)
		}
	}
}

func (r *Room) BroadcastState() {
	if r.Game == nil {
		r.Broadcast(ServerMessage{
			Type: "room_waiting",
			Data: map[string]any{
				"roomId":  r.ID,
				"players": len(r.Clients),
			},
		})
		return
	}

	r.Broadcast(ServerMessage{
		Type: "state",
		Data: r.Game,
	})
}

func decodeData[T any](raw json.RawMessage) (T, error) {
	var payload T
	err := json.Unmarshal(raw, &payload)
	return payload, err
}
