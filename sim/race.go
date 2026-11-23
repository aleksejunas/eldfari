package sim

import "math"

// Race mode: two players race for the same crystals. A crystal one player
// takes is gone for both. When all are taken and both have landed, the one
// with the most crystals wins. pads[0] and pads[1] are the players' start
// pads. After a crash a player starts again after RespawnTime, in the air at
// the last safe place it flew through, not on its pad: a crash must not be a
// free trip home (or a free landing at the end). Like Duel, each player has a
// World of their own sharing the Level.

// A place is safe to start again from after a crash when it has
// SafeClearance of free air all around and SafeDrop straight below (the ship
// falls). The restarted ship hangs still until the player gives any input,
// so there is time to get going.
const (
	SafeClearance = 80.0
	SafeDrop      = 120.0
)

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

// respawnAtSafe starts the ship again, standing still in the air at its last
// safe place with the fuel it had, or on its pad if it never flew anywhere
// safe.
func (p *Player) respawnAtSafe() {
	if !p.hasSafe {
		p.spawn()
		return
	}
	fuel, at := p.Ship.Fuel, p.safe
	p.Ship = Ship{X: at.X, Y: at.Y, Fuel: fuel, State: Flying}
	p.Respawn, p.waiting = 0, true
}

// Waiting reports whether the ship hangs still after starting again, until
// the player gives any input.
func (p *Player) Waiting() bool { return p.waiting }

// NearestCrystal returns the crystal nobody has taken that is nearest to
// player i, or false if all are taken.
func (r *Race) NearestCrystal(i int) (Pt, bool) {
	s := &r.Players[i].Ship
	var best Pt
	bestD, found := math.Inf(1), false
	for c, pos := range r.Level.Crystals {
		if d := math.Hypot(pos.X-s.X, pos.Y-s.Y); r.Taken[c] < 0 && d < bestD {
			best, bestD, found = pos, d, true
		}
	}
	return best, found
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
				p.respawnAtSafe()
			}
			continue
		}
		if p.waiting {
			if in[i] == (Input{}) {
				continue // hangs still where it started again
			}
			p.waiting = false
		}
		p.Step(in[i])
		if !p.Alive() {
			p.Respawn = RespawnTime
			continue
		}
		if s := &p.Ship; s.State == Flying && r.Level.hasRoom(s.X, s.Y, SafeClearance) && !r.Level.Solid(s.X, s.Y+SafeDrop) {
			p.safe, p.hasSafe = Pt{X: s.X, Y: s.Y}, true
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
