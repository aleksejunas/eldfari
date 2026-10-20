package sim

import (
	"math"
	"testing"
)

// Open cave with a floor at y=500: start pad 20..120, goal pad 600..700, and
// the cargo resting on the floor at x=300.
func cargoWorld() *World {
	floor := []Pt{{0, 500}, {800, 500}, {800, 600}, {0, 600}}
	l := NewLevel(800, 600, [][]Pt{floor}, []Pad{{20, 120, 500}, {600, 700, 500}}, Pt{70, 500 - ShipRear - 0.5})
	l.Mode = ModeCargo
	l.CargoStart = Pt{300, 500 - CargoRadius - 0.5}
	l.Goal = 1
	return NewWorld(l)
}

// hook puts the ship straight above the cargo with the rope just tight.
func hook(w *World, shipY float64) {
	c := &w.Cargo
	fly(w, c.X, shipY, 0, 0, 0)
	c.X, c.Y = w.Ship.X, shipY+RopeLength
	c.VX, c.VY, c.Attached = 0, 0, true
}

func ropeLength(w *World) float64 { return math.Hypot(w.Cargo.X-w.Ship.X, w.Cargo.Y-w.Ship.Y) }

func TestCargoHooksOnWhenClose(t *testing.T) {
	w := cargoWorld()
	fly(w, 300, 500-CargoRadius-RopeLength-20, 0, 0, 0) // 20 px too far
	w.Step(Input{Thrust: true})
	if w.Cargo.Attached {
		t.Fatal("the cargo hooked on from too far away")
	}
	fly(w, 300, 500-CargoRadius-RopeLength+10, 0, 0, 0)
	w.Step(Input{Thrust: true})
	if !w.Cargo.Attached {
		t.Error("the cargo should hook on within RopeLength")
	}
}

// The rope never stretches, whatever the ship does. A big empty world, so the
// test flight doesn't hit anything.
func TestRopeNeverLongerThanL(t *testing.T) {
	floor := []Pt{{0, 3900}, {4000, 3900}, {4000, 4000}, {0, 4000}}
	l := NewLevel(4000, 4000, [][]Pt{floor}, []Pad{{20, 120, 3900}}, Pt{70, 3900 - ShipRear - 0.5})
	l.Mode = ModeCargo
	w := NewWorld(l)
	w.Cargo.X = 2000
	hook(w, 1500)
	for i := range 600 {
		// Tilt left, right and back to upright every 2 s, thrusting half the time.
		k := i % 120
		in := Input{Thrust: i%40 < 20, Left: k < 15 || (k >= 75 && k < 90), Right: k >= 30 && k < 60}
		w.Step(in)
		if w.Ship.State != Flying {
			t.Fatalf("frame %d: ship state %v at (%.0f,%.0f), the test flight should stay in the air",
				i, w.Ship.State, w.Ship.X, w.Ship.Y)
		}
		if d := ropeLength(w); d > RopeLength+1e-6 {
			t.Fatalf("frame %d: rope is %.3f px, longer than %v", i, d, RopeLength)
		}
	}
}

// With the ship held still, a cargo let go with the rope horizontal swings
// down and past the middle to the other side. With an 80 px rope and gravity
// 120 a full swing takes about 5 s, so half of it about 2.6 s.
func TestCargoSwingsLikeAPendulum(t *testing.T) {
	w := cargoWorld()
	hook(w, 200)
	w.Cargo.X, w.Cargo.Y = w.Ship.X+RopeLength, w.Ship.Y // rope horizontal
	crossed := false
	for range 3 * 60 * Substeps { // three seconds
		w.moveCargo(Dt/Substeps, true)
		if w.Cargo.X < w.Ship.X-RopeLength/2 {
			crossed = true
		}
	}
	if !crossed {
		t.Error("the cargo should swing to the other side of the ship")
	}
}

