package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"webGameGo/engine"
)

func main() {
	rng := rand.New(rand.NewSource(7))

	players := []engine.Player{
		{Id: 1},
		{Id: 2},
	}

	gs := engine.NewGameState(players, rng)

	// Friendly opening market so the first round is playable.
	// After these are picked, the normal random deck refills the market.
	gs.Market = []engine.DraftItem{
		{Kind: engine.DraftTile, Biome: engine.Forest},
		{Kind: engine.DraftTile, Biome: engine.Mountain},
		{Kind: engine.DraftTile, Biome: engine.Plain},
		{Kind: engine.DraftTile, Biome: engine.River},
		{Kind: engine.DraftAction, Action: engine.Expansion},
		{Kind: engine.DraftUpgrade},
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Frontiers — 1v1 terminal simulation")
	fmt.Println("Hex coordinates use axial coordinates: X = q, Y = r.")
	fmt.Println("Example neighbors of (0,0): (1,0), (1,-1), (0,-1), (-1,0), (-1,1), (0,1).")

	for !gs.IsGameOver() {
		printState(gs)

		switch gs.CurrentPhase {
		case engine.PhasePick:
			handlePick(reader, gs)

		case engine.PhasePlace:
			handlePlace(reader, gs)

		case engine.PhaseBuild:
			handleBuild(reader, gs)

		case engine.PhaseRecount:
			fmt.Println("\nResolving influence and producing resources...")
			pressEnter(reader)
			gs.PhaseCompleted()
		}
	}

	fmt.Println("\nGame over.")
	printState(gs)
	printWinner(gs)
}

func handlePick(reader *bufio.Reader, gs *engine.GameState) {
	playerId := gs.CurrentPlayer

	fmt.Printf("\nPlayer %d, pick one market item.\n", playerId)
	printMarket(gs)

	for {
		index := askInt(reader, "Market index: ")

		err := gs.PickMarketItem(playerId, index)
		if err != nil {
			fmt.Println("Invalid pick:", err)
			continue
		}

		return
	}
}

func handlePlace(reader *bufio.Reader, gs *engine.GameState) {
	playerId := gs.CurrentPlayer
	player := mustPlayer(gs, playerId)

	fmt.Printf("\nPlayer %d, use your drafted item.\n", playerId)

	if player.Hand == nil {
		fmt.Println("No item in hand. Passing place phase.")
		_ = gs.PassPlace(playerId)
		return
	}

	fmt.Println("Hand:", describeDraftItem(*player.Hand))

	switch player.Hand.Kind {
	case engine.DraftTile:
		for {
			fmt.Println("Place tile, or type 'p' to discard/pass.")
			x, y, pass := askCoordsOrPass(reader)

			if pass {
				_ = gs.PassPlace(playerId)
				return
			}

			err := gs.PlaceTile(playerId, x, y)
			if err != nil {
				fmt.Println("Could not place tile:", err)
				continue
			}

			return
		}

	case engine.DraftAction:
		if player.Hand.Action == engine.Expansion {
			err := gs.BuildOrUseDraft(playerId, 0, 0)
			if err != nil {
				fmt.Println("Could not use Expansion:", err)
				_ = gs.PassPlace(playerId)
			}
			return
		}

		for {
			fmt.Println("Use action on a tile, or type 'p' to discard/pass.")
			x, y, pass := askCoordsOrPass(reader)

			if pass {
				_ = gs.PassPlace(playerId)
				return
			}

			err := gs.BuildOrUseDraft(playerId, x, y)
			if err != nil {
				fmt.Println("Could not use action:", err)
				continue
			}

			return
		}

	case engine.DraftUpgrade, engine.DraftStructure:
		for {
			fmt.Println("Use drafted item on a tile, or type 'p' to discard/pass.")
			x, y, pass := askCoordsOrPass(reader)

			if pass {
				_ = gs.PassPlace(playerId)
				return
			}

			err := gs.BuildOrUseDraft(playerId, x, y)
			if err != nil {
				fmt.Println("Could not use drafted item:", err)
				continue
			}

			return
		}

	default:
		fmt.Println("Unknown hand item. Passing.")
		_ = gs.PassPlace(playerId)
	}
}

func handleBuild(reader *bufio.Reader, gs *engine.GameState) {
	playerId := gs.CurrentPlayer

	fmt.Printf("\nPlayer %d, build phase.\n", playerId)
	fmt.Println("1. Build Outpost")
	fmt.Println("2. Upgrade Outpost to City")
	fmt.Println("3. Pass")

	for {
		choice := askInt(reader, "Choice: ")

		switch choice {
		case 1:
			x, y := askCoords(reader)
			err := gs.BuildOutpost(playerId, x, y)
			if err != nil {
				fmt.Println("Could not build outpost:", err)
				continue
			}
			return

		case 2:
			x, y := askCoords(reader)
			err := gs.UpgradeCity(playerId, x, y)
			if err != nil {
				fmt.Println("Could not upgrade city:", err)
				continue
			}
			return

		case 3:
			err := gs.PassBuild(playerId)
			if err != nil {
				fmt.Println("Could not pass:", err)
				continue
			}
			return

		default:
			fmt.Println("Invalid choice.")
		}
	}
}

func printState(gs *engine.GameState) {
	fmt.Println("\n==================================================")
	fmt.Printf("Round %d | Phase: %s | Current Player: %d\n",
		gs.Round,
		describePhase(gs.CurrentPhase),
		gs.CurrentPlayer,
	)

	fmt.Print("Place order: ")
	for _, id := range gs.PlaceOrder() {
		fmt.Printf("%d ", id)
	}
	fmt.Println()

	fmt.Print("Pick order:  ")
	for _, id := range gs.PickOrder() {
		fmt.Printf("%d ", id)
	}
	fmt.Println()

	printPlayers(gs)
	printMap(gs)
	fmt.Println("==================================================")
}

func printPlayers(gs *engine.GameState) {
	fmt.Println("\nPlayers:")

	for _, p := range gs.Players {
		hand := "empty"
		if p.Hand != nil {
			hand = describeDraftItem(*p.Hand)
		}

		fmt.Printf(
			"  Player %d | VP:%d | Hand:%s | Wood:%d Stone:%d Grain:%d\n",
			p.Id,
			gs.VictoryPoints(p.Id),
			hand,
			p.Resources[engine.Wood],
			p.Resources[engine.Stone],
			p.Resources[engine.Grain],
		)
	}
}

func printMarket(gs *engine.GameState) {
	fmt.Println("\nMarket:")

	for i, item := range gs.Market {
		fmt.Printf("  [%d] %s\n", i, describeDraftItem(item))
	}
}

func printMap(gs *engine.GameState) {
	fmt.Println("\nMap:")

	for _, t := range gs.Map {
		owner := "none"
		if t.HasOwner {
			owner = fmt.Sprintf("%d", t.Owner)
		}

		active := "-"
		if t.Structure != engine.NoneStructure {
			if gs.StructureActive(&t) {
				active = "active"
			} else {
				active = "inactive"
			}
		}

		fmt.Printf(
			"  (%2d,%2d) %-8s | Owner:%-4s | Structure:%-10s Owner:%d %-8s | Upgrade:%d | Influence:%v Temp:%v\n",
			t.X,
			t.Y,
			describeBiome(t.Biome),
			owner,
			describeStructure(t.Structure),
			t.StructureOwner,
			active,
			t.UpgradeLevel,
			t.Influence,
			t.TempInfluence,
		)
	}
}

func printWinner(gs *engine.GameState) {
	var bestPlayer engine.PlayerId
	var bestScore uint
	tied := false

	for _, p := range gs.Players {
		score := gs.VictoryPoints(p.Id)

		if score > bestScore {
			bestScore = score
			bestPlayer = p.Id
			tied = false
		} else if score == bestScore {
			tied = true
		}
	}

	if tied {
		fmt.Println("Result: tied game.")
		return
	}

	fmt.Printf("Winner: Player %d with %d victory points.\n", bestPlayer, bestScore)
}

func askInt(reader *bufio.Reader, prompt string) int {
	for {
		fmt.Print(prompt)

		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Input error:", err)
			continue
		}

		line = strings.TrimSpace(line)

		value, err := strconv.Atoi(line)
		if err != nil {
			fmt.Println("Please enter a number.")
			continue
		}

		return value
	}
}

