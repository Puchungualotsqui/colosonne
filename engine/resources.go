package engine

import "errors"

func StructureCost(structure Structure) (map[Resource]uint, error) {
	switch structure {
	case Outpost:
		return map[Resource]uint{
			Wood:  2,
			Stone: 1,
		}, nil

	case City:
		return map[Resource]uint{
			Stone: 2,
			Grain: 3,
		}, nil

	case Settlement:
		return map[Resource]uint{
			Wood:  2,
			Stone: 2,
			Grain: 2,
		}, nil

	case Bridge:
		return map[Resource]uint{}, nil

	case Watchtower:
		return map[Resource]uint{}, nil

	default:
		return nil, errors.New("structure has no cost")
	}
}

func (gs *GameState) OutpostCost(playerId PlayerId) map[Resource]uint {
	outposts := gs.CountActiveStructures(playerId, Outpost)

	return map[Resource]uint{
		Wood:  2 + outposts,
		Stone: 1,
	}
}

func (gs *GameState) CityCost(playerId PlayerId) map[Resource]uint {
	cities := gs.CountActiveStructures(playerId, City)

	return map[Resource]uint{
		Stone: 2,
		Grain: 3 + cities,
	}
}

func (gs *GameState) SettlementCost(playerId PlayerId) map[Resource]uint {
	settlements := gs.CountActiveStructures(playerId, Settlement)

	return map[Resource]uint{
		Wood:  2,
		Stone: 2,
		Grain: 2 + settlements,
	}
}

func (gs *GameState) BlockadeCost() map[Resource]uint {
	return map[Resource]uint{
		Wood:  1,
		Grain: 1,
	}
}

func (gs *GameState) FloodworksCost(playerId PlayerId) map[Resource]uint {
	player, err := gs.playerById(playerId)
	if err != nil {
		return map[Resource]uint{Crystal: 999999}
	}

	return map[Resource]uint{
		Crystal: 3 + player.FloodworksBought*2,
	}
}

func (gs *GameState) CountActiveStructures(playerId PlayerId, structure Structure) uint {
	var count uint

	for i := range gs.Map {
		tile := &gs.Map[i]

		if tile.Structure != structure {
			continue
		}

		if tile.StructureOwner != playerId {
			continue
		}

		if !gs.StructureActive(tile) {
			continue
		}

		count++
	}

	return count
}

func (gs *GameState) CanPay(playerId PlayerId, cost map[Resource]uint) error {
	player, err := gs.playerById(playerId)
	if err != nil {
		return err
	}

	player.ensureResources()

	for resource, amount := range cost {
		if player.Resources[resource] < amount {
			return errors.New("not enough resources")
		}
	}

	return nil
}

func (gs *GameState) PayResources(playerId PlayerId, cost map[Resource]uint) error {
	if err := gs.CanPay(playerId, cost); err != nil {
		return err
	}

	player, err := gs.playerById(playerId)
	if err != nil {
		return err
	}

	player.ensureResources()

	for resource, amount := range cost {
		player.Resources[resource] -= amount
	}

	return nil
}

func (gs *GameState) AddResource(playerId PlayerId, resource Resource, amount uint) error {
	if resource == NoneResource || amount == 0 {
		return nil
	}

	player, err := gs.playerById(playerId)
	if err != nil {
		return err
	}

	player.ensureResources()
	player.Resources[resource] += amount

	return nil
}
