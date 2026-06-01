package engine

func (gs *GameState) ResolveInfluence() {
	gs.clearInfluence()
	gs.applyStructureInfluence()
	gs.applyStructureInfluence()
	gs.resolveTileOwners()
	gs.produceResources()
	gs.clearTempInfluence()
}

func (gs *GameState) clearInfluence() {
	for i := range gs.Map {
		gs.Map[i].Influence = make(map[PlayerId]uint)
	}
}

func (gs *GameState) addInfluence(x, y int, playerId PlayerId, amount uint) {
	if amount == 0 {
		return
	}

	tile := gs.TileAt(x, y)
	if tile == nil {
		return
	}

	if tile.Influence == nil {
		tile.Influence = make(map[PlayerId]uint)
	}

	tile.Influence[playerId] += amount
}

func (gs *GameState) applyStructureInfluence() {
	for i := range gs.Map {
		tile := &gs.Map[i]

		if !gs.StructureActive(tile) {
			continue
		}

		owner := tile.StructureOwner
		upgradeBonus := tile.UpgradeLevel

		switch tile.Structure {
		case Outpost:
			gs.addInfluence(tile.X, tile.Y, owner, 2+upgradeBonus)

			for _, n := range HexNeighbors(tile.X, tile.Y) {
				gs.addInfluence(n[0], n[1], owner, 1)
			}

			if tile.Biome == Forest {
				for _, n := range HexRingRadius2(tile.X, tile.Y) {
					gs.addInfluence(n[0], n[1], owner, 1)
				}
			}

		case City:
			gs.addInfluence(tile.X, tile.Y, owner, 5+upgradeBonus)

			for _, n := range HexNeighbors(tile.X, tile.Y) {
				gs.addInfluence(n[0], n[1], owner, 2)
			}

		case Watchtower:
			gs.addInfluence(tile.X, tile.Y, owner, 1+upgradeBonus)

			for _, n := range HexNeighbors(tile.X, tile.Y) {
				gs.addInfluence(n[0], n[1], owner, 1)
			}

			for _, n := range HexRingRadius2(tile.X, tile.Y) {
				gs.addInfluence(n[0], n[1], owner, 1)
			}

		case Road:
			gs.addInfluence(tile.X, tile.Y, owner, 1+upgradeBonus)

			for _, n := range HexNeighbors(tile.X, tile.Y) {
				gs.addInfluence(n[0], n[1], owner, 1)
			}

		case Bridge:
			gs.addInfluence(tile.X, tile.Y, owner, 1+upgradeBonus)

			// Bridges are intentionally light for now.
			// Later, you can make them interact specially with river tiles.

		case Caravan:
			gs.addInfluence(tile.X, tile.Y, owner, 1+upgradeBonus)

			for _, n := range HexNeighbors(tile.X, tile.Y) {
				gs.addInfluence(n[0], n[1], owner, 1)
			}
		}
	}
}

func HexRingRadius2(x, y int) [][2]int {
	results := make([][2]int, 0, 12)

	for dx := -2; dx <= 2; dx++ {
		for dy := -2; dy <= 2; dy++ {
			dz := -dx - dy

			if abs(dx) == 2 || abs(dy) == 2 || abs(dz) == 2 {
				if abs(dx) <= 2 && abs(dy) <= 2 && abs(dz) <= 2 {
					results = append(results, [2]int{x + dx, y + dy})
				}
			}
		}
	}

	return results
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func (gs *GameState) resolveTileOwners() {
	for i := range gs.Map {
		tile := &gs.Map[i]

		bestPlayer := PlayerId(0)
		bestValue := uint(0)
		tied := false

		for playerId, value := range tile.Influence {
			if value > bestValue {
				bestPlayer = playerId
				bestValue = value
				tied = false
			} else if value == bestValue && value > 0 {
				tied = true
			}
		}

		if bestValue == 0 {
			tile.HasOwner = false
			tile.Owner = 0
			continue
		}

		if tied {
			// Tie rule:
			// ownership does not change.
			continue
		}

		tile.HasOwner = true
		tile.Owner = bestPlayer
	}
}

func (gs *GameState) produceResources() {
	for i := range gs.Map {
		tile := &gs.Map[i]

		if !tile.HasOwner {
			continue
		}

		resource, err := tile.Biome.Resource()
		if err != nil {
			continue
		}

		if resource == NoneResource {
			continue
		}

		amount := uint(1 + tile.UpgradeLevel)

		// Mountain rule:
		// If an owned Mountain belongs to a connected owned Mountain group of size 2+,
		// it produces +1 Stone.
		if tile.Biome == Mountain && gs.ConnectedOwnedMountainGroupSize(tile.X, tile.Y, tile.Owner) >= 2 {
			amount += 1
		}

		_ = gs.AddResource(tile.Owner, resource, amount)
	}
}

func (gs *GameState) StructureActive(tile *Tile) bool {
	if tile == nil {
		return false
	}

	if tile.Structure == NoneStructure || tile.StructureOwner == 0 {
		return false
	}

	// Early-game case: newly built outposts on unowned tiles are active.
	if !tile.HasOwner {
		return true
	}

	// Later case: captured structures become inactive.
	return tile.Owner == tile.StructureOwner
}

func (gs *GameState) applyTempInfluence() {
	for i := range gs.Map {
		tile := &gs.Map[i]

		for playerId, amount := range tile.TempInfluence {
			gs.addInfluence(tile.X, tile.Y, playerId, amount)
		}
	}
}

func (gs *GameState) clearTempInfluence() {
	for i := range gs.Map {
		gs.Map[i].TempInfluence = make(map[PlayerId]uint)
	}
}

func (gs *GameState) ConnectedOwnedMountainGroupSize(x, y int, owner PlayerId) int {
	start := gs.TileAt(x, y)
	if start == nil {
		return 0
	}

	if start.Biome != Mountain || !start.HasOwner || start.Owner != owner {
		return 0
	}

	type coord struct {
		x int
		y int
	}

	visited := make(map[coord]bool)
	stack := []coord{{x: x, y: y}}
	size := 0

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if visited[current] {
			continue
		}

		visited[current] = true

		tile := gs.TileAt(current.x, current.y)
		if tile == nil {
			continue
		}

		if tile.Biome != Mountain || !tile.HasOwner || tile.Owner != owner {
			continue
		}

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
