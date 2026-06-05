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

	if gs.CurrentPhase != PhasePick {
		t.Fatalf("expected PhasePick, got %v", gs.CurrentPhase)
	}

	// Depending on shuffle/order this may be 1 or 2.
	firstPicker := gs.CurrentPlayer

	if err := gs.PickMarketItem(firstPicker, 0); err != nil {
		t.Fatalf("first pick failed: %v", err)
	}

	secondPicker := gs.CurrentPlayer

	if err := gs.PickMarketItem(secondPicker, 0); err != nil {
		t.Fatalf("second pick failed: %v", err)
	}

	if gs.CurrentPhase != PhasePlace {
		t.Fatalf("expected PhasePlace after both picks, got %v", gs.CurrentPhase)
	}

	firstPlacer := gs.CurrentPlayer

	if err := gs.PlaceTile(firstPlacer, 1, 0); err != nil {
		t.Fatalf("first place failed: %v", err)
	}

	if gs.CurrentPhase != PhaseBuild {
		t.Fatalf("expected PhaseBuild after first place, got %v", gs.CurrentPhase)
	}

	if err := gs.PassBuild(firstPlacer); err != nil {
		t.Fatalf("first pass build failed: %v", err)
	}

	if gs.CurrentPhase != PhasePlace {
		t.Fatalf("expected PhasePlace for second player, got %v", gs.CurrentPhase)
	}

	secondPlacer := gs.CurrentPlayer

	if err := gs.PlaceTile(secondPlacer, -1, 0); err != nil {
		t.Fatalf("second place failed: %v", err)
	}

	if gs.CurrentPhase != PhaseBuild {
		t.Fatalf("expected PhaseBuild after second place, got %v", gs.CurrentPhase)
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
		t.Fatalf("expected a real current player after recount")
	}
}
