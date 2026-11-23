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
	BulletDamage  = 0.5  // of full health: two hits destroy a ship (four made shooting pointless)
	ShipHitRadius = 14.0 // px from the ship's centre; a bit larger than the ship, so hits land
	RespawnTime   = 2.0  // s from death to spawning at the base
	RepairRate    = 0.5  // health per second while landed on the own base

	// Bombs drop from the ship with its velocity and fall with gravity, so
	// you can attack from above without aiming the nose. They explode on the
	// terrain or near the other ship and hurt every ship but your own within
	// BlastRadius.
	BombReload  = 1.0  // s between bombs
	BombDamage  = 0.5  // of full health: two bombs destroy a ship
	BlastRadius = 45.0 // px
	BlastTime   = 0.4  // s a blast stays visible
)

type Duel struct {
	Level   *Level
	Players [2]Player
	Bullets []Bullet
	Bombs   []Bomb
	Blasts  []Blast // recent explosions, for drawing
	Winner  int     // -1 until a player reaches DuelScore
}

type Player struct {
	*World          // the player's ship in the shared level
	Base    int     // index in Level.Pads
	Score   int     // points: deaths of the other player
	Health  float64 // 0..1
	Respawn float64 // seconds until the ship spawns again; 0 while alive
	reload  float64
	bombIn  float64 // seconds until the next bomb

	safe    Pt   // race: the last place the ship flew with room around it
	hasSafe bool // false until the ship has flown somewhere safe
	waiting bool // race: hangs still after starting again, until any input
}

type Bomb struct {
	X, Y, VX, VY float64
	Owner        int
}

// Blast is an explosion that has just happened, for drawing.
type Blast struct {
	X, Y float64
	Age  float64 // seconds since it went off
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
	d.Bullets, d.Bombs, d.Blasts = nil, nil, nil
	d.Winner = -1
}

func (p *Player) spawn() {
	pad := p.Level.Pads[p.Base]
	p.Reset()
	p.Ship.X, p.Ship.Y = (pad.X0+pad.X1)/2, pad.Y-ShipRear-0.5
	p.Health, p.Respawn, p.reload, p.bombIn = 1, 0, 0, 0
	p.hasSafe, p.waiting = false, false
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
			// A wreck doesn't thrust or refuel: without this the flags from
			// the crash step stayed on while it waits.
			p.Ship.Thrusting, p.Ship.Refuelling = false, false
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
		p.bombIn -= Dt
		if in[i].Bomb && p.bombIn <= 0 && p.Ship.State == Flying {
			d.dropBomb(i)
		}
		if pad := p.Level.Pads[p.Base]; p.Ship.State == Landed && p.Ship.X >= pad.X0 && p.Ship.X <= pad.X1 {
			p.Health = min(1, p.Health+RepairRate*Dt)
		}
	}
	d.moveBullets()
	d.moveBombs()
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
			d.damage(1-b.Owner, BulletDamage)
			return false
		}
	}
	b.Life -= Dt
	return b.Life > 0
}

// damage takes health from player i and destroys the ship at zero.
func (d *Duel) damage(i int, amount float64) {
	p := &d.Players[i]
	if !p.Alive() {
		return
	}
	if p.Health -= amount; p.Health <= 1e-9 {
		p.Ship.State = Crashed
		d.died(i)
	}
}

func (d *Duel) dropBomb(i int) {
	s := &d.Players[i].Ship
	rear := s.ToWorld(0, ShipRear+4)
	d.Bombs = append(d.Bombs, Bomb{X: rear.X, Y: rear.Y, VX: s.VX, VY: s.VY, Owner: i})
	d.Players[i].bombIn = BombReload
}

// moveBombs lets the bombs fall in substeps and sets them off on the terrain
// or near the other ship. Old blasts fade out.
func (d *Duel) moveBombs() {
	blasts := d.Blasts[:0]
	for _, b := range d.Blasts {
		if b.Age += Dt; b.Age < BlastTime {
			blasts = append(blasts, b)
		}
	}
	d.Blasts = blasts

	falling := d.Bombs[:0]
	for _, b := range d.Bombs {
		if d.moveBomb(&b) {
			falling = append(falling, b)
		}
	}
	d.Bombs = falling
}

// moveBomb reports whether the bomb is still falling.
func (d *Duel) moveBomb(b *Bomb) bool {
	h := Dt / Substeps
	other := &d.Players[1-b.Owner]
	for range Substeps {
		b.VY += Gravity * h
		b.X += b.VX * h
		b.Y += b.VY * h
		near := other.Alive() && math.Hypot(b.X-other.Ship.X, b.Y-other.Ship.Y) <= ShipHitRadius
		if near || d.Level.Solid(b.X, b.Y) {
			d.explode(b)
			return false
		}
	}
	return b.Y < float64(d.Level.H) // gone below the world
}

// explode damages every ship but the bomb owner's within BlastRadius.
func (d *Duel) explode(b *Bomb) {
	d.Blasts = append(d.Blasts, Blast{X: b.X, Y: b.Y})
	other := &d.Players[1-b.Owner]
	if math.Hypot(b.X-other.Ship.X, b.Y-other.Ship.Y) <= BlastRadius {
		d.damage(1-b.Owner, BombDamage)
	}
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
