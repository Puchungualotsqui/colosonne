package engine

import (
	"errors"
	"math/rand"
)

type DraftKind uint

const (
	DraftTile DraftKind = iota
	DraftUpgrade
	DraftStructure
	DraftAction
)

type Action uint

const (
	NoAction Action = iota
	Harvest
	Reinforce
	Expansion
)

type DraftItem struct {
	Kind      DraftKind
	Biome     Biome
	Structure Structure
	Action    Action
}

func GenerateDraftDeck() []DraftItem {
	deck := make([]DraftItem, 0, 80)

	addTile := func(biome Biome, count int) {
		for range count {
			deck = append(deck, DraftItem{
				Kind:  DraftTile,
				Biome: biome,
			})
		}
	}

	addStructure := func(structure Structure, count int) {
		for range count {
			deck = append(deck, DraftItem{
				Kind:      DraftStructure,
				Structure: structure,
			})
		}
	}

	addAction := func(action Action, count int) {
		for range count {
			deck = append(deck, DraftItem{
				Kind:   DraftAction,
				Action: action,
			})
		}
	}

	addUpgrade := func(count int) {
		for range count {
			deck = append(deck, DraftItem{
				Kind: DraftUpgrade,
			})
		}
	}

	addTile(Forest, 14)
	addTile(Mountain, 12)
	addTile(Plain, 16)
	addTile(River, 8)

	addUpgrade(10)

	addStructure(Bridge, 5)
	addStructure(Watchtower, 6)
	addStructure(Road, 6)

	addAction(Harvest, 4)
	addAction(Reinforce, 4)
	addAction(Expansion, 4)

	return deck
}

func ShuffleDraftItems(items []DraftItem, rng *rand.Rand) {
	rng.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})
}

func (gs *GameState) useDraftAction(playerId PlayerId, action Action, x, y int) error {
	switch action {
	case Harvest:
		return gs.actionHarvest(playerId, x, y)

	case Reinforce:
		return gs.actionReinforce(playerId, x, y)

	case Expansion:
		return gs.actionExpansion(playerId)

	default:
		return errors.New("unknown action")
	}
}

func (gs *GameState) actionHarvest(playerId PlayerId, x, y int) error {
	tile := gs.TileAt(x, y)
	if tile == nil {
		return errors.New("tile not found")
	}

	if !gs.playerControlsTile(playerId, tile) {
		return errors.New("player does not control this tile")
	}

	resource, err := tile.Biome.Resource()
	if err != nil {
		return err
	}

	if resource == NoneResource {
		return errors.New("this biome produces no resource")
	}

	player, err := gs.playerById(playerId)
	if err != nil {
		return err
	}

	player.Resources[resource] += 2
	return nil
}

func (gs *GameState) actionReinforce(playerId PlayerId, x, y int) error {
	tile := gs.TileAt(x, y)
	if tile == nil {
		return errors.New("tile not found")
	}

	if tile.TempInfluence == nil {
		tile.TempInfluence = make(map[PlayerId]uint)
	}

	tile.TempInfluence[playerId] += 2
	return nil
}

func (gs *GameState) actionExpansion(playerId PlayerId) error {
	player, err := gs.playerById(playerId)
	if err != nil {
		return err
	}

	player.Resources[Wood] += 1
	player.Resources[Grain] += 1

	return nil
}

func (gs *GameState) useDraftStructure(playerId PlayerId, structure Structure, x, y int) error {
	tile := gs.TileAt(x, y)
	if tile == nil {
		return errors.New("tile not found")
	}

	if tile.Structure != NoneStructure {
		return errors.New("tile already has a structure")
	}

	switch structure {
	case Bridge:
		if tile.Biome != River {
			return errors.New("bridge can only be built on river tiles")
		}

		// For MVP, Bridge may be placed on an unowned river if adjacent to player territory.
		// This makes it usable despite rivers being unownable/resource-less.
		if !gs.HasAdjacentOwnedTile(playerId, x, y) {
			return errors.New("bridge must be adjacent to a tile you control")
		}

		tile.Structure = Bridge
		tile.StructureOwner = playerId
		return nil

	case Road, Watchtower:
		if tile.Biome == River {
			return errors.New("only bridges can be built on river tiles")
		}

		if !gs.playerControlsTile(playerId, tile) {
			return errors.New("player does not control this tile")
		}

		tile.Structure = structure
		tile.StructureOwner = playerId
		return nil

	case Outpost:
		if tile.Biome == River {
			return errors.New("cannot build outpost on river tiles")
		}

		if tile.HasOwner && tile.Owner != playerId {
			return errors.New("cannot build outpost on enemy controlled tile")
		}

		tile.Structure = Outpost
		tile.StructureOwner = playerId
		return nil

	case City:
		return errors.New("city cannot be placed directly; upgrade an outpost instead")

	default:
		return errors.New("invalid structure")
	}
}

func (gs *GameState) useDraftUpgrade(playerId PlayerId, x, y int) error {
	tile := gs.TileAt(x, y)
	if tile == nil {
		return errors.New("tile not found")
	}

	if tile.Biome == River {
		return errors.New("cannot upgrade on river tiles")
	}

	if !gs.playerControlsTile(playerId, tile) {
		return errors.New("player does not control this tile")
	}

	if tile.UpgradeLevel >= 3 {
		return errors.New("tile is already fully upgraded")
	}

	tile.UpgradeLevel++
	return nil
}
