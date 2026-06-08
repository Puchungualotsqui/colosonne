package main

import (
	"fmt"
	"math/rand"

	"webGameGo/engine"
)

func main() {
	rng := rand.New(rand.NewSource(7))

	players := []engine.Player{
		{Id: 1},
		{Id: 2},
	}

	gs := engine.NewGameState(players, rng)

	// Make the first round deterministic.
	// 2 players * 3 cards each = 6 drafted cards.
	gs.Market = []engine.DraftItem{
		{Kind: engine.DraftTile, Biome: engine.Forest},
		{Kind: engine.DraftTile, Biome: engine.Mountain},
		{Kind: engine.DraftTile, Biome: engine.Plain},
		{Kind: engine.DraftTile, Biome: engine.Ruins},
		{Kind: engine.DraftTile, Biome: engine.Forest},
		{Kind: engine.DraftTile, Biome: engine.Mountain},
	}

	fmt.Println("=== Initial Game State ===")
	printState(gs)

	fmt.Println("\n=== Round 1: Pick Phase ===")
	for gs.CurrentPhase == engine.PhasePick {
		playerId := gs.CurrentPlayer

		fmt.Printf("Player %d picks market item 0: %s\n",
			playerId,
			describeDraftItem(gs.Market[0]),
		)

		if err := gs.PickMarketItem(playerId, 0); err != nil {
			panic(err)
		}

		printState(gs)
	}

	fmt.Println("\n=== Round 1: Place + Build Phase ===")

	for gs.CurrentPhase == engine.PhasePlace || gs.CurrentPhase == engine.PhaseBuild {
		playerId := gs.CurrentPlayer

		switch gs.CurrentPhase {
		case engine.PhasePlace:
			player := mustPlayer(gs, playerId)

			if len(player.Hand) == 0 {
				fmt.Printf("Player %d has no hand cards. Passing place phase.\n", playerId)
				if err := gs.PassPlace(playerId); err != nil {
					panic(err)
				}
				continue
			}

			fmt.Printf("Player %d has %d cards to use.\n", playerId, len(player.Hand))

			for gs.CurrentPhase == engine.PhasePlace && gs.CurrentPlayer == playerId {
				player = mustPlayer(gs, playerId)
				if len(player.Hand) == 0 {
					break
				}

				item := player.Hand[0]
				fmt.Printf("Player %d uses hand item 0: %s\n", playerId, describeDraftItem(item))

				if item.Kind != engine.DraftTile {
					if err := gs.DiscardDraftItem(playerId, 0); err != nil {
						panic(err)
					}
					continue
				}

				x, y := firstPlacementCandidate(gs)
				fmt.Printf("Player %d places tile at (%d,%d)\n", playerId, x, y)

				if err := gs.PlaceTileFromHand(playerId, 0, x, y); err != nil {
					panic(err)
				}

				printState(gs)
			}

		case engine.PhaseBuild:
			fmt.Printf("Player %d passes build in round 1.\n", playerId)

			if err := gs.PassBuild(playerId); err != nil {
				panic(err)
			}

			printState(gs)
		}
	}

	fmt.Println("\n=== Round 2 Begins ===")
	printTurnOrders(gs)
}

func firstPlacementCandidate(gs *engine.GameState) (int, int) {
	for _, tile := range gs.Map {
		for _, n := range engine.HexNeighbors(tile.X, tile.Y) {
			if gs.TileAt(n[0], n[1]) == nil {
				return n[0], n[1]
			}
		}
	}

	panic("no placement candidate found")
}

func mustPlayer(gs *engine.GameState, id engine.PlayerId) *engine.Player {
	index, err := gs.PlayerIndex(id)
	if err != nil {
		panic(err)
	}
	return &gs.Players[index]
}

func printState(gs *engine.GameState) {
	fmt.Printf("\nPhase: %s | Current Player: %d | RoundFirstIndex: %d | TurnIndex: %d\n",
		describePhase(gs.CurrentPhase),
		gs.CurrentPlayer,
		gs.RoundFirstIndex,
		gs.TurnIndex,
	)

	printPlayers(gs)
	printMap(gs)
}

