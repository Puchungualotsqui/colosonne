package engine

func (gs *GameState) VictoryPoints(playerId PlayerId) uint {
	var points uint

	for i := range gs.Map {
		tile := &gs.Map[i]

		if tile.HasOwner && tile.Owner == playerId {
			points += 1
		}

		if tile.StructureOwner == playerId && tile.Structure == City && gs.StructureActive(tile) {
			points += 2
		}
	}

	return points
}

func (gs *GameState) IsGameOver() bool {
	if len(gs.Deck) > 0 || len(gs.Market) > 0 {
		return false
	}

	for _, player := range gs.Players {
		if player.Hand != nil {
			return false
		}
	}

	return true
}
