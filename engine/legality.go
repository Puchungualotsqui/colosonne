package engine

func (gs *GameState) CanUseHand(playerId PlayerId) bool {
	player, err := gs.playerById(playerId)
	if err != nil {
		return false
	}

	for i := range player.Hand {
		if gs.CanUseHandItem(playerId, i) {
			return true
		}
	}

	return false
}

func (gs *GameState) CanUseHandItem(playerId PlayerId, handIndex int) bool {
	player, err := gs.playerById(playerId)
	if err != nil {
		return false
	}

	if handIndex < 0 || handIndex >= len(player.Hand) {
		return false
	}

	item := player.Hand[handIndex]

	switch item.Kind {
	case DraftTile:
		return gs.canPlaceAnyTile()

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

	case Watchtower:
		if tile.Biome == River {
			return false
		}

		if tile.HasOwner && tile.Owner != playerId {
			return false
		}

		return true

	case Outpost, City, Settlement:
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

	case Raid:
		return gs.hasRaidTarget(playerId)

	case Harvest:
		for i := range gs.Map {
			tile := &gs.Map[i]

			if !gs.playerControlsTile(playerId, tile) {
				continue
			}

			if !gs.StructureActive(tile) {
				continue
			}

			if tile.HasBlockade {
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
		for i := range gs.Map {
			if gs.canReceiveInfluence(&gs.Map[i]) {
				return true
			}
		}

		return false

	default:
		return false
	}
}

func (gs *GameState) hasRaidTarget(playerId PlayerId) bool {
	for i := range gs.Players {
		player := &gs.Players[i]

		if player.Id == playerId {
			continue
		}

		player.ensureResources()

		if player.Resources[Wood]+player.Resources[Stone]+player.Resources[Grain]+player.Resources[Relic] > 0 {
			return true
		}
	}

	return false
}
