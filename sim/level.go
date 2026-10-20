package sim

import (
	"math"
	"sort"
)

// Pt is a point in world coordinates (pixels, y points down).
type Pt struct{ X, Y float64 }

// Pad is a flat landing pad: the top of the terrain between X0 and X1 at Y.
type Pad struct{ X0, X1, Y float64 }

// Mode is the game mode a level is made for. The folder the level file is
// in decides it (levels/crystals, levels/cargo).
type Mode int

const (
	ModeCrystals Mode = iota // collect all crystals, land on the start pad
	ModeCargo                // deliver the cargo to the goal pad
)

// Level holds the terrain as polygons (drawn as outlines) and an invisible
// bitmask used for pixel collision. Everything outside the world is solid.
type Level struct {
	W, H  int
	Polys [][]Pt
	Pads  []Pad
	Start Pt // where the ship spawns (its centre)

	Mode Mode

	Crystals []Pt // crystals mode: collected by flying onto them

	CargoStart Pt  // cargo mode: where the cargo rests at the start (its centre)
	Goal       int // cargo mode: index in Pads of the goal pad

	Name      string
	polyNames []string // only used in error messages
	padNames  []string
	goalName  string // cargo mode: the "goal" from the file, for error messages

	mask []uint64 // one bit per pixel, row by row; see Solid
}

func NewLevel(w, h int, polys [][]Pt, pads []Pad, start Pt) *Level {
	l := &Level{W: w, H: h, Polys: polys, Pads: pads, Start: start}
	l.rasterize()
	return l
}

// Solid reports whether the point is inside the terrain (or outside the world).
func (l *Level) Solid(x, y float64) bool {
	ix, iy := int(math.Floor(x)), int(math.Floor(y))
	if ix < 0 || iy < 0 || ix >= l.W || iy >= l.H {
		return true
	}
	i := iy*l.W + ix
	return l.mask[i/64]&(1<<(i%64)) != 0
}

// rasterize fills the polygons (even-odd, scanline) into the mask.
// The pixel centre (x+0.5, y+0.5) decides whether a pixel is solid.
// The mask has one bit per pixel: a 4000x1600 level is 800 KB instead of
// 6.4 MB, which keeps the heap small and garbage collection fast in WASM.
func (l *Level) rasterize() {
	l.mask = make([]uint64, (l.W*l.H+63)/64)
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
			x1 := int(math.Ceil(xs[i+1] - 0.5)) // exclusive
			for x := max(x0, 0); x < min(x1, l.W); x++ {
				i := y*l.W + x
				l.mask[i/64] |= 1 << (i % 64)
			}
		}
	}
}
