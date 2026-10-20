package sim

import "math"

// Cargo mode: fly to within HookRange of the cargo, it hooks on to a rope
// automatically (a winch reels it in to RopeLength), and you deliver it to
// the goal pad. The rope can go slack. When it is tight it acts
// as a stiff, damped spring, so pulling the cargo up is a firm tug rather
// than a jolt, and it never stretches more than MaxStretch.
const (
	RopeLength    = 80.0  // px, once reeled in
	HookRange     = 120.0 // px from the ship's centre to the cargo's: it hooks on
	ReelSpeed     = 60.0  // px/s the winch shortens a longer rope to RopeLength
	RopeStiffness = 40.0  // pull per px of stretch (mass × px/s²); the cargo sags ~1 px
	RopeDamping   = 5.0   // pull per px/s that the ends move apart; stops bouncing
	MaxStretch    = 10.0  // px; beyond this the rope is a hard limit
	SwingDamping  = 0.4   // 1/s; slows the cargo's swing across the rope, so it settles
	CargoMass     = 0.4   // compare: an empty ship is BaseMass (1), a full one 1.6
	CargoRadius   = 8.0   // px
	CargoBounce   = 0.3   // share of the speed into a wall that bounces back
	CargoFriction = 0.8   // share of the speed along a wall kept at each touch
	DeliverSpeed  = 40.0  // px/s: the cargo must be this calm on the goal pad
)

// Cargo is the load that hangs in the rope under the ship.
type Cargo struct {
	X, Y, VX, VY float64
	Attached     bool // hooked to the ship; stays hooked until delivered
	Delivered    bool
	Rope         float64 // current rope length while reeling in; see ropeLength
}

// RopeLen is the rope's current length: from the hook distance down to
// RopeLength as the winch reels it in.
func (c *Cargo) RopeLen() float64 { return max(c.Rope, RopeLength) }

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
		if d := math.Hypot(c.X-s.X, c.Y-s.Y); d <= HookRange {
			c.Attached, c.Rope = true, d
		}
		return
	}
	c.Rope = max(RopeLength, c.Rope-ReelSpeed*h)

	px, py := c.X, c.Y
	c.VY += Gravity * h
	k := 1 - Drag*h
	c.VX *= k
	c.VY *= k
	c.X += c.VX * h
	c.Y += c.VY * h
	touching := w.collideCargo(px, py)
	w.rope(h, shipFixed)

	if touching && w.onGoalPad() && math.Hypot(c.VX, c.VY) <= DeliverSpeed {
		c.Delivered, c.Attached = true, false
		c.VX, c.VY = 0, 0
		w.Won = true
	}
}

// rope pulls the ship and the cargo together when the rope is tight: a
// spring with damping, shared by inverse mass (the lighter end moves more).
// A slack rope does nothing. Beyond RopeLength + MaxStretch it is a hard
// limit. While tight, the cargo's swing across the rope is damped a little.
func (w *World) rope(h float64, shipFixed bool) {
	c, s := &w.Cargo, &w.Ship
	dx, dy := c.X-s.X, c.Y-s.Y
	d, length := math.Hypot(dx, dy), c.RopeLen()
	if d <= length {
		return
	}
	nx, ny := dx/d, dy/d // from the ship towards the cargo

	// Inverse masses: how easily each end moves.
	ws, wc := 1/s.Mass(), 1/CargoMass
	if shipFixed {
		ws = 0
	}

	apart := (c.VX-s.VX)*nx + (c.VY-s.VY)*ny // speed the ends move apart
	if pull := RopeStiffness*(d-length) + RopeDamping*apart; pull > 0 {
		s.VX += nx * pull * ws * h
		s.VY += ny * pull * ws * h
		c.VX -= nx * pull * wc * h
		c.VY -= ny * pull * wc * h
	}

	if over := d - length - MaxStretch; over > 0 {
		share := over / (ws + wc)
		s.X += nx * share * ws
		s.Y += ny * share * ws
		c.X -= nx * share * wc
		c.Y -= ny * share * wc
		if apart := (c.VX-s.VX)*nx + (c.VY-s.VY)*ny; apart > 0 {
			j := apart / (ws + wc)
			s.VX += nx * j * ws
			s.VY += ny * j * ws
			c.VX -= nx * j * wc
			c.VY -= ny * j * wc
		}
	}

	// Swing damping: slow the cargo's speed across the rope, relative to the
	// ship.
	rvx, rvy := c.VX-s.VX, c.VY-s.VY
	along := rvx*nx + rvy*ny
	k := SwingDamping * h
	c.VX -= (rvx - along*nx) * k
	c.VY -= (rvy - along*ny) * k
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
