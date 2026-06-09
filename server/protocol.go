package server

import (
	"encoding/json"

	"webGameGo/engine"
)

type ScoresByPlayerResponse map[engine.PlayerId]uint

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
	HandIndex int `json:"handIndex"`
	X         int `json:"x"`
	Y         int `json:"y"`
}

type UseDraftPayload struct {
	HandIndex      int             `json:"handIndex"`
	X              int             `json:"x"`
	Y              int             `json:"y"`
	TargetPlayerID engine.PlayerId `json:"targetPlayerId"`
}

type DiscardDraftPayload struct {
	HandIndex int `json:"handIndex"`
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

type ResourceCostResponse struct {
	Wood  uint `json:"wood"`
	Stone uint `json:"stone"`
	Grain uint `json:"grain"`
	Relic uint `json:"relic"`
}

type BuildCostsResponse struct {
	Outpost    ResourceCostResponse `json:"outpost"`
	City       ResourceCostResponse `json:"city"`
	Settlement ResourceCostResponse `json:"settlement"`
	Blockade   ResourceCostResponse `json:"blockade"`
	Floodworks ResourceCostResponse `json:"floodworks"`
}

func decodeData[T any](raw json.RawMessage) (T, error) {
	var payload T
	err := json.Unmarshal(raw, &payload)
	return payload, err
}
