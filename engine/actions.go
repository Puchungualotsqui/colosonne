package engine

import (
	"errors"
	"math/rand"
)

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
	player, err := gs.playerById(playerId)
	if err != nil {
		return err
	}

	for i, item := range player.Hand {
		if item.Kind == DraftTile {
			return gs.PlaceTileFromHand(playerId, i, x, y)
		}
	}

	return errors.New("player has no tile card in hand")
}

func (gs *GameState) PlaceTileFromHand(playerId PlayerId, handIndex int, x, y int) error {
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

	if handIndex < 0 || handIndex >= len(player.Hand) {
		return errors.New("invalid hand index")
	}

	item := player.Hand[handIndex]

	if item.Kind != DraftTile {
		return errors.New("selected drafted item is not a tile")
	}

	if err := gs.CanPlaceTileAt(x, y); err != nil {
		return err
	}

	tile := NewTile(item.Biome)
	tile.X = x
	tile.Y = y

	gs.Map = append(gs.Map, tile)

	player.Hand = removeDraftItem(player.Hand, handIndex)

	gs.PhaseCompleted()

	return nil
}

func (gs *GameState) BuildOrUseDraft(playerId PlayerId, x, y int) error {
	player, err := gs.playerById(playerId)
	if err != nil {
		return err
	}

	for i, item := range player.Hand {
		if item.Kind != DraftTile {
			return gs.UseDraftItem(playerId, i, x, y, 0, nil)
		}
	}

	return errors.New("player has no non-tile draft item in hand")
}

func (gs *GameState) UseDraftItem(
	playerId PlayerId,
	handIndex int,
	x int,
	y int,
	targetPlayerId PlayerId,
	rng *rand.Rand,
) error {
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

	if handIndex < 0 || handIndex >= len(player.Hand) {
		return errors.New("invalid hand index")
	}

	item := player.Hand[handIndex]

	switch item.Kind {
	case DraftTile:
		return errors.New("selected drafted item is a tile; use PlaceTileFromHand instead")

	case DraftStructure:
		if err := gs.useDraftStructure(playerId, item.Structure, x, y); err != nil {
			return err
		}

	case DraftAction:
		if err := gs.useDraftAction(playerId, item.Action, x, y, targetPlayerId, rng); err != nil {
			return err
		}

	default:
		return errors.New("unknown draft item kind")
	}

	player.Hand = removeDraftItem(player.Hand, handIndex)

	gs.PhaseCompleted()

	return nil
}

func (gs *GameState) DiscardDraftItem(playerId PlayerId, handIndex int) error {
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

	if handIndex < 0 || handIndex >= len(player.Hand) {
		return errors.New("invalid hand index")
	}

	if gs.CanUseHandItem(playerId, handIndex) {
		return errors.New("you must use this drafted item because it has a valid use")
	}

	player.Hand = removeDraftItem(player.Hand, handIndex)

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

	if gs.CanUseHand(playerId) {
		return errors.New("you must use your valid drafted items before passing")
	}

	player, err := gs.playerById(playerId)
	if err != nil {
		return err
	}

	player.Hand = nil

	gs.PhaseCompleted()

	return nil
}

func removeDraftItem(items []DraftItem, index int) []DraftItem {
	return append(items[:index], items[index+1:]...)
}

func (gs *GameState) HasAdjacentOwnedTile(playerId PlayerId, x, y int) bool {
	for _, n := range HexNeighbors(x, y) {
		tile := gs.TileAt(n[0], n[1])
		if tile == nil {
			continue
		}

		if tile.HasOwner && tile.Owner == playerId {
			return true
		}
	}

	return false
}

func (gs *GameState) UseFloodToken(playerId PlayerId, x, y int) error {
	if gs.CurrentPlayer != playerId {
		return errors.New("not current player")
	}

	if gs.CurrentPhase != PhasePlace && gs.CurrentPhase != PhaseBuild {
		return errors.New("flood tokens can only be used during your place or build phase")
	}

	player, err := gs.playerById(playerId)
	if err != nil {
		return err
	}

	if player.FloodTokens == 0 {
		return errors.New("player has no flood tokens")
	}

	tile := gs.TileAt(x, y)
	if tile == nil {
		return errors.New("tile not found")
	}

	if !gs.CanFloodTile(tile) {
		return errors.New("this tile cannot be converted into river")
	}

	tile.Biome = River
	tile.HasOwner = false
	tile.Owner = 0
	tile.Influence = make(map[PlayerId]uint)
	tile.TempInfluence = make(map[PlayerId]uint)
	tile.HasBlockade = false
	tile.BlockadeOwner = 0

	player.FloodTokens--

	return nil
}

func (gs *GameState) CanFloodTile(tile *Tile) bool {
	if tile == nil {
		return false
	}

	if tile.Biome == River {
		return false
	}

	if tile.Structure != NoneStructure {
		return false
	}

	if tile.HasBlockade {
		return false
	}

	return true
}
