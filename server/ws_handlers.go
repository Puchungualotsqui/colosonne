package server

import (
	"errors"
)

func (s *WebSocketServer) handleMessage(c *Client, msg ClientMessage) error {
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
			},
		}

		room.BroadcastState()
		return nil

	case "pick":
		room, err := s.requireRoom(c)
		if err != nil {
			return err
		}

		payload, err := decodeData[PickPayload](msg.Data)
		if err != nil {
			return err
		}

		room.mu.Lock()
		err = room.Game.PickMarketItem(c.PlayerID, payload.MarketIndex)
		room.mu.Unlock()

		if err != nil {
			return err
		}

		room.BroadcastState()
		return nil

	case "place_tile":
		room, err := s.requireRoom(c)
		if err != nil {
			return err
		}

		payload, err := decodeData[PlaceTilePayload](msg.Data)
		if err != nil {
			return err
		}

		room.mu.Lock()
		err = room.Game.PlaceTile(c.PlayerID, payload.X, payload.Y)
		room.mu.Unlock()

		if err != nil {
			return err
		}

		room.BroadcastState()
		return nil

	case "use_draft":
		room, err := s.requireRoom(c)
		if err != nil {
			return err
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

		payload, err := decodeData[BuildPayload](msg.Data)
		if err != nil {
			return err
		}

		room.mu.Lock()

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

		room.BroadcastState()
		return nil

	default:
		return errors.New("unknown message type")
	}
}

func (s *WebSocketServer) requireRoom(c *Client) (*Room, error) {
	if c.RoomID == "" {
		return nil, errors.New("client is not in a room")
	}

	room, ok := s.Rooms.GetRoom(c.RoomID)
	if !ok {
		return nil, errors.New("room not found")
	}

	if room.Game == nil {
		return nil, errors.New("game has not started")
	}

	return room, nil
}
