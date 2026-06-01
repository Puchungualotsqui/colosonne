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
	// Both players will pick tile cards.
	gs.Market = []engine.DraftItem{
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

	placeCoords := map[engine.PlayerId][2]int{}
	nextCoordIndex := 0

	// Hex coordinates adjacent to the starting tile at 0,0.
	coords := [][2]int{
		{1, 0},
		{-1, 0},
	}

	for gs.CurrentPhase == engine.PhasePlace || gs.CurrentPhase == engine.PhaseBuild {
		playerId := gs.CurrentPlayer

		switch gs.CurrentPhase {
		case engine.PhasePlace:
			player := mustPlayer(gs, playerId)

			if player.Hand == nil {
				fmt.Printf("Player %d has no hand item. Passing place phase.\n", playerId)
				if err := gs.PassPlace(playerId); err != nil {
					panic(err)
				}
				continue
			}

			fmt.Printf("Player %d uses hand item: %s\n", playerId, describeDraftItem(*player.Hand))

			if player.Hand.Kind == engine.DraftTile {
				coord := coords[nextCoordIndex]
				nextCoordIndex++

				fmt.Printf("Player %d places tile at (%d,%d)\n", playerId, coord[0], coord[1])

				if err := gs.PlaceTile(playerId, coord[0], coord[1]); err != nil {
					panic(err)
				}

				placeCoords[playerId] = coord
			} else {
				// For non-tile draft items, try to use on 0,0.
				// If invalid, pass.
				err := gs.BuildOrUseDraft(playerId, 0, 0)
				if err != nil {
					fmt.Printf("Player %d could not use draft item: %v. Passing.\n", playerId, err)
					if err := gs.PassPlace(playerId); err != nil {
						panic(err)
					}
				}
			}

			printState(gs)

		case engine.PhaseBuild:
			coord, ok := placeCoords[playerId]
			if !ok {
				fmt.Printf("Player %d has no placed tile this round. Passing build.\n", playerId)
				if err := gs.PassBuild(playerId); err != nil {
					panic(err)
				}
				continue
			}

			fmt.Printf("Player %d builds Outpost at (%d,%d)\n", playerId, coord[0], coord[1])

			if err := gs.BuildOutpost(playerId, coord[0], coord[1]); err != nil {
				fmt.Printf("Player %d could not build Outpost: %v. Passing.\n", playerId, err)
				if err := gs.PassBuild(playerId); err != nil {
					panic(err)
				}
			}

			printState(gs)
		}
	}

	fmt.Println("\n=== Recount Phase ===")
	if gs.CurrentPhase == engine.PhaseRecount {
		fmt.Println("Resolving influence and producing resources...")
		gs.PhaseCompleted()
	}

	printState(gs)

	fmt.Println("\n=== Round 2 Begins ===")
	printTurnOrders(gs)
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
		if p.Hand != nil {
			hand = describeDraftItem(*p.Hand)
		}

		fmt.Printf(
			"  Player %d | Hand: %s | Wood:%d Stone:%d Grain:%d\n",
			p.Id,
			hand,
			p.Resources[engine.Wood],
			p.Resources[engine.Stone],
			p.Resources[engine.Grain],
		)
	}
}

func printMap(gs *engine.GameState) {
	fmt.Println("Map:")

	for _, t := range gs.Map {
		owner := "none"
		if t.HasOwner {
			owner = fmt.Sprintf("%d", t.Owner)
		}

		fmt.Printf(
			"  (%d,%d) %-8s | Owner:%s | Structure:%-10s StructureOwner:%d | Upgrade:%d | Influence:%v\n",
			t.X,
			t.Y,
			describeBiome(t.Biome),
			owner,
			describeStructure(t.Structure),
			t.StructureOwner,
			t.UpgradeLevel,
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
	case engine.PhaseRecount:
		return "Recount"
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
	case engine.Caravan:
		return "Caravan"
	case engine.Bridge:
		return "Bridge"
	case engine.Watchtower:
		return "Watchtower"
	case engine.Road:
		return "Road"
	default:
		return "UnknownStructure"
	}
}

func describeDraftItem(item engine.DraftItem) string {
	switch item.Kind {
	case engine.DraftTile:
		return fmt.Sprintf("Tile(%s)", describeBiome(item.Biome))

	case engine.DraftUpgrade:
		return "Upgrade"

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
	case engine.NoAction:
		return "NoAction"
	default:
		return "UnknownAction"
	}
}