// Thrusting straight up, a ship with cargo accelerates slower than without.
func TestCargoSlowsTheShip(t *testing.T) {
	free, loaded := cargoWorld(), cargoWorld()
	fly(free, 300, 200, 0, 0, 0)
	hook(loaded, 200)
	run(free, 30, Input{Thrust: true})
	run(loaded, 30, Input{Thrust: true})
	if loaded.Ship.VY <= free.Ship.VY { // more negative = faster up
		t.Errorf("loaded VY=%.1f, free VY=%.1f: the cargo should slow the ship", loaded.Ship.VY, free.Ship.VY)
	}
	if loaded.Ship.VY >= 0 {
		t.Errorf("loaded VY=%.1f: a full thrust should still lift the cargo", loaded.Ship.VY)
	}
}

// Dropped onto the floor, the cargo bounces a little and stays out of the
// terrain. The ship is held still and close, so the rope stays slack.
func TestCargoBouncesOffTerrain(t *testing.T) {
	w := cargoWorld()
	fly(w, 400, 430, 0, 0, 0)
	w.Cargo = Cargo{X: 400, Y: 450, VY: 300, Attached: true}
	bounced := false
	for range 30 * Substeps {
		w.moveCargo(Dt/Substeps, true)
		if w.Level.Solid(w.Cargo.X, w.Cargo.Y+CargoRadius) {
			t.Fatalf("the cargo is inside the terrain at (%.1f,%.1f)", w.Cargo.X, w.Cargo.Y)
		}
		bounced = bounced || w.Cargo.VY < -10
	}
	if !bounced {
		t.Error("the cargo should bounce up a little")
	}
	if w.Cargo.Delivered || w.Won {
		t.Error("landing on the floor is not a delivery")
	}
}

func TestDeliverCargoOnGoalPad(t *testing.T) {
	w := cargoWorld()
	pad := w.Level.Pads[1]
	x := (pad.X0 + pad.X1) / 2
	fly(w, x, pad.Y-CargoRadius-RopeLength+10, 0, 0, 0) // rope slack
	w.Cargo = Cargo{X: x, Y: pad.Y - CargoRadius - 3, VY: 20, Attached: true}
	run(w, 30, Input{})
	if !w.Cargo.Delivered || !w.Won {
		t.Errorf("delivered=%v won=%v, the cargo rests on the goal pad", w.Cargo.Delivered, w.Won)
	}
}

func TestTooFastIsNotDelivered(t *testing.T) {
	w := cargoWorld()
	pad := w.Level.Pads[1]
	x := (pad.X0 + pad.X1) / 2
	fly(w, x, pad.Y-CargoRadius-RopeLength+10, 0, 0, 0)
	w.Cargo = Cargo{X: x, Y: pad.Y - CargoRadius - 3, VY: 400, Attached: true}
	w.Step(Input{})
	if w.Cargo.Delivered {
		t.Error("a cargo hitting the pad at 400 px/s should bounce, not count as delivered")
	}
}

// While the ship stands on a pad, the cargo still swings and settles, and the
// rope can't pull the ship off the pad.
func TestCargoSettlesWhileLanded(t *testing.T) {
	w := cargoWorld()
	s := &w.Ship
	w.Cargo = Cargo{X: s.X + RopeLength, Y: s.Y - 20, Attached: true} // up to the side
	sx, sy := s.X, s.Y
	run(w, 120, Input{})
	if s.State != Landed || s.X != sx || s.Y != sy {
		t.Errorf("state=%v at (%.1f,%.1f), the ship should stay on the pad", s.State, s.X, s.Y)
	}
	if d := ropeLength(w); d > RopeLength+1e-6 {
		t.Errorf("rope is %.2f px", d)
	}
	if math.Hypot(w.Cargo.VX, w.Cargo.VY) > 5 {
		t.Errorf("cargo speed %.1f after 2 s, it should have settled", math.Hypot(w.Cargo.VX, w.Cargo.VY))
	}
}

// Crystals levels are not affected by the cargo code.
func TestCrystalsModeHasNoCargo(t *testing.T) {
	w := testWorld()
	fly(w, 200, 300, 0, 0, 0)
	run(w, 10, Input{Thrust: true})
	if w.Cargo.Attached {
		t.Error("crystals mode should never hook a cargo")
	}
}
