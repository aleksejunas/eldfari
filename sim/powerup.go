package sim

import (
	"math"
	"math/rand/v2"
)

// Power-ups in a duel: rare and random. At most one is out at a time. It
// appears somewhere with room around it every PowerUpEvery seconds (a random
// time in that range), stays PowerUpLife seconds, and is taken by flying
// through it. Its effect lasts PowerUpTime seconds.
const (
	PowerUpMin    = 20.0 // s, shortest wait for the next power-up
	PowerUpMax    = 35.0 // s, longest wait
	PowerUpLife   = 15.0 // s it stays before it disappears
	PowerUpTime   = 10.0 // s the effect lasts
	PowerUpRadius = 20.0 // px: taken when the ship's centre comes this close
	powerUpRoom   = 60.0 // px of free air it needs around it
)

type PowerUpKind int

const (
	Shield  PowerUpKind = iota // bullets and bombs do no damage (crashes still kill)
	OneShot                    // every hit destroys the other ship
	numPowerUpKinds
)

type PowerUp struct {
	X, Y float64
	Kind PowerUpKind
	Life float64 // seconds left before it disappears
}

// powerUps places and hands out power-ups. Duel and Race each have one; kinds
// says which power-ups the mode uses.
type powerUps struct {
	PowerUp     *PowerUp // the power-up that is out, or nil
	kinds       []PowerUpKind
	rng         *rand.Rand
	nextPowerUp float64 // seconds until the next power-up appears
}

// SetSeed decides where and which power-ups appear. New duels and races use
// a fixed seed, so tests are repeatable; the game sets a new one each time.
func (u *powerUps) SetSeed(seed uint64) {
	u.rng = rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	u.reset()
}

func (u *powerUps) reset() {
	u.PowerUp = nil
	if u.rng != nil {
		u.nextPowerUp = u.wait()
	}
}

func (u *powerUps) wait() float64 {
	return PowerUpMin + u.rng.Float64()*(PowerUpMax-PowerUpMin)
}

// step counts down, places a new power-up, lets it run out, and gives it to
// a ship that flies through it.
func (u *powerUps) step(l *Level, ps *[2]Player) {
	for i := range ps {
		ps[i].ShieldLeft = max(0, ps[i].ShieldLeft-Dt)
		ps[i].OneShotLeft = max(0, ps[i].OneShotLeft-Dt)
	}

	if u.PowerUp == nil {
		if u.nextPowerUp -= Dt; u.nextPowerUp <= 0 {
			u.place(l)
		}
		return
	}
	pu := u.PowerUp
	if pu.Life -= Dt; pu.Life <= 0 {
		u.PowerUp, u.nextPowerUp = nil, u.wait()
		return
	}
	for i := range ps {
		p := &ps[i]
		if p.Alive() && math.Hypot(p.Ship.X-pu.X, p.Ship.Y-pu.Y) <= PowerUpRadius {
			switch pu.Kind {
			case Shield:
				p.ShieldLeft = PowerUpTime
			case OneShot:
				p.OneShotLeft = PowerUpTime
			}
			u.PowerUp, u.nextPowerUp = nil, u.wait()
			return
		}
	}
}

// place tries random places until one has room around it. If none is found
// it tries again after a short while.
func (u *powerUps) place(l *Level) {
	for range 100 {
		x, y := u.rng.Float64()*float64(l.W), u.rng.Float64()*float64(l.H)
		if l.hasRoom(x, y, powerUpRoom) {
			kind := u.kinds[u.rng.IntN(len(u.kinds))]
			u.PowerUp = &PowerUp{X: x, Y: y, Kind: kind, Life: PowerUpLife}
			return
		}
	}
	u.nextPowerUp = 1
}
