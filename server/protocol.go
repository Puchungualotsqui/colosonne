package server

import "encoding/json"

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
