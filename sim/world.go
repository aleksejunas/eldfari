package sim

import "math"

// Alle tall er gjetninger som skal justeres ved å spille. Enheter: piksler og sekunder.
const (
	Dt = 1.0 / 60 // fast tidssteg; Ebitengine kjører Update 60 ganger i sekundet

	Gravity  = 120.0 // px/s²
	Thrust   = 360.0 // kraft; akselerasjon = Thrust / masse
	TurnRate = 2.4   // rad/s
	Drag     = 0.15  // lineær luftmotstand, 1/s

	BaseMass = 1.0 // tomt skip
	FuelMass = 0.6 // ekstra masse ved full tank. Last = tregere akselerasjon.

	BurnRate   = 0.03 // tank/s ved full thrust
	RefuelRate = 0.25 // tank/s når du står på en plattform

	// Landingskrav
	MaxLandVY    = 80.0 // px/s
	MaxLandVX    = 50.0 // px/s
	MaxLandAngle = 0.5  // rad fra rett opp

	ShipRear = 9.0 // avstand fra skipets midtpunkt til bakkanten (lokal y)

	CrystalRadius = 18.0 // krystallen er tatt når skipets midtpunkt er nærmere enn dette
)

type Input struct{ Left, Right, Thrust bool }

type State int

const (
	Flying State = iota
	Landed
	Crashed
)

// Ship: Angle 0 = nesa peker rett opp, positiv vinkel = med klokka.
type Ship struct {
	X, Y, VX, VY float64
	Angle        float64
	Fuel         float64 // 0..1
	State        State
	Thrusting    bool // sant hvis motoren gikk i dette tidssteget
}

func (s *Ship) Mass() float64 { return BaseMass + FuelMass*s.Fuel }

// ToWorld gjør et lokalt skipspunkt (nese = (0,-12)) om til verdenskoordinater.
func (s *Ship) ToWorld(lx, ly float64) Pt {
	sin, cos := math.Sincos(s.Angle)
	return Pt{
		X: s.X + lx*cos - ly*sin,
		Y: s.Y + lx*sin + ly*cos,
	}
}

var shipShape = [3]Pt{{0, -12}, {-8, ShipRear}, {8, ShipRear}}

// Points gir de tre hjørnene (nese, venstre bak, høyre bak) i verdenskoordinater.
func (s *Ship) Points() [3]Pt {
	var out [3]Pt
	for i, p := range shipShape {
		out[i] = s.ToWorld(p.X, p.Y)
	}
	return out
}

// contactPoints: hjørner + midtpunkt på hver kant. Det er disse som testes mot terrenget.
func (s *Ship) contactPoints() []Pt {
	v := s.Points()
	mid := func(a, b Pt) Pt { return Pt{(a.X + b.X) / 2, (a.Y + b.Y) / 2} }
	return []Pt{v[0], v[1], v[2], mid(v[0], v[1]), mid(v[1], v[2]), mid(v[2], v[0])}
}

type World struct {
	Level *Level
	Ship  Ship

	// Spillstatus. Got har samme indeks som Level.Crystals.
	Got     []bool
	NGot    int
	Time    float64 // sekunder siden første avgang (teller også mens du fyller på)
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

// Step flytter verden ett fast tidssteg (Dt).
func (w *World) Step(in Input) {
	s := &w.Ship
	s.Thrusting = false

	if w.Won || s.State == Crashed {
		return
	}
	if w.Started {
		w.Time += Dt
	}

	if s.State == Landed {
		if in.Thrust && s.Fuel > 0 {
			s.State = Flying // løfter fra bakken, fortsetter under
			w.Started = true
		} else {
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
		a := Thrust / s.Mass() // masse regnes før drivstoffet brennes
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

	w.pickup() // før kollisjon, så siste krystall teller ved en landing i samme steg
	w.collide()
}

// pickup tar krystaller som skipets midtpunkt er innenfor rekkevidde av.
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
		s.Y = pad.Y - ShipRear - 0.5 // bakkanten rett over plattformen
		s.State = Landed
		// Vinn: alle krystaller tatt og landet på startplattformen (Pads[0]).
		if n := len(w.Level.Crystals); n > 0 && w.NGot == n && pad == w.Level.Pads[0] {
			w.Won = true
		}
		return
	}
	s.VX, s.VY = 0, 0
	s.State = Crashed
}

// padFor: alle treffpunkter må ligge på toppen av samme plattform
// (ikke dypt inni den, ikke utenfor kantene).
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
	a := math.Remainder(s.Angle, 2*math.Pi) // til (-π, π]
	return math.Abs(s.VY) <= MaxLandVY &&
		math.Abs(s.VX) <= MaxLandVX &&
		math.Abs(a) <= MaxLandAngle
}
