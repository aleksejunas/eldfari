package game

import (
	"image/color"
	"math"

	"eldfari/sim"
)

// drawCargo draws the goal pad in the cargo's colour with a marker bobbing
// above it, the rope and the cargo. ln draws a world line segment.
func (g *Game) drawCargo(ln func(x0, y0, x1, y1, w float64, c color.RGBA)) {
	w := g.world
	c, s := &w.Cargo, &w.Ship

	goal := w.Level.Pads[w.Level.Goal]
	ln(goal.X0, goal.Y, goal.X1, goal.Y, 4, colGem)
	if !c.Delivered {
		mx, my := (goal.X0+goal.X1)/2, goal.Y-40+4*math.Sin(float64(g.tick)/10)
		ln(mx-8, my-6, mx, my+2, 2, colGem)
		ln(mx, my+2, mx+8, my-6, 2, colGem)
	}

	// Before it is hooked: a dashed ring around the cargo shows how close the
	// ship must come.
	if !c.Attached && !c.Delivered {
		const n = 32
		for i := 0; i < n; i += 2 {
			a0, a1 := float64(i)*2*math.Pi/n, float64(i+1)*2*math.Pi/n
			r := sim.HookRange
			ln(c.X+r*math.Cos(a0), c.Y+r*math.Sin(a0), c.X+r*math.Cos(a1), c.Y+r*math.Sin(a1), 1, colDim)
		}
	}

	// The rope goes from the rear of the ship to the edge of the cargo. When
	// it is slack it sags in the middle.
	if c.Attached && s.State != sim.Crashed {
		top := s.ToWorld(0, sim.ShipRear)
		d := math.Hypot(c.X-s.X, c.Y-s.Y)
		ex, ey := c.X, c.Y
		if dd := math.Hypot(top.X-c.X, top.Y-c.Y); dd > 0 {
			ex, ey = c.X+(top.X-c.X)/dd*sim.CargoRadius, c.Y+(top.Y-c.Y)/dd*sim.CargoRadius
		}
		sag := 0.0
		if l := c.RopeLen(); d < l {
			sag = math.Sqrt(l*l-d*d) / 2
		}
		mx, my := (top.X+ex)/2, (top.Y+ey)/2+sag
		ln(top.X, top.Y, mx, my, 1, colShip)
		ln(mx, my, ex, ey, 1, colShip)
	}

	// The cargo: an octagon with a line across.
	r := sim.CargoRadius
	for i := range 8 {
		a0, a1 := float64(i)*math.Pi/4+math.Pi/8, float64(i+1)*math.Pi/4+math.Pi/8
		ln(c.X+r*math.Cos(a0), c.Y+r*math.Sin(a0), c.X+r*math.Cos(a1), c.Y+r*math.Sin(a1), 2, colGem)
	}
	ln(c.X-r*0.6, c.Y, c.X+r*0.6, c.Y, 1, colGem)
}
