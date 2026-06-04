package server

import (
	"encoding/json"
	"webGameGo/engine"
)

type ClientMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type ServerMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type JoinRoomPayload struct {
	RoomID string `json:"roomId"`
	Name   string `json:"name"`
}

type PickPayload struct {
	MarketIndex int `json:"marketIndex"`
}

type PlaceTilePayload struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type UseDraftPayload struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type BuildPayload struct {
	Action string `json:"action"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
}

type ReadyPayload struct {
	Ready bool `json:"ready"`
}

type KickPayload struct {
	PlayerID engine.PlayerId `json:"playerId"`
}

type UpdateSettingsPayload struct {
	MaxPlayers int  `json:"maxPlayers"`
	Spectators bool `json:"spectators"`
}

func decodeData[T any](raw json.RawMessage) (T, error) {
	var payload T
	err := json.Unmarshal(raw, &payload)
	return payload, err
}
