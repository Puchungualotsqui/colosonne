package engine

import "math/rand"

type PlayerId uint

type Player struct {
	Id PlayerId

	Hand []DraftItem

	Resources map[Resource]uint

	FloodTokens      uint
	FloodworksBought uint
}

func ShufflePlayers(players []Player, rng *rand.Rand) {
	rng.Shuffle(len(players), func(i, j int) {
		players[i], players[j] = players[j], players[i]
	})
}

func (p *Player) ensureResources() {
	if p.Resources == nil {
		p.Resources = make(map[Resource]uint)
	}

	if _, ok := p.Resources[Wood]; !ok {
		p.Resources[Wood] = 0
	}

	if _, ok := p.Resources[Stone]; !ok {
		p.Resources[Stone] = 0
	}

	if _, ok := p.Resources[Grain]; !ok {
		p.Resources[Grain] = 0
	}

	if _, ok := p.Resources[Relic]; !ok {
		p.Resources[Relic] = 0
	}
}
