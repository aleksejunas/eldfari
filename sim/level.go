package sim

import (
	"math"
	"sort"
)

// Pt er et punkt i verdenskoordinater (piksler, y peker nedover).
type Pt struct{ X, Y float64 }

// Pad er en flat landingsplattform: toppen av terrenget mellom X0 og X1 ved Y.
type Pad struct{ X0, X1, Y float64 }

// Level har terrenget som polygoner (tegnes som omriss) og en usynlig
// bitmask som brukes til pikselkollisjon. Alt utenfor verdenen er fast stoff.
type Level struct {
	W, H  int
	Polys [][]Pt
	Pads  []Pad
	Start Pt // hvor skipet spawner (midtpunktet)

	Crystals []Pt // Collected by flying onto them

	Name      string
	polyNames []string // just for errors
	padNames  []string

	mask []bool
}

func NewLevel(w, h int, polys [][]Pt, pads []Pad, start Pt) *Level {
	l := &Level{W: w, H: h, Polys: polys, Pads: pads, Start: start}
	l.rasterize()
	return l
}

// Solid sier om punktet er inni terrenget (eller utenfor verdenen).
func (l *Level) Solid(x, y float64) bool {
	ix, iy := int(math.Floor(x)), int(math.Floor(y))
	if ix < 0 || iy < 0 || ix >= l.W || iy >= l.H {
		return true
	}
	return l.mask[iy*l.W+ix]
}

// rasterize fyller polygonene (even-odd, scanline) inn i masken.
// Pikselsenter (x+0.5, y+0.5) avgjør om en piksel er solid.
func (l *Level) rasterize() {
	l.mask = make([]bool, l.W*l.H)
	for y := 0; y < l.H; y++ {
		sy := float64(y) + 0.5
		var xs []float64
		for _, p := range l.Polys {
			for i := range p {
				a, b := p[i], p[(i+1)%len(p)]
				if (a.Y <= sy) == (b.Y <= sy) {
					continue
				}
				t := (sy - a.Y) / (b.Y - a.Y)
				xs = append(xs, a.X+t*(b.X-a.X))
			}
		}
		sort.Float64s(xs)
		for i := 0; i+1 < len(xs); i += 2 {
			x0 := int(math.Ceil(xs[i] - 0.5))
			x1 := int(math.Ceil(xs[i+1] - 0.5)) // eksklusiv
			for x := max(x0, 0); x < min(x1, l.W); x++ {
				l.mask[y*l.W+x] = true
			}
		}
	}
}
