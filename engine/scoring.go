package engine

import "slices"

type EndGameBonuses struct {
	LargestCityWinners            []PlayerId
	LargestCitySize               int
	LargestConnectedTerritoryWins []PlayerId
	LargestConnectedTerritorySize int
}

func (gs *GameState) VictoryPoints(playerId PlayerId) uint {
	var points uint

	for i := range gs.Map {
		tile := &gs.Map[i]

		if tile.HasOwner && tile.Owner == playerId && !gs.isUnbridgedRiver(tile) {
			points += 1
		}

		if tile.StructureOwner == playerId && tile.Structure == City && gs.StructureActive(tile) {
			points += 2
		}
	}

	bonuses := gs.CalculateEndGameBonuses()

	if slices.Contains(bonuses.LargestCityWinners, playerId) {
		points += 5
	}

	if slices.Contains(bonuses.LargestConnectedTerritoryWins, playerId) {
		points += 5
	}

	return points
}

func (gs *GameState) IsGameOver() bool {
	if len(gs.Deck) > 0 || gs.CountCardsInMarket() > 0 {
		return false
	}

	for _, player := range gs.Players {
		if len(player.Hand) > 0 {
			return false
		}
	}

	return true
}

func (gs *GameState) CalculateEndGameBonuses() EndGameBonuses {
	return EndGameBonuses{
		LargestCityWinners:            gs.playersWithLargestCity(3),
		LargestCitySize:               gs.largestCitySize(),
		LargestConnectedTerritoryWins: gs.playersWithLargestConnectedTerritory(7),
		LargestConnectedTerritorySize: gs.largestConnectedTerritorySize(),
	}
}

type coord struct {
	x int
	y int
}

func (gs *GameState) largestConnectedTerritorySize() int {
	best := 0

	for _, player := range gs.Players {
		size := gs.largestConnectedTerritoryFor(player.Id)
		if size > best {
			best = size
		}
	}

	return best
}

func (gs *GameState) playersWithLargestConnectedTerritory(minSize int) []PlayerId {
	best := 0
	winners := []PlayerId{}

	for _, player := range gs.Players {
		size := gs.largestConnectedTerritoryFor(player.Id)
		if size < minSize {
			continue
		}

		if size > best {
			best = size
			winners = []PlayerId{player.Id}
		} else if size == best && size > 0 {
			winners = append(winners, player.Id)
		}
	}

	return winners
}

func (gs *GameState) largestConnectedTerritoryFor(playerId PlayerId) int {
	visited := make(map[coord]bool)
	best := 0

	for i := range gs.Map {
		tile := &gs.Map[i]

		if !tile.HasOwner || tile.Owner != playerId {
			continue
		}

		if gs.isUnbridgedRiver(tile) {
			continue
		}

		start := coord{tile.X, tile.Y}
		if visited[start] {
			continue
		}

		size := gs.floodOwnedRegion(playerId, start, visited, func(t *Tile) bool {
			return t.HasOwner && t.Owner == playerId && !gs.isUnbridgedRiver(t)
		})

		if size > best {
			best = size
		}
	}

	return best
}

func (gs *GameState) largestCitySize() int {
	best := 0

	for _, player := range gs.Players {
		size := gs.largestCityFor(player.Id)
		if size > best {
			best = size
		}
	}

	return best
}

func (gs *GameState) playersWithLargestCity(minSize int) []PlayerId {
	best := 0
	winners := []PlayerId{}

	for _, player := range gs.Players {
		size := gs.largestCityFor(player.Id)
		if size < minSize {
			continue
		}

		if size > best {
			best = size
			winners = []PlayerId{player.Id}
		} else if size == best && size > 0 {
			winners = append(winners, player.Id)
		}
	}

	return winners
}

func (gs *GameState) largestCityFor(playerId PlayerId) int {
	visited := make(map[coord]bool)
	best := 0

	for i := range gs.Map {
		tile := &gs.Map[i]

		if !tile.HasOwner || tile.Owner != playerId {
			continue
		}

		if gs.isUnbridgedRiver(tile) {
			continue
		}

		start := coord{tile.X, tile.Y}
		if visited[start] {
			continue
		}

		hasCity := false

		size := gs.floodOwnedRegion(playerId, start, visited, func(t *Tile) bool {
			if !t.HasOwner || t.Owner != playerId || gs.isUnbridgedRiver(t) {
				return false
			}

			if t.Structure == City && t.StructureOwner == playerId && gs.StructureActive(t) {
				hasCity = true
			}

			return true
		})

		if hasCity && size > best {
			best = size
		}
	}

	return best
}

func (gs *GameState) floodOwnedRegion(
	playerId PlayerId,
	start coord,
	visited map[coord]bool,
	include func(*Tile) bool,
) int {
	stack := []coord{start}
	size := 0

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if visited[current] {
			continue
		}

		tile := gs.TileAt(current.x, current.y)
		if tile == nil {
			continue
		}

		if !include(tile) {
			continue
		}

		visited[current] = true
		size++

		for _, n := range HexNeighbors(current.x, current.y) {
			next := coord{x: n[0], y: n[1]}
			if !visited[next] {
				stack = append(stack, next)
			}
		}
	}

	return size
}
