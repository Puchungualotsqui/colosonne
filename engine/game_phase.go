package engine

import (
	"errors"
	"math/rand"
)

type GamePhase uint

const (
	PhasePick GamePhase = iota
	PhasePlace
	PhaseBuild
	PhaseRecount
)

type GameState struct {
	Players         []Player
	Map             []Tile
	Deck            []DraftItem
	Market          []DraftItem
	CurrentPlayer   PlayerId
	CurrentPhase    GamePhase
	RoundFirstIndex int
	TurnIndex       int
	Round           uint
}

func NewGameState(players []Player, rng *rand.Rand) *GameState {
	deck := GenerateDraftDeck()
	ShuffleDraftItems(deck, rng)

	startTile := NewTile(Plain)
	startTile.X = 0
	startTile.Y = 0

	for i := range players {
		if players[i].Resources == nil {
			players[i].Resources = map[Resource]uint{
				Wood:  2,
				Stone: 2,
				Grain: 2,
			}
		}
	}

	ShufflePlayers(players, rng)

	gs := &GameState{
		Players:         players,
		Map:             []Tile{startTile},
		Deck:            deck,
		Market:          []DraftItem{},
		CurrentPhase:    PhasePick,
		RoundFirstIndex: 0,
		TurnIndex:       0,
		Round:           1,
	}

	gs.RefillMarket(6)
	gs.beginPhase(PhasePick)
	return gs
}

func (gs *GameState) RefillMarket(size int) {
	for len(gs.Market) < size && len(gs.Deck) > 0 {
		item := gs.Deck[0]
		gs.Deck = gs.Deck[1:]
		gs.Market = append(gs.Market, item)
	}
}

func (gs *GameState) PlayerIndex(id PlayerId) (int, error) {
	for i := range gs.Players {
		if gs.Players[i].Id == id {
			return i, nil
		}
	}
	return 0, errors.New("player not found")
}

func (gs *GameState) PlaceOrder() []PlayerId {
	n := len(gs.Players)
	order := make([]PlayerId, 0, n)

	for i := range n {
		index := (gs.RoundFirstIndex + i) % n
		order = append(order, gs.Players[index].Id)
	}

	return order
}

func reversePlayerIds(ids []PlayerId) []PlayerId {
	out := make([]PlayerId, len(ids))
	for i := range ids {
		out[i] = ids[len(ids)-1-i]
	}
	return out
}

func (gs *GameState) PickOrder() []PlayerId {
	return reversePlayerIds(gs.PlaceOrder())
}

func (gs *GameState) beginPhase(phase GamePhase) {
	gs.CurrentPhase = phase
	gs.TurnIndex = 0

	switch phase {
	case PhasePick:
		order := gs.PickOrder()
		if len(order) > 0 {
			gs.CurrentPlayer = order[0]
		}

	case PhasePlace, PhaseBuild:
		order := gs.PlaceOrder()
		if len(order) > 0 {
			gs.CurrentPlayer = order[gs.TurnIndex]
		}

	case PhaseRecount:
		gs.CurrentPlayer = 0
	}
}

func (gs *GameState) advanceToNextRound() {
	if len(gs.Players) == 0 {
		return
	}

	gs.Round++
	gs.RoundFirstIndex = (gs.RoundFirstIndex + 1) % len(gs.Players)
	gs.beginPhase(PhasePick)
}

func (gs *GameState) PhaseCompleted() {
	if len(gs.Players) == 0 {
		return
	}

	switch gs.CurrentPhase {
	case PhasePick:
		order := gs.PickOrder()

		if gs.TurnIndex+1 < len(order) {
			gs.TurnIndex++
			gs.CurrentPlayer = order[gs.TurnIndex]
			return
		}

		gs.beginPhase(PhasePlace)

	case PhasePlace:
		gs.CurrentPhase = PhaseBuild

	case PhaseBuild:
		order := gs.PlaceOrder()

		if gs.TurnIndex+1 < len(order) {
			gs.TurnIndex++
			gs.CurrentPhase = PhasePlace
			gs.CurrentPlayer = order[gs.TurnIndex]
			return
		}

		gs.beginPhase(PhaseRecount)

	case PhaseRecount:
		gs.ResolveInfluence()
		gs.advanceToNextRound()
	}
}

func (gs *GameState) PickMarketItem(playerId PlayerId, marketIndex int) error {
	if gs.CurrentPhase != PhasePick {
		return errors.New("not in pick phase")
	}

	if gs.CurrentPlayer != playerId {
		return errors.New("not current player")
	}

	if marketIndex < 0 || marketIndex >= len(gs.Market) {
		return errors.New("invalid market index")
	}

	playerIndex, err := gs.PlayerIndex(playerId)
	if err != nil {
		return errors.New("player not found")
	}

	if gs.Players[playerIndex].Hand != nil {
		return errors.New("player already has item in hand")
	}

	item := gs.Market[marketIndex]
	gs.Players[playerIndex].Hand = &item

	gs.Market = append(gs.Market[:marketIndex], gs.Market[marketIndex+1:]...)
	gs.RefillMarket(6)

	gs.PhaseCompleted()
	return nil
}
