package server

import (
	"errors"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"

	"webGameGo/engine"
)

type Client struct {
	ID        string
	SessionID string
	Conn      WSConn
	RoomID    string
	PlayerID  engine.PlayerId
	UserID    *int64
	Name      string
	IsGuest   bool
	Role      string // "player" or "spectator"
	Send      chan ServerMessage

	closeOnce sync.Once
}

func (c *Client) CloseSend() {
	c.closeOnce.Do(func() {
		close(c.Send)
	})
}

type WSConn interface {
	WriteJSON(v any) error
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
}

type RoomStatus string

const (
	RoomStatusLobby   RoomStatus = "lobby"
	RoomStatusPlaying RoomStatus = "playing"
	RoomStatusEnded   RoomStatus = "ended"
)

type RoomSettings struct {
	MaxPlayers int  `json:"maxPlayers"`
	Spectators bool `json:"spectators"`
}

type RoomPlayer struct {
	PlayerID engine.PlayerId `json:"playerId"`
	ClientID string          `json:"clientId"`
	UserID   *int64          `json:"userId,omitempty"`
	Name     string          `json:"name"`
	IsGuest  bool            `json:"isGuest"`
	Ready    bool            `json:"ready"`
	IsHost   bool            `json:"isHost"`
}

type RoomSpectator struct {
	ClientID string `json:"clientId"`
	UserID   *int64 `json:"userId,omitempty"`
	Name     string `json:"name"`
	IsGuest  bool   `json:"isGuest"`
}

type Room struct {
	ID       string
	Status   RoomStatus
	Game     *engine.GameState
	Settings RoomSettings
	HostID   string
	RNG      *rand.Rand

	Clients    map[engine.PlayerId]*Client
	Spectators map[string]*Client
	Ready      map[engine.PlayerId]bool

	mu sync.Mutex
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
		ID:     id,
		Status: RoomStatusLobby,
		Game:   nil,
		Settings: RoomSettings{
			MaxPlayers: 2,
			Spectators: true,
		},
		HostID:     "",
		RNG:        rand.New(rand.NewSource(time.Now().UnixNano())),
		Clients:    make(map[engine.PlayerId]*Client),
		Spectators: make(map[string]*Client),
		Ready:      make(map[engine.PlayerId]bool),
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

	c.RoomID = r.ID

	// Same session reconnecting/rejoining: return existing player seat.
	// This is checked before rejecting non-lobby joins.
	for playerID, existing := range r.Clients {
		if existing.SessionID != "" && existing.SessionID == c.SessionID {
			c.RoomID = r.ID
			c.PlayerID = playerID
			c.Role = "player"

			r.Clients[playerID] = c

			if r.HostID == existing.ID {
				r.HostID = c.ID
			}

			existing.CloseSend()
			_ = existing.Conn.Close()

			log.Printf(
				"[room:%s] reconnect player client=%s oldClient=%s session=%s player=%d",
				r.ID,
				c.ID,
				existing.ID,
				c.SessionID,
				playerID,
			)

			return nil
		}
	}

	for clientID, existing := range r.Spectators {
		if existing.SessionID != "" && existing.SessionID == c.SessionID {
			c.RoomID = r.ID
			c.PlayerID = 0
			c.Role = "spectator"

			delete(r.Spectators, clientID)
			r.Spectators[c.ID] = c

			existing.CloseSend()
			_ = existing.Conn.Close()

			log.Printf(
				"[room:%s] reconnect spectator client=%s oldClient=%s session=%s name=%s",
				r.ID,
				c.ID,
				existing.ID,
				c.SessionID,
				c.Name,
			)

			return nil
		}
	}

	if r.Status != RoomStatusLobby {
		return errors.New("game already started")
	}

	if len(r.Clients) < r.Settings.MaxPlayers {
		var playerID engine.PlayerId = 1
		for {
			if _, exists := r.Clients[playerID]; !exists {
				break
			}
			playerID++
		}

		c.PlayerID = playerID
		c.Role = "player"
		r.Clients[playerID] = c
		r.Ready[playerID] = false

		if r.HostID == "" {
			r.HostID = c.ID
		}

		log.Printf(
			"[room:%s] add player client=%s session=%s player=%d name=%s",
			r.ID,
			c.ID,
			c.SessionID,
			c.PlayerID,
			c.Name,
		)

		return nil
	}

	if !r.Settings.Spectators {
		return errors.New("room is full")
	}

	c.PlayerID = 0
	c.Role = "spectator"
	r.Spectators[c.ID] = c

	log.Printf(
		"[room:%s] add spectator client=%s session=%s name=%s",
		r.ID,
		c.ID,
		c.SessionID,
		c.Name,
	)

	return nil
}

