package engine

import "errors"

func (gs *GameState) BuildOutpost(playerId PlayerId, x, y int) error {
	if err := gs.requireBuildTurn(playerId); err != nil {
		return err
	}

	tile := gs.TileAt(x, y)
	if tile == nil {
		return errors.New("tile not found")
	}

	if tile.Biome == River {
		return errors.New("cannot build outpost on river")
	}

	if tile.Structure != NoneStructure {
		return errors.New("tile already has a structure")
	}

	if tile.HasOwner && tile.Owner != playerId {
		return errors.New("cannot build outpost on enemy controlled tile")
	}

	cost := gs.OutpostCost(playerId)
	if err := gs.PayResources(playerId, cost); err != nil {
		return err
	}

	tile.Structure = Outpost
	tile.StructureOwner = playerId

	gs.PhaseCompleted()

	return nil
}

func (gs *GameState) BuildSettlement(playerId PlayerId, x, y int) error {
	if err := gs.requireBuildTurn(playerId); err != nil {
		return err
	}

	tile := gs.TileAt(x, y)
	if tile == nil {
		return errors.New("tile not found")
	}

	if tile.Biome == River {
		return errors.New("cannot build settlement on river")
	}

	if !gs.playerControlsTile(playerId, tile) {
		return errors.New("settlement requires your controlled tile")
	}

	if tile.Structure != NoneStructure {
		return errors.New("tile already has a structure")
	}

	if tile.HasBlockade {
		return errors.New("cannot build settlement on blockaded tile")
	}

	cost := gs.SettlementCost(playerId)
	if err := gs.PayResources(playerId, cost); err != nil {
		return err
	}

	tile.Structure = Settlement
	tile.StructureOwner = playerId

	gs.PhaseCompleted()

	return nil
}

func (gs *GameState) UpgradeCity(playerId PlayerId, x, y int) error {
	if err := gs.requireBuildTurn(playerId); err != nil {
		return err
	}

	tile := gs.TileAt(x, y)
	if tile == nil {
		return errors.New("tile not found")
	}

	if tile.Structure != Outpost || tile.StructureOwner != playerId {
		return errors.New("city upgrade requires your own outpost")
	}

	if !gs.StructureActive(tile) {
		return errors.New("cannot upgrade inactive outpost")
	}

	cost := gs.CityCost(playerId)
	if err := gs.PayResources(playerId, cost); err != nil {
		return err
	}

	tile.Structure = City
	tile.StructureOwner = playerId

	gs.PhaseCompleted()

	return nil
}

func (gs *GameState) BuildBlockade(playerId PlayerId, x, y int) error {
	if err := gs.requireBuildTurn(playerId); err != nil {
		return err
	}

	tile := gs.TileAt(x, y)
	if tile == nil {
		return errors.New("tile not found")
	}

	if tile.Biome == River {
		return errors.New("cannot blockade river")
	}

	if tile.HasBlockade {
		return errors.New("tile is already blockaded")
	}

	if tile.HasOwner && tile.Owner == playerId {
		return errors.New("cannot blockade your own tile")
	}

	if !gs.hasAdjacentControlledTile(playerId, x, y) {
		return errors.New("blockade must be adjacent to your controlled territory")
	}

	cost := gs.BlockadeCost()
	if err := gs.PayResources(playerId, cost); err != nil {
		return err
	}

	tile.HasBlockade = true
	tile.BlockadeOwner = playerId

	gs.PhaseCompleted()

	return nil
}

func (gs *GameState) BuyFloodworks(playerId PlayerId) error {
	if err := gs.requireBuildTurn(playerId); err != nil {
		return err
	}

	player, err := gs.playerById(playerId)
	if err != nil {
		return err
	}

	cost := gs.FloodworksCost(playerId)
	if err := gs.PayResources(playerId, cost); err != nil {
		return err
	}

	player.FloodworksBought++
	player.FloodTokens += 3

	gs.PhaseCompleted()

	return nil
}

func (gs *GameState) PassBuild(playerId PlayerId) error {
	if err := gs.requireBuildTurn(playerId); err != nil {
		return err
	}

	gs.PhaseCompleted()

	return nil
}

func (gs *GameState) requireBuildTurn(playerId PlayerId) error {
	if gs.CurrentPhase != PhaseBuild {
		return errors.New("not in build phase")
	}

	if gs.CurrentPlayer != playerId {
		return errors.New("not current player")
	}

	return nil
}
