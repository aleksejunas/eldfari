package sim

import "math"

// Duel mode: two players in the same level shoot each other. Each player has
// a World of their own for their ship, sharing the Level, so all the ship
// physics is the same as in the other modes. pads[0] and pads[1] are the
// players' bases: they spawn there, and landing there refuels and repairs.
// Dying (shot down or crashed) gives the other player a point.
const (
	DuelScore     = 5    // points to win
	BulletSpeed   = 420  // px/s, added to the ship's own velocity
	BulletLife    = 1.6  // s
	ReloadTime    = 0.25 // s between shots
	BulletDamage  = 0.25 // of full health: four hits destroy a ship
	ShipHitRadius = 11.0 // px from the ship's centre
	RespawnTime   = 2.0  // s from death to spawning at the base
	RepairRate    = 0.5  // health per second while landed on the own base
)

type Duel struct {
	Level   *Level
	Players [2]Player
	Bullets []Bullet
	Winner  int // -1 until a player reaches DuelScore
}

type Player struct {
	*World          // the player's ship in the shared level
	Base    int     // index in Level.Pads
	Score   int     // points: deaths of the other player
	Health  float64 // 0..1
	Respawn float64 // seconds until the ship spawns again; 0 while alive
	reload  float64
}

type Bullet struct {
	X, Y, VX, VY float64
	Life         float64 // seconds left
	Owner        int     // index in Players
}

func NewDuel(l *Level) *Duel {
	d := &Duel{Level: l}
	for i := range d.Players {
		d.Players[i] = Player{World: NewWorld(l), Base: i}
	}
	d.Reset()
	return d
}

// Reset starts the duel again: both players at their bases, no points.
func (d *Duel) Reset() {
	for i := range d.Players {
		p := &d.Players[i]
		p.Score = 0
		p.spawn()
	}
	d.Bullets = nil
	d.Winner = -1
}

func (p *Player) spawn() {
	pad := p.Level.Pads[p.Base]
	p.Reset()
	p.Ship.X, p.Ship.Y = (pad.X0+pad.X1)/2, pad.Y-ShipRear-0.5
	p.Health, p.Respawn, p.reload = 1, 0, 0
}

// Alive reports whether the player's ship is in the game (not waiting to
// respawn).
func (p *Player) Alive() bool { return p.Ship.State != Crashed }

// Step advances the duel by one fixed time step (Dt).
func (d *Duel) Step(in [2]Input) {
	if d.Winner >= 0 {
		return
	}
	for i := range d.Players {
		p := &d.Players[i]
		if !p.Alive() {
			if p.Respawn -= Dt; p.Respawn <= 0 {
				p.spawn()
			}
			continue
		}
		p.Step(in[i])
		if !p.Alive() {
			d.died(i) // crashed into the terrain
			continue
		}
		p.reload -= Dt
		if in[i].Fire && p.reload <= 0 && p.Ship.State == Flying {
			d.fire(i)
		}
		if pad := p.Level.Pads[p.Base]; p.Ship.State == Landed && p.Ship.X >= pad.X0 && p.Ship.X <= pad.X1 {
			p.Health = min(1, p.Health+RepairRate*Dt)
		}
	}
	d.moveBullets()
}

func (d *Duel) fire(i int) {
	s := &d.Players[i].Ship
	sin, cos := math.Sincos(s.Angle)
	nose := s.ToWorld(0, -12)
	d.Bullets = append(d.Bullets, Bullet{
		X: nose.X, Y: nose.Y,
		VX: s.VX + sin*BulletSpeed, VY: s.VY - cos*BulletSpeed,
		Life: BulletLife, Owner: i,
	})
	d.Players[i].reload = ReloadTime
}

// moveBullets moves the bullets in substeps, so they can't skip a thin wall
// or a ship, and removes those that hit something or run out of time.
func (d *Duel) moveBullets() {
	alive := d.Bullets[:0]
	for _, b := range d.Bullets {
		if d.moveBullet(&b) {
			alive = append(alive, b)
		}
	}
	d.Bullets = alive
}

// moveBullet reports whether the bullet is still flying.
func (d *Duel) moveBullet(b *Bullet) bool {
	h := Dt / Substeps
	for range Substeps {
		b.X += b.VX * h
		b.Y += b.VY * h
		if d.Level.Solid(b.X, b.Y) {
			return false
		}
		other := &d.Players[1-b.Owner]
		if other.Alive() && math.Hypot(b.X-other.Ship.X, b.Y-other.Ship.Y) <= ShipHitRadius {
			if other.Health -= BulletDamage; other.Health <= 1e-9 {
				other.Ship.State = Crashed
				d.died(1 - b.Owner)
			}
			return false
		}
	}
	b.Life -= Dt
	return b.Life > 0
}

// died gives the other player a point and starts the respawn timer.
func (d *Duel) died(i int) {
	p := &d.Players[i]
	p.Respawn, p.Health = RespawnTime, 0
	other := &d.Players[1-i]
	if other.Score++; other.Score >= DuelScore {
		d.Winner = 1 - i
	}
}
