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
)

type GameState struct {
	Players []Player
	Map     []Tile

	Deck   []DraftItem
	Market []DraftItem

	CurrentPlayer PlayerId
	CurrentPhase  GamePhase

	RoundFirstIndex int
	TurnIndex       int
	Round           uint
}

const CardsPerDraftTurn = 3
const MarketSize = 6

func NewGameState(players []Player, rng *rand.Rand) *GameState {
	deck := GenerateDraftDeck()
	ShuffleDraftItems(deck, rng)

	ShufflePlayers(players, rng)

	for i := range players {
		players[i].ensureResources()

		// One starting Wood means every random home biome can afford
		// a first Outpost after the first Settlement production.
		if players[i].Resources[Wood] == 0 &&
			players[i].Resources[Stone] == 0 &&
			players[i].Resources[Grain] == 0 &&
			players[i].Resources[Relic] == 0 {
			players[i].Resources[Wood] = 1
		}
	}

	gs := &GameState{
		Players:         players,
		Map:             []Tile{},
		Deck:            deck,
		Market:          []DraftItem{},
		CurrentPhase:    PhasePick,
		RoundFirstIndex: 0,
		TurnIndex:       0,
		Round:           1,
	}

	gs.createHomeTiles(rng)
	gs.RefillMarket()
	gs.beginPhase(PhasePick)

	return gs
}

func (gs *GameState) createHomeTiles(rng *rand.Rand) {
	homeBiomes := []Biome{Forest, Mountain, Plain}
	coords := homeCoordinates(len(gs.Players))

	for i := range gs.Players {
		playerID := gs.Players[i].Id
		coord := coords[i]

		biome := homeBiomes[rng.Intn(len(homeBiomes))]

		tile := NewTile(biome)
		tile.X = coord.x
		tile.Y = coord.y
		tile.HasOwner = true
		tile.Owner = playerID
		tile.Structure = Settlement
		tile.StructureOwner = playerID

		gs.Map = append(gs.Map, tile)
	}
}

type homeCoord struct {
	x int
	y int
}

func homeCoordinates(playerCount int) []homeCoord {
	coords := make([]homeCoord, 0, playerCount)

	if playerCount <= 0 {
		return coords
	}

	start := -2 * (playerCount - 1)

	for i := range playerCount {
		coords = append(coords, homeCoord{
			x: start + i*4,
			y: 0,
		})
	}

	return coords
}

func (gs *GameState) RequiredTileCardsInMarket() int {
	if len(gs.Players) == 0 {
		return 0
	}

	if gs.Round <= 2 {
		return 3
	}

	return 2
}

func (gs *GameState) CountTileCardsInMarket() int {
	count := 0

	for _, item := range gs.Market {
		if item.Kind == DraftTile {
			count++
		}
	}

	return count
}

func (gs *GameState) drawFirstTileFromDeck() (DraftItem, bool) {
	for i, item := range gs.Deck {
		if item.Kind == DraftTile {
			gs.Deck = append(gs.Deck[:i], gs.Deck[i+1:]...)
			return item, true
		}
	}

	return DraftItem{}, false
}

func (gs *GameState) drawTopFromDeck() (DraftItem, bool) {
	if len(gs.Deck) == 0 {
		return DraftItem{}, false
	}

	item := gs.Deck[0]
	gs.Deck = gs.Deck[1:]

	return item, true
}

func (gs *GameState) RefillMarket() {
	requiredTiles := gs.RequiredTileCardsInMarket()

	for len(gs.Market) < MarketSize && gs.CountTileCardsInMarket() < requiredTiles {
		item, ok := gs.drawFirstTileFromDeck()
		if !ok {
			break
		}

		gs.Market = append(gs.Market, item)
	}

	for len(gs.Market) < MarketSize {
		item, ok := gs.drawTopFromDeck()
		if !ok {
			break
		}

		gs.Market = append(gs.Market, item)
	}

	for gs.CountTileCardsInMarket() < requiredTiles {
		tileItem, ok := gs.drawFirstTileFromDeck()
		if !ok {
			break
		}

		replaced := false

		for i := len(gs.Market) - 1; i >= 0; i-- {
			if gs.Market[i].Kind != DraftTile {
				gs.Deck = append(gs.Deck, gs.Market[i])
				gs.Market[i] = tileItem
				replaced = true
				break
			}
		}

		if !replaced {
			gs.Market = append(gs.Market, tileItem)
		}
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
		// Refill only when a fresh draft turn sequence begins.
		// This matters after a full round ends, because the last picker may
		// have left the market partially empty.
		gs.RefillMarket()

		order := gs.PickOrder()
		if len(order) > 0 {
			gs.CurrentPlayer = order[0]
		}

	case PhasePlace, PhaseBuild:
		order := gs.PlaceOrder()
		if len(order) > 0 {
			gs.CurrentPlayer = order[0]
		}
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
		player, err := gs.playerById(gs.CurrentPlayer)
		if err == nil && len(player.Hand) < CardsPerDraftTurn && len(gs.Market) > 0 {
			return
		}

		order := gs.PickOrder()

		if gs.TurnIndex+1 < len(order) {
			gs.TurnIndex++
			gs.CurrentPlayer = order[gs.TurnIndex]

			// Refill only when the next player starts drafting.
			// This preserves the cards left by the previous player, and fills
			// only the empty market slots.
			gs.RefillMarket()

			return
		}

		gs.beginPhase(PhasePlace)

	case PhasePlace:
		player, err := gs.playerById(gs.CurrentPlayer)
		if err == nil && len(player.Hand) > 0 {
			return
		}

		gs.CurrentPhase = PhaseBuild

	case PhaseBuild:
		order := gs.PlaceOrder()

		if gs.TurnIndex+1 < len(order) {
			gs.TurnIndex++
			gs.CurrentPhase = PhasePlace
			gs.CurrentPlayer = order[gs.TurnIndex]
			return
		}

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

	const maxHandSize = 3

	if len(gs.Players[playerIndex].Hand) >= maxHandSize {
		return errors.New("player already has 3 drafted items")
	}

	item := gs.Market[marketIndex]
	gs.Players[playerIndex].Hand = append(gs.Players[playerIndex].Hand, item)

	gs.Market = append(gs.Market[:marketIndex], gs.Market[marketIndex+1:]...)

	if len(gs.Players[playerIndex].Hand) >= maxHandSize {
		gs.PhaseCompleted()
	}

	return nil
}
