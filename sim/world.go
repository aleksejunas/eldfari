package sim

import "math"

// All numbers are guesses meant to be tuned by playing. Units: pixels and seconds.
const (
	Dt = 1.0 / 60 // fixed time step; Ebitengine runs Update 60 times per second

	Gravity  = 120.0 // px/s²
	Thrust   = 360.0 // force; acceleration = Thrust / mass
	TurnRate = 2.4   // rad/s
	Drag     = 0.15  // linear air resistance, 1/s

	BaseMass = 1.0 // empty ship
	FuelMass = 0.6 // extra mass with a full tank. More load = slower acceleration.

	BurnRate   = 0.03 // tank/s at full thrust
	RefuelRate = 0.25 // tank/s while standing on a pad

	// Landing limits
	MaxLandVY    = 80.0 // px/s
	MaxLandVX    = 50.0 // px/s
	MaxLandAngle = 0.5  // rad from straight up

	ShipRear = 9.0 // distance from the ship's centre to its rear edge (local y)

	CrystalRadius = 18.0 // a crystal is taken when the ship's centre is closer than this
)

type Input struct{ Left, Right, Thrust bool }

type State int

const (
	Flying State = iota
	Landed
	Crashed
)

// Ship: Angle 0 = nose points straight up, positive angle = clockwise.
type Ship struct {
	X, Y, VX, VY float64
	Angle        float64
	Fuel         float64 // 0..1
	State        State
	Thrusting    bool // true if the engine fired this time step
	Refuelling   bool // true if the tank was filled this time step
}

func (s *Ship) Mass() float64 { return BaseMass + FuelMass*s.Fuel }

// ToWorld converts a local ship point (nose = (0,-12)) to world coordinates.
func (s *Ship) ToWorld(lx, ly float64) Pt {
	sin, cos := math.Sincos(s.Angle)
	return Pt{
		X: s.X + lx*cos - ly*sin,
		Y: s.Y + lx*sin + ly*cos,
	}
}

var shipShape = [3]Pt{{0, -12}, {-8, ShipRear}, {8, ShipRear}}

// Points returns the three corners (nose, rear left, rear right) in world coordinates.
func (s *Ship) Points() [3]Pt {
	var out [3]Pt
	for i, p := range shipShape {
		out[i] = s.ToWorld(p.X, p.Y)
	}
	return out
}

// contactPoints: corners + the midpoint of each edge. These are tested against the terrain.
func (s *Ship) contactPoints() []Pt {
	v := s.Points()
	mid := func(a, b Pt) Pt { return Pt{(a.X + b.X) / 2, (a.Y + b.Y) / 2} }
	return []Pt{v[0], v[1], v[2], mid(v[0], v[1]), mid(v[1], v[2]), mid(v[2], v[0])}
}

type World struct {
	Level *Level
	Ship  Ship

	// Game state. Got uses the same index as Level.Crystals.
	Got     []bool
	NGot    int
	Time    float64 // seconds since first liftoff (keeps counting while refuelling)
	Started bool
	Won     bool
}

func NewWorld(l *Level) *World {
	w := &World{Level: l}
	w.Reset()
	return w
}

func (w *World) Reset() {
	w.Ship = Ship{X: w.Level.Start.X, Y: w.Level.Start.Y, Fuel: 1, State: Landed}
	w.Got = make([]bool, len(w.Level.Crystals))
	w.NGot, w.Time, w.Started, w.Won = 0, 0, false, false
}

