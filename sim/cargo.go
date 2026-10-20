package sim

import "math"

// Cargo mode: fly to the cargo, it hooks on to a rope automatically, and you
// deliver it to the goal pad. The rope can go slack but never stretches, so a
// swinging, heavy cargo pulls the ship around.
const (
	RopeLength    = 80.0 // px; the cargo hooks on when the ship is this close
	CargoMass     = 0.8  // compare: an empty ship is BaseMass (1), a full one 1.6
	CargoRadius   = 8.0  // px
	CargoBounce   = 0.3  // share of the speed into a wall that bounces back
	CargoFriction = 0.8  // share of the speed along a wall kept at each touch
	DeliverSpeed  = 40.0 // px/s: the cargo must be this calm on the goal pad
)

// Cargo is the load that hangs in the rope under the ship.
type Cargo struct {
	X, Y, VX, VY float64
	Attached     bool // hooked to the ship; stays hooked until delivered
	Delivered    bool
}

// moveCargo moves the cargo h seconds: hooking on, gravity, terrain, the rope
// and delivery. shipFixed is true while the ship stands on a pad, so the rope
// can't move it.
func (w *World) moveCargo(h float64, shipFixed bool) {
	c, s := &w.Cargo, &w.Ship
	if c.Delivered {
		return
	}
	if !c.Attached {
		// The cargo rests where it is until the ship comes close enough.
		if math.Hypot(c.X-s.X, c.Y-s.Y) <= RopeLength {
			c.Attached = true
		}
		return
	}

	px, py := c.X, c.Y
	c.VY += Gravity * h
	k := 1 - Drag*h
	c.VX *= k
	c.VY *= k
	c.X += c.VX * h
	c.Y += c.VY * h
	touching := w.collideCargo(px, py)
	w.rope(shipFixed)

	if touching && w.onGoalPad() && math.Hypot(c.VX, c.VY) <= DeliverSpeed {
		c.Delivered, c.Attached = true, false
		c.VX, c.VY = 0, 0
		w.Won = true
	}
}

// rope keeps the cargo within RopeLength of the ship. If it is too far away,
// both ends are pulled together (the lighter one moves more), and the speed
// that pulls them apart is removed in the same way. A slack rope does nothing.
func (w *World) rope(shipFixed bool) {
	c, s := &w.Cargo, &w.Ship
	dx, dy := c.X-s.X, c.Y-s.Y
	d := math.Hypot(dx, dy)
	if d <= RopeLength {
		return
	}
	nx, ny := dx/d, dy/d // from the ship towards the cargo

	// Inverse masses: how easily each end moves.
	ws, wc := 1/s.Mass(), 1/CargoMass
	if shipFixed {
		ws = 0
	}
	stretch := (d - RopeLength) / (ws + wc)
	s.X += nx * stretch * ws
	s.Y += ny * stretch * ws
	c.X -= nx * stretch * wc
	c.Y -= ny * stretch * wc

	if apart := (c.VX-s.VX)*nx + (c.VY-s.VY)*ny; apart > 0 {
		j := apart / (ws + wc)
		s.VX += nx * j * ws
		s.VY += ny * j * ws
		c.VX -= nx * j * wc
		c.VY -= ny * j * wc
	}
}

// collideCargo keeps the cargo out of the terrain. If it has moved into it,
// it goes back to where it was (px, py), bounces a little and slides along.
// It takes no damage. It reports whether the cargo touched the terrain.
func (w *World) collideCargo(px, py float64) bool {
	c := &w.Cargo
	nx, ny, hit := w.cargoNormal()
	if !hit {
		return false
	}
	c.X, c.Y = px, py
	if vn := c.VX*nx + c.VY*ny; vn < 0 { // moving into the wall
		tx, ty := c.VX-vn*nx, c.VY-vn*ny
		c.VX = tx*CargoFriction - vn*CargoBounce*nx
		c.VY = ty*CargoFriction - vn*CargoBounce*ny
	}
	return true
}

// cargoNormal checks 8 points on the edge of the cargo. If any is inside the
// terrain, it returns a direction pointing away from those points.
func (w *World) cargoNormal() (nx, ny float64, hit bool) {
	c := &w.Cargo
	for i := range 8 {
		dx, dy := math.Cos(float64(i)*math.Pi/4), math.Sin(float64(i)*math.Pi/4)
		if w.Level.Solid(c.X+dx*CargoRadius, c.Y+dy*CargoRadius) {
			nx, ny, hit = nx-dx, ny-dy, true
		}
	}
	if !hit {
		return 0, 0, false
	}
	if d := math.Hypot(nx, ny); d > 0 {
		return nx / d, ny / d, true
	}
	return 0, -1, true // walls on all sides: push up
}

// onGoalPad reports whether the cargo rests on top of the goal pad.
func (w *World) onGoalPad() bool {
	c, pad := &w.Cargo, w.Level.Pads[w.Level.Goal]
	return c.X >= pad.X0 && c.X <= pad.X1 && math.Abs(c.Y+CargoRadius-pad.Y) <= 2
}
