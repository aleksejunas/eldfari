package sim

import "math"

// Race mode: two players race for the same crystals. A crystal one player
// takes is gone for both. When all are taken and both have landed, the one
// with the most crystals wins. pads[0] and pads[1] are the players' start
// pads; after a crash a player starts there again after RespawnTime. Like
// Duel, each player has a World of their own sharing the Level.

// RaceDraw is Race.Winner when both took as many crystals.
const RaceDraw = 2

type Race struct {
	Level   *Level
	Players [2]Player // Score is the number of crystals taken
	Taken   []int     // per crystal: -1, or the player who took it
	Winner  int       // -1 while racing, then 0, 1 or RaceDraw
}

func NewRace(l *Level) *Race {
	r := &Race{Level: l}
	for i := range r.Players {
		r.Players[i] = Player{World: NewWorld(l), Base: i}
	}
	r.Reset()
	return r
}

func (r *Race) Reset() {
	for i := range r.Players {
		r.Players[i].Score = 0
		r.Players[i].spawn()
	}
	r.Taken = make([]int, len(r.Level.Crystals))
	for i := range r.Taken {
		r.Taken[i] = -1
	}
	r.Winner = -1
}

// Left is the number of crystals nobody has taken yet.
func (r *Race) Left() int {
	n := 0
	for _, t := range r.Taken {
		if t < 0 {
			n++
		}
	}
	return n
}

// Step advances the race by one fixed time step (Dt).
func (r *Race) Step(in [2]Input) {
	if r.Winner >= 0 {
		return
	}
	for i := range r.Players {
		p := &r.Players[i]
		if !p.Alive() {
			if p.Respawn -= Dt; p.Respawn <= 0 {
				p.spawn()
			}
			continue
		}
		p.Step(in[i])
		if !p.Alive() {
			p.Respawn = RespawnTime
			continue
		}
		for c, pos := range r.Level.Crystals {
			if r.Taken[c] < 0 && math.Hypot(p.Ship.X-pos.X, p.Ship.Y-pos.Y) <= CrystalRadius {
				r.Taken[c] = i
				p.Score++
			}
		}
	}

	a, b := &r.Players[0], &r.Players[1]
	if r.Left() == 0 && a.Ship.State == Landed && b.Ship.State == Landed {
		switch {
		case a.Score > b.Score:
			r.Winner = 0
		case b.Score > a.Score:
			r.Winner = 1
		default:
			r.Winner = RaceDraw
		}
	}
}
