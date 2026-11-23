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

// SetSeed decides where and which power-ups appear. NewDuel uses a fixed
// seed, so tests are repeatable; the game sets a new one for each duel.
func (d *Duel) SetSeed(seed uint64) {
	d.rng = rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	d.nextPowerUp = d.powerUpWait()
}

func (d *Duel) powerUpWait() float64 {
	return PowerUpMin + d.rng.Float64()*(PowerUpMax-PowerUpMin)
}

// stepPowerUps counts down, places a new power-up, lets it run out, and
// gives it to a ship that flies through it.
func (d *Duel) stepPowerUps() {
	for i := range d.Players {
		p := &d.Players[i]
		p.ShieldLeft = max(0, p.ShieldLeft-Dt)
		p.OneShotLeft = max(0, p.OneShotLeft-Dt)
	}

	if d.PowerUp == nil {
		if d.nextPowerUp -= Dt; d.nextPowerUp <= 0 {
			d.placePowerUp()
		}
		return
	}
	pu := d.PowerUp
	if pu.Life -= Dt; pu.Life <= 0 {
		d.PowerUp, d.nextPowerUp = nil, d.powerUpWait()
		return
	}
	for i := range d.Players {
		p := &d.Players[i]
		if p.Alive() && math.Hypot(p.Ship.X-pu.X, p.Ship.Y-pu.Y) <= PowerUpRadius {
			switch pu.Kind {
			case Shield:
				p.ShieldLeft = PowerUpTime
			case OneShot:
				p.OneShotLeft = PowerUpTime
			}
			d.PowerUp, d.nextPowerUp = nil, d.powerUpWait()
			return
		}
	}
}

// placePowerUp tries random places until one has room around it. If none is
// found it tries again after a short while.
func (d *Duel) placePowerUp() {
	l := d.Level
	for range 100 {
		x, y := d.rng.Float64()*float64(l.W), d.rng.Float64()*float64(l.H)
		if l.hasRoom(x, y, powerUpRoom) {
			kind := PowerUpKind(d.rng.IntN(int(numPowerUpKinds)))
			d.PowerUp = &PowerUp{X: x, Y: y, Kind: kind, Life: PowerUpLife}
			return
		}
	}
	d.nextPowerUp = 1
}
