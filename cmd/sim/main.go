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

	// Friendly opening market.
	// 2 players * 3 cards each, plus a few tactical cards.
	gs.Market = marketSlots(
		engine.DraftItem{Kind: engine.DraftTile, Biome: engine.Forest},
		engine.DraftItem{Kind: engine.DraftTile, Biome: engine.Mountain},
		engine.DraftItem{Kind: engine.DraftTile, Biome: engine.Plain},
		engine.DraftItem{Kind: engine.DraftTile, Biome: engine.Ruins},
		engine.DraftItem{Kind: engine.DraftTile, Biome: engine.Forest},
		engine.DraftItem{Kind: engine.DraftTile, Biome: engine.Mountain},
	)

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
			handlePlace(reader, gs, rng)

		case engine.PhaseBuild:
			handleBuild(reader, gs)
		}
	}

	fmt.Println("\nGame over.")
	printState(gs)
	printWinner(gs)
}

func handlePick(reader *bufio.Reader, gs *engine.GameState) {
	playerId := gs.CurrentPlayer
	player := mustPlayer(gs, playerId)

	fmt.Printf("\nPlayer %d, pick a market item. Cards drafted: %d/%d\n",
		playerId,
		len(player.Hand),
		engine.CardsPerDraftTurn,
	)

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

func handlePlace(reader *bufio.Reader, gs *engine.GameState, rng *rand.Rand) {
	playerId := gs.CurrentPlayer
	player := mustPlayer(gs, playerId)

	fmt.Printf("\nPlayer %d, use your drafted cards.\n", playerId)

	if len(player.Hand) == 0 {
		fmt.Println("No cards in hand. Passing place phase.")
		_ = gs.PassPlace(playerId)
		return
	}

	printHand(player)

	for {
		fmt.Println("\nChoose a hand index to use.")
		fmt.Println("Type 'p' to pass if no card has a valid use.")
		fmt.Println("Type 'd <index>' to discard a specific unusable card.")

		line := askLine(reader, "Command: ")
		parts := strings.Fields(line)

		if len(parts) == 0 {
			continue
		}

		if parts[0] == "p" || parts[0] == "pass" {
			err := gs.PassPlace(playerId)
			if err != nil {
				fmt.Println("Cannot pass:", err)
				continue
			}
			return
		}

		if parts[0] == "d" || parts[0] == "discard" {
			if len(parts) != 2 {
				fmt.Println("Usage: d <handIndex>")
				continue
			}

			index, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("Hand index must be a number.")
				continue
			}

			err = gs.DiscardDraftItem(playerId, index)
			if err != nil {
				fmt.Println("Cannot discard:", err)
				continue
			}

			return
		}

		handIndex, err := strconv.Atoi(parts[0])
		if err != nil {
			fmt.Println("Enter a hand index, p, or d <index>.")
			continue
		}

		player = mustPlayer(gs, playerId)
		if handIndex < 0 || handIndex >= len(player.Hand) {
			fmt.Println("Invalid hand index.")
			continue
		}

		item := player.Hand[handIndex]
		err = useHandItem(reader, gs, rng, playerId, handIndex, item)
		if err != nil {
			fmt.Println("Could not use card:", err)
			continue
		}

		return
	}
}

func useHandItem(
	reader *bufio.Reader,
	gs *engine.GameState,
	rng *rand.Rand,
	playerId engine.PlayerId,
	handIndex int,
	item engine.DraftItem,
) error {
	switch item.Kind {
	case engine.DraftTile:
		x, y := askCoords(reader)
		return gs.PlaceTileFromHand(playerId, handIndex, x, y)

	case engine.DraftStructure:
		x, y := askCoords(reader)
		return gs.UseDraftItem(playerId, handIndex, x, y, 0, rng)

	case engine.DraftAction:
		switch item.Action {
		case engine.Expansion:
			return gs.UseDraftItem(playerId, handIndex, 0, 0, 0, rng)

		case engine.Raid:
			target := engine.PlayerId(askInt(reader, "Target player ID: "))
			return gs.UseDraftItem(playerId, handIndex, 0, 0, target, rng)

		case engine.Harvest, engine.Reinforce:
			x, y := askCoords(reader)
			return gs.UseDraftItem(playerId, handIndex, x, y, 0, rng)

		default:
			return fmt.Errorf("unknown action")
		}

	default:
		return fmt.Errorf("unknown draft item")
	}
}

