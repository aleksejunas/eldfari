package sim

import (
	"math"
	"testing"
)

// Floor from y=500 down, pad x=100..300 at y=500. The world is 400x600.
func testWorld() *World {
	floor := []Pt{{0, 500}, {400, 500}, {400, 600}, {0, 600}}
	l := NewLevel(400, 600, [][]Pt{floor}, []Pad{{100, 300, 500}}, Pt{200, 500 - ShipRear - 0.5})
	return NewWorld(l)
}

func fly(w *World, x, y, vx, vy, angle float64) {
	w.Ship = Ship{X: x, Y: y, VX: vx, VY: vy, Angle: angle, Fuel: 1, State: Flying}
}

func run(w *World, frames int, in Input) {
	for i := 0; i < frames; i++ {
		w.Step(in)
	}
}

func TestRasterize(t *testing.T) {
	sq := []Pt{{10, 10}, {20, 10}, {20, 20}, {10, 20}}
	l := NewLevel(40, 40, [][]Pt{sq}, nil, Pt{})
	cases := []struct {
		x, y float64
		want bool
	}{
		{15, 15, true}, {5, 5, false}, {25, 15, false},
		{-1, 5, true}, {5, -1, true}, {40, 5, true}, // outside the world = solid
	}
	for _, c := range cases {
		if got := l.Solid(c.x, c.y); got != c.want {
			t.Errorf("Solid(%v,%v) = %v, want %v", c.x, c.y, got, c.want)
		}
	}
}

func TestFreeFall(t *testing.T) {
	w := testWorld()
	fly(w, 200, 100, 0, 0, 0)
	run(w, 30, Input{}) // 0.5 s
	// 120*0.5 = 60 without drag, a bit less with drag
	if vy := w.Ship.VY; vy < 50 || vy > 60.5 {
		t.Errorf("VY after 0.5 s free fall = %.2f, expected about 58", vy)
	}
}

func TestThrustLiftsShip(t *testing.T) {
	w := testWorld()
	fly(w, 200, 300, 0, 0, 0)
	run(w, 60, Input{Thrust: true})
	if w.Ship.VY >= 0 {
		t.Errorf("VY = %.2f, the ship should rise (negative VY)", w.Ship.VY)
	}
}

func TestFuelBurn(t *testing.T) {
	w := testWorld()
	fly(w, 200, 300, 0, 0, 0)
	run(w, 60, Input{Thrust: true}) // 1 s
	if f := w.Ship.Fuel; math.Abs(f-(1-BurnRate)) > 0.001 {
		t.Errorf("Fuel = %.4f, expected %.4f", f, 1-BurnRate)
	}
}

func TestNoFuelNoThrust(t *testing.T) {
	w := testWorld()
	fly(w, 200, 100, 0, 0, 0)
	w.Ship.Fuel = 0
	run(w, 30, Input{Thrust: true})
	if w.Ship.Thrusting {
		t.Error("Thrusting set without fuel")
	}
	if vy := w.Ship.VY; vy < 50 {
		t.Errorf("VY = %.2f, should be free fall without fuel", vy)
	}
}

func TestLessFuelMeansMoreAcceleration(t *testing.T) {
	heavy, light := testWorld(), testWorld()
	fly(heavy, 200, 300, 0, 0, 0)
	fly(light, 200, 300, 0, 0, 0)
	light.Ship.Fuel = 0.2
	run(heavy, 10, Input{Thrust: true})
	run(light, 10, Input{Thrust: true})
	if light.Ship.VY >= heavy.Ship.VY { // more negative = faster upwards
		t.Errorf("light VY=%.2f, heavy VY=%.2f: a light ship should accelerate more", light.Ship.VY, heavy.Ship.VY)
	}
}

func TestRotationDirection(t *testing.T) {
	w := testWorld()
	fly(w, 200, 300, 0, 0, 0)
	run(w, 30, Input{Right: true, Thrust: true})
	if w.Ship.Angle <= 0 || w.Ship.VX <= 0 {
		t.Errorf("Right: angle=%.2f vx=%.2f, both should be positive", w.Ship.Angle, w.Ship.VX)
	}
}

func TestSoftLanding(t *testing.T) {
	w := testWorld()
	fly(w, 200, 500-ShipRear-0.5-3, 0, 10, 0)
	run(w, 120, Input{})
	if w.Ship.State != Landed {
		t.Fatalf("state = %v, expected Landed", w.Ship.State)
	}
	if want := 500 - ShipRear - 0.5; math.Abs(w.Ship.Y-want) > 1e-9 {
		t.Errorf("Y = %v, expected %v", w.Ship.Y, want)
	}
}

func TestHardLandingCrashes(t *testing.T) {
	w := testWorld()
	fly(w, 200, 400, 0, 0, 0) // falls about 90 px, too fast
	run(w, 300, Input{})
	if w.Ship.State != Crashed {
		t.Errorf("state = %v, expected Crashed", w.Ship.State)
	}
}

func TestMissingPadCrashes(t *testing.T) {
	w := testWorld()
	fly(w, 350, 500-ShipRear-0.5-3, 0, 10, 0) // slow and upright, but beside the pad
	run(w, 120, Input{})
	if w.Ship.State != Crashed {
		t.Errorf("state = %v, expected Crashed", w.Ship.State)
	}
}

func TestTiltedLandingCrashes(t *testing.T) {
	w := testWorld()
	fly(w, 200, 500-ShipRear-0.5-6, 0, 10, 0.6)
	run(w, 120, Input{})
	if w.Ship.State != Crashed {
		t.Errorf("state = %v, expected Crashed", w.Ship.State)
	}
}

