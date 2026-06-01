package engine

import (
	"errors"
)

type Resource uint

const (
	NoneResource Resource = iota
	Wood
	Stone
	Grain
)

type Biome uint

const (
	NoneBiome Biome = iota
	Forest
	Mountain
	Plain
	River
)

func (b Biome) Resource() (Resource, error) {
	switch b {
	case Forest:
		return Wood, nil
	case Mountain:
		return Stone, nil
	case Plain:
		return Grain, nil
	case River:
		return NoneResource, nil
	default:
		return NoneResource, errors.New("unknown biome")
	}
}

type Structure uint

const (
	NoneStructure Structure = iota
	Outpost
	City
	Bridge
	Watchtower
	Road
)

type Tile struct {
	X, Y           int
	Biome          Biome
	Influence      map[PlayerId]uint
	TempInfluence  map[PlayerId]uint
	Owner          PlayerId
	HasOwner       bool
	Structure      Structure
	StructureOwner PlayerId
	UpgradeLevel   uint
}

func NewTile(biome Biome) Tile {
	return Tile{
		X:              0,
		Y:              0,
		Biome:          biome,
		Influence:      make(map[PlayerId]uint),
		TempInfluence:  make(map[PlayerId]uint),
		Owner:          0,
		HasOwner:       false,
		Structure:      NoneStructure,
		StructureOwner: 0,
		UpgradeLevel:   0,
	}
}
