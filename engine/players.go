package engine

import (
	"math/rand"
)

type PlayerId uint

type Player struct {
	Id        PlayerId
	Hand      *DraftItem
	Resources map[Resource]uint
}

func ShufflePlayers(players []Player, rng *rand.Rand) {
	rng.Shuffle(len(players), func(i, j int) {
		players[i], players[j] = players[j], players[i]
	})
}
