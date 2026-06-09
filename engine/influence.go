package engine

func (gs *GameState) ResolveInfluence() {
	gs.clearInfluence()
	gs.applyStructureInfluence()
	gs.applyTempInfluence()
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

	if !gs.canReceiveInfluence(tile) {
		return
	}

	if tile.Influence == nil {
		tile.Influence = make(map[PlayerId]uint)
	}

	tile.Influence[playerId] += amount
}

func (gs *GameState) canReceiveInfluence(tile *Tile) bool {
	if tile == nil {
		return false
	}

	if tile.Biome == River && tile.Structure != Bridge {
		return false
	}

	return true
}

func (gs *GameState) isUnbridgedRiver(tile *Tile) bool {
	return tile != nil && tile.Biome == River && tile.Structure != Bridge
}

func (gs *GameState) applyStructureInfluence() {
	for i := range gs.Map {
		tile := &gs.Map[i]

		if !gs.StructureActive(tile) {
			continue
		}

		owner := tile.StructureOwner

		switch tile.Structure {
		case Settlement:
			// Pure production structure. Minimal control.
			gs.addInfluence(tile.X, tile.Y, owner, 1)

		case Outpost:
			gs.addInfluence(tile.X, tile.Y, owner, 2)

			for _, n := range HexNeighbors(tile.X, tile.Y) {
				gs.addInfluence(n[0], n[1], owner, 1)
			}

		case City:
			// Economy/defensive anchor, not an expansion tool.
			gs.addInfluence(tile.X, tile.Y, owner, 4)

		case Watchtower:
			// Strongest control structure. Draft-only.
			gs.addInfluence(tile.X, tile.Y, owner, 3)

			for _, n := range HexNeighbors(tile.X, tile.Y) {
				gs.addInfluence(n[0], n[1], owner, 2)
			}

			for _, n := range HexRingRadius2(tile.X, tile.Y) {
				if gs.canInfluenceDistance2(tile.X, tile.Y, n[0], n[1]) {
					gs.addInfluence(n[0], n[1], owner, 1)
				}
			}

		case Bridge:
			// Bridge makes a river tile controllable/crossable.
			gs.addInfluence(tile.X, tile.Y, owner, 1)

			for _, n := range HexNeighbors(tile.X, tile.Y) {
				gs.addInfluence(n[0], n[1], owner, 1)
			}
		}
	}
}

func (gs *GameState) canInfluenceDistance2(fromX, fromY, toX, toY int) bool {
	for _, a := range HexNeighbors(fromX, fromY) {
		for _, b := range HexNeighbors(toX, toY) {
			if a[0] != b[0] || a[1] != b[1] {
				continue
			}

			middle := gs.TileAt(a[0], a[1])
			if middle == nil {
				continue
			}

			if !gs.isUnbridgedRiver(middle) {
				return true
			}
		}
	}

	return false
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

		if gs.isUnbridgedRiver(tile) {
			tile.HasOwner = false
			tile.Owner = 0
			continue
		}

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

		if tile.HasBlockade {
			continue
		}

		if !gs.StructureActive(tile) {
			continue
		}

		switch tile.Structure {
		case Settlement:
			if err := gs.AddResource(tile.Owner, Wood, 1); err == nil {
				gs.EmitResourceGainFromTile(tile.Owner, Wood, 1, tile.X, tile.Y, NoAction)
			}

			if err := gs.AddResource(tile.Owner, Stone, 1); err == nil {
				gs.EmitResourceGainFromTile(tile.Owner, Stone, 1, tile.X, tile.Y, NoAction)
			}

			if err := gs.AddResource(tile.Owner, Grain, 1); err == nil {
				gs.EmitResourceGainFromTile(tile.Owner, Grain, 1, tile.X, tile.Y, NoAction)
			}

			resource, err := tile.Biome.Resource()
			if err == nil && resource != NoneResource {
				if err := gs.AddResource(tile.Owner, resource, 1); err == nil {
					gs.EmitResourceGainFromTile(tile.Owner, resource, 1, tile.X, tile.Y, NoAction)
				}
			}

		case Outpost:
			resource, err := tile.Biome.Resource()
			if err == nil && resource != NoneResource {
				if err := gs.AddResource(tile.Owner, resource, 1); err == nil {
					gs.EmitResourceGainFromTile(tile.Owner, resource, 1, tile.X, tile.Y, NoAction)
				}
			}

		case City:
			resource, err := tile.Biome.Resource()
			if err == nil && resource != NoneResource {
				if err := gs.AddResource(tile.Owner, resource, 2); err == nil {
					gs.EmitResourceGainFromTile(tile.Owner, resource, 2, tile.X, tile.Y, NoAction)
				}
			}
		}
	}
}

func (gs *GameState) StructureActive(tile *Tile) bool {
	if tile == nil {
		return false
	}

	if tile.Structure == NoneStructure || tile.StructureOwner == 0 {
		return false
	}

	if !tile.HasOwner {
		return true
	}

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