func (r *Room) AddSpectator(c *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.Settings.Spectators {
		return errors.New("spectators are disabled")
	}

	c.RoomID = r.ID
	c.PlayerID = 0
	c.Role = "spectator"

	r.Spectators[c.ID] = c

	log.Printf(
		"[room:%s] add spectator client=%s session=%s name=%s",
		r.ID,
		c.ID,
		c.SessionID,
		c.Name,
	)

	return nil
}

func (r *Room) RemoveClient(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	hostLeft := c.ID == r.HostID

	if c.Role == "player" && c.PlayerID != 0 {
		current, ok := r.Clients[c.PlayerID]
		if ok && current == c {
			delete(r.Clients, c.PlayerID)
			delete(r.Ready, c.PlayerID)
		}

		if hostLeft {
			r.promoteHostLocked()
		}

		return
	}

	if c.ID != "" {
		current, ok := r.Spectators[c.ID]
		if ok && current == c {
			delete(r.Spectators, c.ID)
		}
	}

	if hostLeft {
		r.promoteHostLocked()
	}
}

func (r *Room) promoteHostLocked() {
	if r.HostID != "" {
		for _, client := range r.Clients {
			if client.ID == r.HostID {
				return
			}
		}
	}

	r.HostID = ""

	var lowestPlayerID engine.PlayerId = 0
	var nextHost *Client

	for playerID, client := range r.Clients {
		if nextHost == nil || playerID < lowestPlayerID {
			lowestPlayerID = playerID
			nextHost = client
		}
	}

	if nextHost != nil {
		r.HostID = nextHost.ID

		for playerID := range r.Ready {
			r.Ready[playerID] = false
		}
	}
}

func (r *Room) Broadcast(msg ServerMessage) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for playerID, client := range r.Clients {
		select {
		case client.Send <- msg:
		default:
			client.CloseSend()
			delete(r.Clients, playerID)
			delete(r.Ready, playerID)
		}
	}

	for clientID, client := range r.Spectators {
		select {
		case client.Send <- msg:
		default:
			client.CloseSend()
			delete(r.Spectators, clientID)
		}
	}
}

func (r *Room) BroadcastState() {
	r.mu.Lock()

	roomID := r.ID
	status := r.Status
	game := r.Game
	settings := r.Settings
	hostID := r.HostID

	players := make([]RoomPlayer, 0, len(r.Clients))
	for playerID, client := range r.Clients {
		players = append(players, RoomPlayer{
			PlayerID: playerID,
			ClientID: client.ID,
			UserID:   client.UserID,
			Name:     client.Name,
			IsGuest:  client.IsGuest,
			Ready:    r.Ready[playerID],
			IsHost:   client.ID == hostID,
		})
	}

	spectators := make([]RoomSpectator, 0, len(r.Spectators))
	for _, client := range r.Spectators {
		spectators = append(spectators, RoomSpectator{
			ClientID: client.ID,
			UserID:   client.UserID,
			Name:     client.Name,
			IsGuest:  client.IsGuest,
		})
	}

	buildCosts := buildCostsForGame(game)

	r.mu.Unlock()

	currentPhase := any(nil)
	currentPlayer := any(nil)
	round := any(nil)
	mapTiles := 0
	marketCards := 0

	if game != nil {
		currentPhase = game.CurrentPhase
		currentPlayer = game.CurrentPlayer
		round = game.Round
		mapTiles = len(game.Map)
		marketCards = len(game.Market)
	}

	log.Printf(
		"[room:%s] broadcast state status=%s players=%d spectators=%d game=%v phase=%v currentPlayer=%v round=%v tiles=%d market=%d",
		roomID,
		status,
		len(players),
		len(spectators),
		game != nil,
		currentPhase,
		currentPlayer,
		round,
		mapTiles,
		marketCards,
	)

	r.Broadcast(ServerMessage{
		Type: "room_state",
		Data: map[string]any{
			"roomId":     roomID,
			"status":     status,
			"settings":   settings,
			"players":    players,
			"spectators": spectators,
			"game":       game,
			"buildCosts": buildCosts,
		},
	})
}

