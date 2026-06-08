package engine

import (
	"errors"
	"math/rand"
)

type DraftKind uint

const (
	DraftTile DraftKind = iota
	DraftStructure
	DraftAction
)

type Action uint

const (
	NoAction Action = iota
	Harvest
	Reinforce
	Expansion
	Raid
)

type DraftItem struct {
	Kind      DraftKind
	Biome     Biome
	Structure Structure
	Action    Action
}

func GenerateDraftDeck() []DraftItem {
	deck := make([]DraftItem, 0, 90)

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

	addTile(Forest, 12)
	addTile(Mountain, 12)
	addTile(Plain, 12)
	addTile(Ruins, 6)
	addTile(River, 8)

	// Draft-only structures.
	addStructure(Bridge, 5)
	addStructure(Watchtower, 6)

	addAction(Harvest, 4)
	addAction(Reinforce, 4)
	addAction(Expansion, 4)
	addAction(Raid, 3)

	return deck
}

func ShuffleDraftItems(items []DraftItem, rng *rand.Rand) {
	rng.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})
}

func (gs *GameState) useDraftAction(
	playerId PlayerId,
	action Action,
	x int,
	y int,
	targetPlayerId PlayerId,
	rng *rand.Rand,
) error {
	switch action {
	case Harvest:
		return gs.actionHarvest(playerId, x, y)

	case Reinforce:
		return gs.actionReinforce(playerId, x, y)

	case Expansion:
		return gs.actionExpansion(playerId)

	case Raid:
		return gs.actionRaid(playerId, targetPlayerId, rng)

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

	if !gs.StructureActive(tile) {
		return errors.New("tile has no active structure")
	}

	if tile.HasBlockade {
		return errors.New("tile is blockaded")
	}

	resource, err := tile.Biome.Resource()
	if err != nil {
		return err
	}

	if resource == NoneResource {
		return errors.New("this biome produces no resource")
	}

	return gs.AddResource(playerId, resource, 2)
}

func (gs *GameState) actionReinforce(playerId PlayerId, x, y int) error {
	tile := gs.TileAt(x, y)
	if tile == nil {
		return errors.New("tile not found")
	}

	if !gs.canReceiveInfluence(tile) {
		return errors.New("this tile cannot receive influence")
	}

	if tile.TempInfluence == nil {
		tile.TempInfluence = make(map[PlayerId]uint)
	}

	tile.TempInfluence[playerId] += 2

	return nil
}

func (gs *GameState) actionExpansion(playerId PlayerId) error {
	if err := gs.AddResource(playerId, Wood, 1); err != nil {
		return err
	}

	return gs.AddResource(playerId, Grain, 1)
}

func (gs *GameState) actionRaid(playerId PlayerId, targetPlayerId PlayerId, rng *rand.Rand) error {
	if targetPlayerId == 0 {
		return errors.New("raid requires a target player")
	}

	if targetPlayerId == playerId {
		return errors.New("cannot raid yourself")
	}

	raider, err := gs.playerById(playerId)
	if err != nil {
		return err
	}

	target, err := gs.playerById(targetPlayerId)
	if err != nil {
		return err
	}

	pool := make([]Resource, 0)

	for _, resource := range []Resource{Wood, Stone, Grain, Relic} {
		for range target.Resources[resource] {
			pool = append(pool, resource)
		}
	}

	if len(pool) == 0 {
		return errors.New("target player has no resources")
	}

	if rng == nil {
		rng = rand.New(rand.NewSource(rand.Int63()))
	}

	rng.Shuffle(len(pool), func(i, j int) {
		pool[i], pool[j] = pool[j], pool[i]
	})

	steals := 3
	if len(pool) < steals {
		steals = len(pool)
	}

	for i := 0; i < steals; i++ {
		resource := pool[i]

		if target.Resources[resource] == 0 {
			continue
		}

		target.Resources[resource]--
		raider.Resources[resource]++
	}

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

		if !gs.hasAdjacentControlledTile(playerId, x, y) {
			return errors.New("bridge must be adjacent to your controlled territory")
		}

		tile.Structure = Bridge
		tile.StructureOwner = playerId

		return nil

	case Watchtower:
		if tile.Biome == River {
			return errors.New("watchtower cannot be built on river tiles")
		}

		if !gs.playerControlsTile(playerId, tile) {
			return errors.New("player does not control this tile")
		}

		tile.Structure = Watchtower
		tile.StructureOwner = playerId

		return nil

	case Outpost, City, Settlement:
		return errors.New("this structure is not draft-only; use manual build")

	default:
		return errors.New("invalid draft structure")
	}
}