func printPlayers(gs *engine.GameState) {
	fmt.Println("Players:")

	for _, p := range gs.Players {
		hand := "empty"
		if len(p.Hand) > 0 {
			hand = describeHand(p.Hand)
		}

		fmt.Printf(
			"  Player %d | Hand: %s | Wood:%d Stone:%d Grain:%d Relic:%d | Flood:%d\n",
			p.Id,
			hand,
			p.Resources[engine.Wood],
			p.Resources[engine.Stone],
			p.Resources[engine.Grain],
			p.Resources[engine.Relic],
			p.FloodTokens,
		)
	}
}

func describeHand(hand []engine.DraftItem) string {
	out := ""

	for i, item := range hand {
		if i > 0 {
			out += ", "
		}

		out += fmt.Sprintf("[%d]%s", i, describeDraftItem(item))
	}

	return out
}

func printMap(gs *engine.GameState) {
	fmt.Println("Map:")

	for _, t := range gs.Map {
		owner := "none"
		if t.HasOwner {
			owner = fmt.Sprintf("%d", t.Owner)
		}

		blockade := "-"
		if t.HasBlockade {
			blockade = fmt.Sprintf("P%d", t.BlockadeOwner)
		}

		fmt.Printf(
			"  (%d,%d) %-12s | Owner:%s | Structure:%-10s StructureOwner:%d | Blockade:%s | Influence:%v\n",
			t.X,
			t.Y,
			describeBiome(t.Biome),
			owner,
			describeStructure(t.Structure),
			t.StructureOwner,
			blockade,
			t.Influence,
		)
	}
}

func printTurnOrders(gs *engine.GameState) {
	fmt.Println("Place order:")
	for _, id := range gs.PlaceOrder() {
		fmt.Printf("  Player %d\n", id)
	}

	fmt.Println("Pick order:")
	for _, id := range gs.PickOrder() {
		fmt.Printf("  Player %d\n", id)
	}
}

func describePhase(p engine.GamePhase) string {
	switch p {
	case engine.PhasePick:
		return "Pick"
	case engine.PhasePlace:
		return "Place"
	case engine.PhaseBuild:
		return "Build"
	default:
		return "UnknownPhase"
	}
}

func describeBiome(b engine.Biome) string {
	switch b {
	case engine.Forest:
		return "Forest"
	case engine.Mountain:
		return "Mountain"
	case engine.Plain:
		return "Plain"
	case engine.River:
		return "River"
	case engine.Ruins:
		return "Ruins"
	case engine.NoneBiome:
		return "NoneBiome"
	default:
		return "UnknownBiome"
	}
}

func describeResource(r engine.Resource) string {
	switch r {
	case engine.Wood:
		return "Wood"
	case engine.Stone:
		return "Stone"
	case engine.Grain:
		return "Grain"
	case engine.Relic:
		return "Relic"
	case engine.NoneResource:
		return "NoneResource"
	default:
		return "UnknownResource"
	}
}

func describeStructure(s engine.Structure) string {
	switch s {
	case engine.NoneStructure:
		return "None"
	case engine.Outpost:
		return "Outpost"
	case engine.City:
		return "City"
	case engine.Settlement:
		return "Settlement"
	case engine.Bridge:
		return "Bridge"
	case engine.Watchtower:
		return "Watchtower"
	default:
		return "UnknownStructure"
	}
}

func describeDraftItem(item engine.DraftItem) string {
	switch item.Kind {
	case engine.DraftTile:
		return fmt.Sprintf("Tile(%s)", describeBiome(item.Biome))

	case engine.DraftStructure:
		return fmt.Sprintf("Structure(%s)", describeStructure(item.Structure))

	case engine.DraftAction:
		return fmt.Sprintf("Action(%s)", describeAction(item.Action))

	default:
		return "UnknownDraftItem"
	}
}

func describeAction(a engine.Action) string {
	switch a {
	case engine.Harvest:
		return "Harvest"
	case engine.Reinforce:
		return "Reinforce"
	case engine.Expansion:
		return "Expansion"
	case engine.Raid:
		return "Raid"
	case engine.NoAction:
		return "NoAction"
	default:
		return "UnknownAction"
	}
}