func (r *Room) SetReady(c *Client, ready bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if c.Role != "player" || c.PlayerID == 0 {
		return errors.New("only players can ready")
	}

	if r.Status != RoomStatusLobby {
		return errors.New("game already started")
	}

	r.Ready[c.PlayerID] = ready

	log.Printf(
		"[room:%s] ready client=%s player=%d ready=%v",
		r.ID,
		c.ID,
		c.PlayerID,
		ready,
	)

	return nil
}

func (r *Room) CanStartLocked(c *Client) error {
	if r.Status != RoomStatusLobby {
		return errors.New("game already started")
	}

	if c.ID != r.HostID {
		return errors.New("only host can start the game")
	}

	if len(r.Clients) < 2 {
		return errors.New("need at least 2 players")
	}

	for playerID := range r.Clients {
		if !r.Ready[playerID] {
			return errors.New("all players must be ready")
		}
	}

	return nil
}

func (r *Room) StartGame(c *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.CanStartLocked(c); err != nil {
		return err
	}

	if r.RNG == nil {
		r.RNG = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	players := make([]engine.Player, 0, len(r.Clients))
	for playerID := range r.Clients {
		players = append(players, engine.Player{Id: playerID})
	}

	r.Game = engine.NewGameState(players, r.RNG)
	r.Status = RoomStatusPlaying

	log.Printf(
		"[room:%s] start game host=%s players=%d",
		r.ID,
		c.ID,
		len(r.Clients),
	)

	return nil
}

func (r *Room) Kick(c *Client, targetPlayerID engine.PlayerId) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if c.ID != r.HostID {
		return errors.New("only host can kick players")
	}

	if r.Status != RoomStatusLobby {
		return errors.New("cannot kick after game starts")
	}

	target, ok := r.Clients[targetPlayerID]
	if !ok {
		return errors.New("player not found")
	}

	if target.ID == r.HostID {
		return errors.New("host cannot kick themselves")
	}

	delete(r.Clients, targetPlayerID)
	delete(r.Ready, targetPlayerID)

	r.promoteHostLocked()

	target.RoomID = ""
	target.PlayerID = 0
	target.Role = ""

	select {
	case target.Send <- ServerMessage{
		Type: "kicked",
		Data: "you were kicked from the room",
	}:
	default:
	}

	return nil
}

func buildCostsForGame(game *engine.GameState) map[engine.PlayerId]BuildCostsResponse {
	out := make(map[engine.PlayerId]BuildCostsResponse)

	if game == nil {
		return out
	}

	for _, player := range game.Players {
		out[player.Id] = BuildCostsResponse{
			Outpost:    resourceCostResponse(game.OutpostCost(player.Id)),
			City:       resourceCostResponse(game.CityCost(player.Id)),
			Settlement: resourceCostResponse(game.SettlementCost(player.Id)),
			Blockade:   resourceCostResponse(game.BlockadeCost()),
			Floodworks: resourceCostResponse(game.FloodworksCost(player.Id)),
		}
	}

	return out
}

func resourceCostResponse(cost map[engine.Resource]uint) ResourceCostResponse {
	return ResourceCostResponse{
		Wood:  cost[engine.Wood],
		Stone: cost[engine.Stone],
		Grain: cost[engine.Grain],
		Relic: cost[engine.Relic],
	}
}
