package engine

import (
	"math/rand"
	"testing"
)

func TestFirstRoundPhaseFlowDoesNotStopOnRecount(t *testing.T) {
	rng := rand.New(rand.NewSource(1))

	players := []Player{
		{Id: 1},
		{Id: 2},
	}

	gs := NewGameState(players, rng)

	// Make the opening deterministic.
	// 2 players * 3 cards each = 6 drafted cards before Place phase.
	gs.Market = marketSlots(
		DraftItem{Kind: DraftTile, Biome: Forest},
		DraftItem{Kind: DraftTile, Biome: Mountain},
		DraftItem{Kind: DraftTile, Biome: Plain},
		DraftItem{Kind: DraftTile, Biome: Forest},
		DraftItem{Kind: DraftTile, Biome: Mountain},
		DraftItem{Kind: DraftTile, Biome: Plain},
	)

	if gs.CurrentPhase != PhasePick {
		t.Fatalf("expected PhasePick, got %v", gs.CurrentPhase)
	}

	// First picker drafts 3 cards.
	firstPicker := gs.CurrentPlayer
	for i := 0; i < CardsPerDraftTurn; i++ {
		index := firstFilledMarketIndex(t, gs)

		if err := gs.PickMarketItem(firstPicker, index); err != nil {
			t.Fatalf("first picker pick %d failed: %v", i+1, err)
		}
	}

	if gs.CurrentPhase != PhasePick {
		t.Fatalf("expected PhasePick after first player drafted 3 cards, got %v", gs.CurrentPhase)
	}

	secondPicker := gs.CurrentPlayer
	if secondPicker == firstPicker {
		t.Fatalf("expected current player to advance after first player drafted 3 cards")
	}

	// Second picker drafts 3 cards.
	for i := 0; i < CardsPerDraftTurn; i++ {
		index := firstFilledMarketIndex(t, gs)

		if err := gs.PickMarketItem(secondPicker, index); err != nil {
			t.Fatalf("second picker pick %d failed: %v", i+1, err)
		}
	}

	if gs.CurrentPhase != PhasePlace {
		t.Fatalf("expected PhasePlace after both players drafted, got %v", gs.CurrentPhase)
	}

	firstPlacer := gs.CurrentPlayer
	placeAllTilesForCurrentPlayer(t, gs, firstPlacer)

	if gs.CurrentPhase != PhaseBuild {
		t.Fatalf("expected PhaseBuild after first player used all draft cards, got %v", gs.CurrentPhase)
	}

	if err := gs.PassBuild(firstPlacer); err != nil {
		t.Fatalf("first pass build failed: %v", err)
	}

	if gs.CurrentPhase != PhasePlace {
		t.Fatalf("expected PhasePlace for second player, got %v", gs.CurrentPhase)
	}

	secondPlacer := gs.CurrentPlayer
	if secondPlacer == firstPlacer {
		t.Fatalf("expected current player to advance to second placer")
	}

	placeAllTilesForCurrentPlayer(t, gs, secondPlacer)

	if gs.CurrentPhase != PhaseBuild {
		t.Fatalf("expected PhaseBuild after second player used all draft cards, got %v", gs.CurrentPhase)
	}

	if err := gs.PassBuild(secondPlacer); err != nil {
		t.Fatalf("second pass build failed: %v", err)
	}

	if gs.CurrentPhase != PhasePick {
		t.Fatalf("expected automatic next round PhasePick, got %v", gs.CurrentPhase)
	}

	if gs.Round != 2 {
		t.Fatalf("expected round 2, got %d", gs.Round)
	}

	if gs.CurrentPlayer == 0 {
		t.Fatalf("expected a real current player after automatic recount")
	}
}

func marketSlots(items ...DraftItem) []*DraftItem {
	out := make([]*DraftItem, MarketSize)

	for i := 0; i < len(items) && i < MarketSize; i++ {
		item := items[i]
		out[i] = &item
	}

	return out
}

func firstFilledMarketIndex(t *testing.T, gs *GameState) int {
	t.Helper()

	for i, item := range gs.Market {
		if item != nil {
			return i
		}
	}

	t.Fatalf("no filled market slot found")
	return 0
}

func placeAllTilesForCurrentPlayer(t *testing.T, gs *GameState, playerId PlayerId) {
	t.Helper()

	for {
		player := mustTestPlayer(t, gs, playerId)
		if len(player.Hand) == 0 {
			return
		}

		if gs.CurrentPhase != PhasePlace {
			t.Fatalf("expected PhasePlace while placing cards, got %v", gs.CurrentPhase)
		}

		if gs.CurrentPlayer != playerId {
			t.Fatalf("expected current player %d, got %d", playerId, gs.CurrentPlayer)
		}

		if player.Hand[0].Kind != DraftTile {
			t.Fatalf("test expected only tile cards, got %+v", player.Hand[0])
		}

		x, y := firstPlacementCandidate(t, gs)

		if err := gs.PlaceTileFromHand(playerId, 0, x, y); err != nil {
			t.Fatalf("place tile failed for player %d at (%d,%d): %v", playerId, x, y, err)
		}
	}
}

func firstPlacementCandidate(t *testing.T, gs *GameState) (int, int) {
	t.Helper()

	for _, tile := range gs.Map {
		for _, n := range HexNeighbors(tile.X, tile.Y) {
			if gs.TileAt(n[0], n[1]) == nil {
				return n[0], n[1]
			}
		}
	}

	t.Fatalf("no placement candidate found")
	return 0, 0
}

func mustTestPlayer(t *testing.T, gs *GameState, id PlayerId) *Player {
	t.Helper()

	index, err := gs.PlayerIndex(id)
	if err != nil {
		t.Fatalf("player not found: %v", err)
	}

	return &gs.Players[index]
}
