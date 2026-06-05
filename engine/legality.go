package engine

func (gs *GameState) CanUseHand(playerId PlayerId) bool {
	player, err := gs.playerById(playerId)
	if err != nil {
		return false
	}

	if player.Hand == nil {
		return false
	}

	item := player.Hand

	switch item.Kind {
	case DraftTile:
		return gs.canPlaceAnyTile()

	case DraftUpgrade:
		return gs.canUseUpgradeAnywhere(playerId)

	case DraftStructure:
		return gs.canUseStructureAnywhere(playerId, item.Structure)

	case DraftAction:
		return gs.canUseActionAnywhere(playerId, item.Action)

	default:
		return false
	}
}

func (gs *GameState) canPlaceAnyTile() bool {
	if len(gs.Map) == 0 {
		return true
	}

	for _, tile := range gs.Map {
		for _, n := range HexNeighbors(tile.X, tile.Y) {
			if gs.TileAt(n[0], n[1]) == nil {
				return true
			}
		}
	}

	return false
}

func (gs *GameState) canUseUpgradeAnywhere(playerId PlayerId) bool {
	for i := range gs.Map {
		tile := &gs.Map[i]

		if tile.Biome == River {
			continue
		}

		if !gs.playerControlsTile(playerId, tile) {
			continue
		}

		if tile.UpgradeLevel < 3 {
			return true
		}
	}

	return false
}

func (gs *GameState) canUseStructureAnywhere(playerId PlayerId, structure Structure) bool {
	for i := range gs.Map {
		tile := &gs.Map[i]

		if gs.canUseStructureOnTile(playerId, structure, tile) {
			return true
		}
	}

	return false
}

func (gs *GameState) canUseStructureOnTile(playerId PlayerId, structure Structure, tile *Tile) bool {
	if tile == nil {
		return false
	}

	if tile.Structure != NoneStructure {
		return false
	}

	switch structure {
	case Bridge:
		if tile.Biome != River {
			return false
		}

		return gs.hasAdjacentControlledTile(playerId, tile.X, tile.Y)

	case Road, Watchtower:
		if tile.Biome == River {
			return false
		}

		return gs.playerControlsTile(playerId, tile)

	case Outpost:
		if tile.Biome == River {
			return false
		}

		return !tile.HasOwner || tile.Owner == playerId

	case City:
		return false

	default:
		return false
	}
}

func (gs *GameState) hasAdjacentControlledTile(playerId PlayerId, x, y int) bool {
	for _, n := range HexNeighbors(x, y) {
		tile := gs.TileAt(n[0], n[1])
		if gs.playerControlsTile(playerId, tile) {
			return true
		}
	}

	return false
}

func (gs *GameState) canUseActionAnywhere(playerId PlayerId, action Action) bool {
	switch action {
	case Expansion:
		return true

	case Harvest:
		for i := range gs.Map {
			tile := &gs.Map[i]

			if !gs.playerControlsTile(playerId, tile) {
				continue
			}

			resource, err := tile.Biome.Resource()
			if err != nil {
				continue
			}

			if resource != NoneResource {
				return true
			}
		}

		return false

	case Reinforce:
		return len(gs.Map) > 0

	default:
		return false
	}
}