func handleBuild(reader *bufio.Reader, gs *engine.GameState) {
	playerId := gs.CurrentPlayer

	fmt.Printf("\nPlayer %d, build phase.\n", playerId)
	fmt.Println("1. Build Outpost")
	fmt.Println("2. Build Settlement")
	fmt.Println("3. Upgrade Outpost to City")
	fmt.Println("4. Build Blockade")
	fmt.Println("5. Buy Floodworks")
	fmt.Println("6. Use Flood Token")
	fmt.Println("7. Pass")

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
			err := gs.BuildSettlement(playerId, x, y)
			if err != nil {
				fmt.Println("Could not build settlement:", err)
				continue
			}
			return

		case 3:
			x, y := askCoords(reader)
			err := gs.UpgradeCity(playerId, x, y)
			if err != nil {
				fmt.Println("Could not upgrade city:", err)
				continue
			}
			return

		case 4:
			x, y := askCoords(reader)
			err := gs.BuildBlockade(playerId, x, y)
			if err != nil {
				fmt.Println("Could not build blockade:", err)
				continue
			}
			return

		case 5:
			err := gs.BuyFloodworks(playerId)
			if err != nil {
				fmt.Println("Could not buy Floodworks:", err)
				continue
			}
			return

		case 6:
			x, y := askCoords(reader)
			err := gs.UseFloodToken(playerId, x, y)
			if err != nil {
				fmt.Println("Could not use Flood Token:", err)
				continue
			}

			fmt.Println("Flood Token used. You may still perform your build action.")
			printState(gs)

		case 7:
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
		fmt.Printf(
			"  Player %d | VP:%d | Wood:%d Stone:%d Grain:%d Relic:%d | Flood:%d | Floodworks:%d\n",
			p.Id,
			gs.VictoryPoints(p.Id),
			p.Resources[engine.Wood],
			p.Resources[engine.Stone],
			p.Resources[engine.Grain],
			p.Resources[engine.Relic],
			p.FloodTokens,
			p.FloodworksBought,
		)

		printHand(&p)
	}
}

func printHand(p *engine.Player) {
	if len(p.Hand) == 0 {
		fmt.Println("    Hand: empty")
		return
	}

	fmt.Println("    Hand:")
	for i, item := range p.Hand {
		fmt.Printf("      [%d] %s\n", i, describeDraftItem(item))
	}
}

func printMarket(gs *engine.GameState) {
	fmt.Println("\nMarket:")

	for i, item := range gs.Market {
		if item == nil {
			fmt.Printf("  [%d] empty\n", i)
			continue
		}

		fmt.Printf("  [%d] %s\n", i, describeDraftItem(*item))
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

		blockade := "-"
		if t.HasBlockade {
			blockade = fmt.Sprintf("P%d", t.BlockadeOwner)
		}

		fmt.Printf(
			"  (%2d,%2d) %-12s | Owner:%-4s | Structure:%-10s Owner:%d %-8s | Blockade:%-3s | Influence:%v Temp:%v\n",
			t.X,
			t.Y,
			describeBiome(t.Biome),
			owner,
			describeStructure(t.Structure),
			t.StructureOwner,
			active,
			blockade,
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

func askLine(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)

	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Input error:", err)
		return ""
	}

	return strings.TrimSpace(line)
}

func askInt(reader *bufio.Reader, prompt string) int {
	for {
		line := askLine(reader, prompt)

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
		line := askLine(reader, "Coordinates x y: ")

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

func marketSlots(items ...engine.DraftItem) []*engine.DraftItem {
	out := make([]*engine.DraftItem, engine.MarketSize)

	for i := 0; i < len(items) && i < engine.MarketSize; i++ {
		item := items[i]
		out[i] = &item
	}

	return out
}
