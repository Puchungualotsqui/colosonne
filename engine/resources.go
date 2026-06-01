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
			Stone: 3,
			Grain: 2,
		}, nil

	case Road:
		return map[Resource]uint{
			Wood:  1,
			Grain: 1,
		}, nil

	case Bridge:
		return map[Resource]uint{
			Wood:  2,
			Stone: 1,
		}, nil

	case Watchtower:
		return map[Resource]uint{
			Wood:  1,
			Stone: 2,
		}, nil

	default:
		return nil, errors.New("structure has no cost")
	}
}

func (gs *GameState) CanPay(playerId PlayerId, cost map[Resource]uint) error {
	player, err := gs.playerById(playerId)
	if err != nil {
		return err
	}

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

	for resource, amount := range cost {
		player.Resources[resource] -= amount
	}

	return nil
}

func (gs *GameState) AddResource(playerId PlayerId, resource Resource, amount uint) error {
	if resource == NoneResource {
		return nil
	}

	player, err := gs.playerById(playerId)
	if err != nil {
		return err
	}

	if player.Resources == nil {
		player.Resources = make(map[Resource]uint)
	}

	player.Resources[resource] += amount
	return nil
}