func askCoords(reader *bufio.Reader) (int, int) {
	for {
		fmt.Print("Coordinates x y: ")

		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Input error:", err)
			continue
		}

		parts := strings.Fields(line)
		if len(parts) != 2 {
			fmt.Println("Enter two numbers, example: 1 0")
			continue
		}

		x, errX := strconv.Atoi(parts[0])
		y, errY := strconv.Atoi(parts[1])

		if errX != nil || errY != nil {
			fmt.Println("Coordinates must be numbers.")
			continue
		}

		return x, y
	}
}

func askCoordsOrPass(reader *bufio.Reader) (int, int, bool) {
	for {
		fmt.Print("Coordinates x y, or p: ")

		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Input error:", err)
			continue
		}

		line = strings.TrimSpace(line)

		if line == "p" || line == "pass" {
			return 0, 0, true
		}

		parts := strings.Fields(line)
		if len(parts) != 2 {
			fmt.Println("Enter two numbers, example: 1 0, or p to pass.")
			continue
		}

		x, errX := strconv.Atoi(parts[0])
		y, errY := strconv.Atoi(parts[1])

		if errX != nil || errY != nil {
			fmt.Println("Coordinates must be numbers.")
			continue
		}

		return x, y, false
	}
}

func pressEnter(reader *bufio.Reader) {
	fmt.Print("Press Enter to continue...")
	_, _ = reader.ReadString('\n')
}

func mustPlayer(gs *engine.GameState, id engine.PlayerId) *engine.Player {
	index, err := gs.PlayerIndex(id)
	if err != nil {
		panic(err)
	}
	return &gs.Players[index]
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