func TestRefuelOnPad(t *testing.T) {
	w := testWorld()
	w.Ship.Fuel = 0.5
	run(w, 60, Input{})
	if f := w.Ship.Fuel; math.Abs(f-(0.5+RefuelRate)) > 0.001 {
		t.Errorf("Fuel = %.4f, expected %.4f", f, 0.5+RefuelRate)
	}
}

func TestRefuellingFlag(t *testing.T) {
	w := testWorld()
	w.Ship.Fuel = 0.5
	w.Step(Input{})
	if !w.Ship.Refuelling {
		t.Error("should be refuelling while landed with a tank that is not full")
	}
	run(w, 180, Input{}) // 3 s: the tank is full
	if w.Ship.Refuelling {
		t.Error("should stop Refuelling when the tank is full")
	}
	w.Ship.Fuel = 0.5
	w.Step(Input{Thrust: true}) // lifts off
	if w.Ship.Refuelling {
		t.Error("should not be refuelling in the air")
	}
}

func TestLiftoffFromPad(t *testing.T) {
	w := testWorld()
	run(w, 30, Input{Thrust: true})
	if w.Ship.State != Flying || w.Ship.Y >= 500-ShipRear-0.5 {
		t.Errorf("state=%v y=%.2f, the ship should have lifted off", w.Ship.State, w.Ship.Y)
	}
}

// Test world with one crystal at (200,400). The pad is Pads[0].
func crystalWorld() *World {
	w := testWorld()
	w.Level.Crystals = []Pt{{200, 400}}
	w.Reset()
	return w
}

func TestPickupCrystal(t *testing.T) {
	w := crystalWorld()
	fly(w, 200, 400, 0, 0, 0)
	w.Step(Input{})
	if w.NGot != 1 || !w.Got[0] {
		t.Errorf("NGot=%d Got=%v, the crystal should be taken", w.NGot, w.Got)
	}
}

func TestFarCrystalNotPicked(t *testing.T) {
	w := crystalWorld()
	fly(w, 200, 300, 0, 0, 0) // 100 px away
	w.Step(Input{})
	if w.NGot != 0 {
		t.Errorf("NGot=%d, the crystal is too far away", w.NGot)
	}
}

func TestWinNeedsAllCrystalsAndStartPad(t *testing.T) {
	// Lands without the crystal: no win.
	w := crystalWorld()
	fly(w, 200, 500-ShipRear-0.5-3, 0, 10, 0)
	run(w, 120, Input{})
	if w.Ship.State != Landed || w.Won {
		t.Errorf("state=%v won=%v, should land without winning", w.Ship.State, w.Won)
	}

	// Takes the crystal, then lands: win.
	w = crystalWorld()
	fly(w, 200, 400, 0, 0, 0)
	w.Step(Input{})
	w.Ship.X, w.Ship.Y, w.Ship.VX, w.Ship.VY = 200, 500-ShipRear-0.5-3, 0, 10
	run(w, 120, Input{})
	if !w.Won {
		t.Errorf("state=%v won=%v, should have won", w.Ship.State, w.Won)
	}
}

func TestNearestCrystalAndGoalPad(t *testing.T) {
	floor := []Pt{{0, 500}, {800, 500}, {800, 600}, {0, 600}}
	pads := []Pad{{50, 150, 500}, {650, 750, 500}}
	l := NewLevel(800, 600, [][]Pt{floor}, pads, Pt{100, 500 - ShipRear - 0.5})
	l.Crystals = []Pt{{200, 300}, {600, 300}}
	w := NewWorld(l)
	fly(w, 550, 300, 0, 0, 0)

	if c, ok := w.NearestCrystal(); !ok || c != l.Crystals[1] {
		t.Errorf("NearestCrystal = %v %v, expected %v", c, ok, l.Crystals[1])
	}
	if p := w.GoalPad(); p != pads[1] {
		t.Errorf("GoalPad = %v, expected the nearest pad %v", p, pads[1])
	}

	w.Got[1], w.NGot = true, 1
	if c, ok := w.NearestCrystal(); !ok || c != l.Crystals[0] {
		t.Errorf("NearestCrystal = %v %v, expected %v after taking the nearest", c, ok, l.Crystals[0])
	}

	w.Got[0], w.NGot = true, 2
	if _, ok := w.NearestCrystal(); ok {
		t.Error("NearestCrystal should return false when all are taken")
	}
	if p := w.GoalPad(); p != pads[0] {
		t.Errorf("GoalPad = %v, expected the start pad %v when all crystals are taken", p, pads[0])
	}
}

func TestTimerStartsAtLiftoff(t *testing.T) {
	w := crystalWorld()
	run(w, 30, Input{}) // standing on the pad
	if w.Time != 0 || w.Started {
		t.Errorf("Time=%v Started=%v before liftoff", w.Time, w.Started)
	}
	run(w, 30, Input{Thrust: true})
	if !w.Started || w.Time < 0.4 {
		t.Errorf("Time=%v Started=%v after 0.5 s of flight", w.Time, w.Started)
	}
}

func TestResetClearsProgress(t *testing.T) {
	w := crystalWorld()
	fly(w, 200, 400, 0, 0, 0)
	w.Started = true
	run(w, 5, Input{})
	w.Reset()
	if w.NGot != 0 || w.Got[0] || w.Time != 0 || w.Started || w.Won {
		t.Errorf("Reset did not clear: NGot=%d Got=%v Time=%v Started=%v Won=%v",
			w.NGot, w.Got, w.Time, w.Started, w.Won)
	}
}