// Step advances the world by one fixed time step (Dt).
func (w *World) Step(in Input) {
	s := &w.Ship
	s.Thrusting = false
	s.Refuelling = false

	if w.Won || s.State == Crashed {
		return
	}
	if w.Started {
		w.Time += Dt
	}

	if s.State == Landed {
		if in.Thrust && s.Fuel > 0 {
			s.State = Flying // lifts off, continues below
			w.Started = true
		} else {
			s.Refuelling = s.Fuel < 1
			s.Fuel = min(1, s.Fuel+RefuelRate*Dt)
			return
		}
	}

	if in.Left {
		s.Angle -= TurnRate * Dt
	}
	if in.Right {
		s.Angle += TurnRate * Dt
	}

	ax, ay := 0.0, Gravity
	if in.Thrust && s.Fuel > 0 {
		a := Thrust / s.Mass() // mass is computed before the fuel is burned
		sin, cos := math.Sincos(s.Angle)
		ax += sin * a
		ay -= cos * a
		s.Fuel = max(0, s.Fuel-BurnRate*Dt)
		s.Thrusting = true
	}

	s.VX += ax * Dt
	s.VY += ay * Dt
	k := 1 - Drag*Dt
	s.VX *= k
	s.VY *= k
	s.X += s.VX * Dt
	s.Y += s.VY * Dt

	w.pickup() // before collision, so the last crystal counts when landing in the same step
	w.collide()
}

// pickup takes crystals that the ship's centre is within reach of.
func (w *World) pickup() {
	s := &w.Ship
	for i, c := range w.Level.Crystals {
		if w.Got[i] {
			continue
		}
		if math.Hypot(s.X-c.X, s.Y-c.Y) <= CrystalRadius {
			w.Got[i] = true
			w.NGot++
		}
	}
}

// NearestCrystal returns the closest crystal not yet taken, or false if all are taken.
func (w *World) NearestCrystal() (Pt, bool) {
	s := &w.Ship
	var best Pt
	bestD, found := math.Inf(1), false
	for i, c := range w.Level.Crystals {
		if w.Got[i] {
			continue
		}
		if d := math.Hypot(c.X-s.X, c.Y-s.Y); d < bestD {
			best, bestD, found = c, d, true
		}
	}
	return best, found
}

// GoalPad returns the pad to head for: the start pad once all crystals are taken,
// otherwise the nearest pad (for refuelling).
func (w *World) GoalPad() Pad {
	if w.NGot == len(w.Level.Crystals) {
		return w.Level.Pads[0]
	}
	s := &w.Ship
	best := w.Level.Pads[0]
	bestD := math.Inf(1)
	for _, p := range w.Level.Pads {
		if d := math.Hypot((p.X0+p.X1)/2-s.X, p.Y-s.Y); d < bestD {
			best, bestD = p, d
		}
	}
	return best
}

func (w *World) collide() {
	s := &w.Ship
	var hit []Pt
	for _, p := range s.contactPoints() {
		if w.Level.Solid(p.X, p.Y) {
			hit = append(hit, p)
		}
	}
	if len(hit) == 0 {
		return
	}
	if pad, ok := w.padFor(hit); ok && s.gentle() {
		s.VX, s.VY, s.Angle = 0, 0, 0
		s.Y = pad.Y - ShipRear - 0.5 // rear edge just above the pad
		s.State = Landed
		// Win: all crystals taken and landed on the start pad (Pads[0]).
		if n := len(w.Level.Crystals); n > 0 && w.NGot == n && pad == w.Level.Pads[0] {
			w.Won = true
		}
		return
	}
	s.VX, s.VY = 0, 0
	s.State = Crashed
}

// padFor: all hit points must be on top of the same pad
// (not deep inside it, not beyond its edges).
func (w *World) padFor(hit []Pt) (Pad, bool) {
	for _, pad := range w.Level.Pads {
		all := true
		for _, p := range hit {
			if p.X < pad.X0 || p.X > pad.X1 || p.Y-pad.Y > 8 {
				all = false
				break
			}
		}
		if all {
			return pad, true
		}
	}
	return Pad{}, false
}

func (s *Ship) gentle() bool {
	a := math.Remainder(s.Angle, 2*math.Pi) // to (-π, π]
	return math.Abs(s.VY) <= MaxLandVY &&
		math.Abs(s.VX) <= MaxLandVX &&
		math.Abs(a) <= MaxLandAngle
}
