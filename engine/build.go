package engine

import "errors"

func (gs *GameState) BuildOutpost(playerId PlayerId, x, y int) error {
	if gs.CurrentPhase != PhaseBuild {
		return errors.New("not in build phase")
	}

	if gs.CurrentPlayer != playerId {
		return errors.New("not current player")
	}

	tile := gs.TileAt(x, y)
	if tile == nil {
		return errors.New("tile not found")
	}

	if tile.Structure != NoneStructure {
		return errors.New("tile already has a structure")
	}

	// Important early-game rule:
	// You may build an outpost on an unowned tile or on your own controlled tile.
	if tile.HasOwner && tile.Owner != playerId {
		return errors.New("cannot build outpost on enemy controlled tile")
	}

	cost, err := StructureCost(Outpost)
	if err != nil {
		return err
	}

	if err := gs.PayResources(playerId, cost); err != nil {
		return err
	}

	tile.Structure = Outpost
	tile.StructureOwner = playerId

	gs.PhaseCompleted()
	return nil
}

func (gs *GameState) UpgradeCity(playerId PlayerId, x, y int) error {
	if gs.CurrentPhase != PhaseBuild {
		return errors.New("not in build phase")
	}

	if gs.CurrentPlayer != playerId {
		return errors.New("not current player")
	}

	tile := gs.TileAt(x, y)
	if tile == nil {
		return errors.New("tile not found")
	}

	if tile.Structure != Outpost || tile.StructureOwner != playerId {
		return errors.New("city upgrade requires your own outpost")
	}

	cost, err := StructureCost(City)
	if err != nil {
		return err
	}

	if err := gs.PayResources(playerId, cost); err != nil {
		return err
	}

	tile.Structure = City
	tile.StructureOwner = playerId

	gs.PhaseCompleted()
	return nil
}

func (gs *GameState) PassBuild(playerId PlayerId) error {
	if gs.CurrentPhase != PhaseBuild {
		return errors.New("not in build phase")
	}

	if gs.CurrentPlayer != playerId {
		return errors.New("not current player")
	}

	gs.PhaseCompleted()
	return nil
}
