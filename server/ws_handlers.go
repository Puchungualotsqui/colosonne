package server

import (
	"errors"
	"log"
)

func (s *WebSocketServer) handleMessage(c *Client, msg ClientMessage) error {
	log.Printf(
		"[cmd] type=%s client=%s room=%s role=%s player=%d",
		msg.Type,
		c.ID,
		c.RoomID,
		c.Role,
		c.PlayerID,
	)

	switch msg.Type {
	case "create_room":
		room := s.Rooms.CreateRoom()

		if err := room.AddClient(c); err != nil {
			return err
		}

		c.Send <- ServerMessage{
			Type: "room_created",
			Data: map[string]any{
				"roomId":   room.ID,
				"playerId": c.PlayerID,
				"name":     c.Name,
				"isGuest":  c.IsGuest,
				"role":     c.Role,
			},
		}

		room.BroadcastState()
		return nil

	case "join_room":
		payload, err := decodeData[JoinRoomPayload](msg.Data)
		if err != nil {
			return err
		}

		room, ok := s.Rooms.GetRoom(payload.RoomID)
		if !ok {
			return errors.New("room not found")
		}

		if payload.Name != "" {
			c.Name = payload.Name
		}

		if err := room.AddClient(c); err != nil {
			return err
		}

		c.Send <- ServerMessage{
			Type: "room_joined",
			Data: map[string]any{
				"roomId":   room.ID,
				"playerId": c.PlayerID,
				"name":     c.Name,
				"isGuest":  c.IsGuest,
				"role":     c.Role,
			},
		}

		room.BroadcastState()
		return nil

	case "spectate_room":
		payload, err := decodeData[JoinRoomPayload](msg.Data)
		if err != nil {
			return err
		}

		room, ok := s.Rooms.GetRoom(payload.RoomID)
		if !ok {
			return errors.New("room not found")
		}

		if payload.Name != "" {
			c.Name = payload.Name
		}

		room.AddSpectator(c)

		c.Send <- ServerMessage{
			Type: "room_spectating",
			Data: map[string]any{
				"roomId":   room.ID,
				"playerId": c.PlayerID,
				"name":     c.Name,
				"isGuest":  c.IsGuest,
				"role":     c.Role,
			},
		}

		room.BroadcastState()
		return nil

	case "pick":
		room, err := s.requireRoom(c)
		if err != nil {
			return err
		}

		if c.Role != "player" {
			return errors.New("spectators cannot perform game actions")
		}

		payload, err := decodeData[PickPayload](msg.Data)
		if err != nil {
			return err
		}

		room.mu.Lock()

		log.Printf(
			"[cmd:pick] room=%s player=%d marketIndex=%d phase=%d currentPlayer=%d",
			room.ID,
			c.PlayerID,
			payload.MarketIndex,
			room.Game.CurrentPhase,
			room.Game.CurrentPlayer,
		)

		err = room.Game.PickMarketItem(c.PlayerID, payload.MarketIndex)
		room.mu.Unlock()

		if err != nil {
			return err
		}

		log.Printf(
			"[cmd:pick:ok] room=%s player=%d newPhase=%d newCurrentPlayer=%d",
			room.ID,
			c.PlayerID,
			room.Game.CurrentPhase,
			room.Game.CurrentPlayer,
		)

		room.BroadcastState()
		return nil

	case "place_tile":
		room, err := s.requireRoom(c)
		if err != nil {
			return err
		}

		if c.Role != "player" {
			return errors.New("spectators cannot perform game actions")
		}

		payload, err := decodeData[PlaceTilePayload](msg.Data)
		if err != nil {
			return err
		}

		room.mu.Lock()

		log.Printf(
			"[cmd:place_tile] room=%s player=%d x=%d y=%d phase=%d currentPlayer=%d",
			room.ID,
			c.PlayerID,
			payload.X,
			payload.Y,
			room.Game.CurrentPhase,
			room.Game.CurrentPlayer,
		)

		err = room.Game.PlaceTile(c.PlayerID, payload.X, payload.Y)
		room.mu.Unlock()

		if err != nil {
			return err
		}

		log.Printf(
			"[cmd:place_tile:ok] room=%s player=%d newPhase=%d newCurrentPlayer=%d tiles=%d",
			room.ID,
			c.PlayerID,
			room.Game.CurrentPhase,
			room.Game.CurrentPlayer,
			len(room.Game.Map),
		)

		room.BroadcastState()
		return nil

	case "use_draft":
		room, err := s.requireRoom(c)
		if err != nil {
			return err
		}

		if c.Role != "player" {
			return errors.New("spectators cannot perform game actions")
		}

		payload, err := decodeData[UseDraftPayload](msg.Data)
		if err != nil {
			return err
		}

		room.mu.Lock()
		err = room.Game.BuildOrUseDraft(c.PlayerID, payload.X, payload.Y)
		room.mu.Unlock()

		if err != nil {
			return err
		}

		room.BroadcastState()
		return nil

	case "pass_place":
		room, err := s.requireRoom(c)
		if err != nil {
			return err
		}

		if c.Role != "player" {
			return errors.New("spectators cannot perform game actions")
		}

		room.mu.Lock()
		err = room.Game.PassPlace(c.PlayerID)
		room.mu.Unlock()

		if err != nil {
			return err
		}

		room.BroadcastState()
		return nil

	case "build":
		room, err := s.requireRoom(c)
		if err != nil {
			return err
		}

		if c.Role != "player" {
			return errors.New("spectators cannot perform game actions")
		}

		payload, err := decodeData[BuildPayload](msg.Data)
		if err != nil {
			return err
		}

		room.mu.Lock()

		log.Printf(
			"[cmd:build] room=%s player=%d action=%s x=%d y=%d phase=%d currentPlayer=%d",
			room.ID,
			c.PlayerID,
			payload.Action,
			payload.X,
			payload.Y,
			room.Game.CurrentPhase,
			room.Game.CurrentPlayer,
		)

		switch payload.Action {
		case "outpost":
			err = room.Game.BuildOutpost(c.PlayerID, payload.X, payload.Y)
		case "city":
			err = room.Game.UpgradeCity(c.PlayerID, payload.X, payload.Y)
		case "pass":
			err = room.Game.PassBuild(c.PlayerID)
		default:
			err = errors.New("unknown build action")
		}

		room.mu.Unlock()

		if err != nil {
			return err
		}

		log.Printf(
			"[cmd:build:ok] room=%s player=%d newPhase=%d newCurrentPlayer=%d",
			room.ID,
			c.PlayerID,
			room.Game.CurrentPhase,
			room.Game.CurrentPlayer,
		)

		room.BroadcastState()
		return nil

	case "set_ready":
		room, err := s.requireAnyRoom(c)
		if err != nil {
			return err
		}

		payload, err := decodeData[ReadyPayload](msg.Data)
		if err != nil {
			return err
		}

		if err := room.SetReady(c, payload.Ready); err != nil {
			return err
		}

		room.BroadcastState()
		return nil

	case "start_game":
		room, err := s.requireAnyRoom(c)
		if err != nil {
			return err
		}

		if err := room.StartGame(c); err != nil {
			return err
		}

		room.BroadcastState()
		return nil

	case "kick_player":
		room, err := s.requireAnyRoom(c)
		if err != nil {
			return err
		}

		payload, err := decodeData[KickPayload](msg.Data)
		if err != nil {
			return err
		}

		if err := room.Kick(c, payload.PlayerID); err != nil {
			return err
		}

		room.BroadcastState()
		return nil

	default:
		return errors.New("unknown message type")
	}
}

func (s *WebSocketServer) requireRoom(c *Client) (*Room, error) {
	room, err := s.requireAnyRoom(c)
	if err != nil {
		return nil, err
	}

	if room.Status != RoomStatusPlaying || room.Game == nil {
		return nil, errors.New("game has not started")
	}

	return room, nil
}

func (s *WebSocketServer) requireAnyRoom(c *Client) (*Room, error) {
	if c.RoomID == "" {
		return nil, errors.New("client is not in a room")
	}

	room, ok := s.Rooms.GetRoom(c.RoomID)
	if !ok {
		return nil, errors.New("room not found")
	}

	return room, nil
}
