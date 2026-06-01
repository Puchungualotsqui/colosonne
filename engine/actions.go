package engine

import "errors"

func (gs *GameState) playerById(id PlayerId) (*Player, error) {
	index, err := gs.PlayerIndex(id)
	if err != nil {
		return nil, err
	}
	return &gs.Players[index], nil
}

func (gs *GameState) TileAt(x, y int) *Tile {
	for i := range gs.Map {
		if gs.Map[i].X == x && gs.Map[i].Y == y {
			return &gs.Map[i]
		}
	}
	return nil
}

func HexNeighbors(x, y int) [][2]int {
	return [][2]int{
		{x + 1, y},
		{x + 1, y - 1},
		{x, y - 1},
		{x - 1, y},
		{x - 1, y + 1},
		{x, y + 1},
	}
}

func (gs *GameState) HasAdjacentHexTile(x, y int) bool {
	for _, n := range HexNeighbors(x, y) {
		if gs.TileAt(n[0], n[1]) != nil {
			return true
		}
	}
	return false
}

func (gs *GameState) CanPlaceTileAt(x, y int) error {
	if gs.TileAt(x, y) != nil {
		return errors.New("tile already exists at position")
	}

	if len(gs.Map) > 0 && !gs.HasAdjacentHexTile(x, y) {
		return errors.New("tile must be adjacent to an existing hex tile")
	}

	return nil
}

func (gs *GameState) playerControlsTile(playerId PlayerId, tile *Tile) bool {
	return tile != nil && tile.HasOwner && tile.Owner == playerId
}

func (gs *GameState) PlaceTile(playerId PlayerId, x, y int) error {
	if gs.CurrentPhase != PhasePlace {
		return errors.New("not in place phase")
	}

	if gs.CurrentPlayer != playerId {
		return errors.New("not current player")
	}

	player, err := gs.playerById(playerId)
	if err != nil {
		return err
	}

	if player.Hand == nil {
		gs.PhaseCompleted()
		return nil
	}

	if player.Hand.Kind != DraftTile {
		return errors.New("current drafted item is not a tile")
	}

	if err := gs.CanPlaceTileAt(x, y); err != nil {
		return err
	}

	tile := NewTile(player.Hand.Biome)
	tile.X = x
	tile.Y = y

	// Plain rule:
	// If a player places a Plain adjacent to their active City,
	// that Plain is immediately owned by that player.
	if tile.Biome == Plain && gs.HasAdjacentActiveCityOwnedBy(playerId, x, y) {
		tile.HasOwner = true
		tile.Owner = playerId
	}

	gs.Map = append(gs.Map, tile)

	player.Hand = nil
	gs.PhaseCompleted()

	return nil
}

func (gs *GameState) BuildOrUseDraft(playerId PlayerId, x, y int) error {
	if gs.CurrentPhase != PhasePlace {
		return errors.New("draft item can only be used during place phase")
	}

	if gs.CurrentPlayer != playerId {
		return errors.New("not current player")
	}

	player, err := gs.playerById(playerId)
	if err != nil {
		return err
	}

	if player.Hand == nil {
		gs.PhaseCompleted()
		return nil
	}

	item := player.Hand

	switch item.Kind {
	case DraftTile:
		return errors.New("drafted item is a tile; use PlaceTile instead")

	case DraftStructure:
		if err := gs.useDraftStructure(playerId, item.Structure, x, y); err != nil {
			return err
		}

	case DraftUpgrade:
		if err := gs.useDraftUpgrade(playerId, x, y); err != nil {
			return err
		}

	case DraftAction:
		if err := gs.useDraftAction(playerId, item.Action, x, y); err != nil {
			return err
		}

	default:
		return errors.New("unknown draft item kind")
	}

	player.Hand = nil
	gs.PhaseCompleted()

	return nil
}

func (gs *GameState) PassPlace(playerId PlayerId) error {
	if gs.CurrentPhase != PhasePlace {
		return errors.New("not in place phase")
	}

	if gs.CurrentPlayer != playerId {
		return errors.New("not current player")
	}

	player, err := gs.playerById(playerId)
	if err != nil {
		return err
	}

	// Discard unusable drafted item.
	player.Hand = nil

	gs.PhaseCompleted()
	return nil
}

func (gs *GameState) HasAdjacentActiveCityOwnedBy(playerId PlayerId, x, y int) bool {
	for _, n := range HexNeighbors(x, y) {
		tile := gs.TileAt(n[0], n[1])
		if tile == nil {
			continue
		}

		if tile.Structure == City &&
			tile.StructureOwner == playerId &&
			gs.StructureActive(tile) {
			return true
		}
	}

	return false
}
