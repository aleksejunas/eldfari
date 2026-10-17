package sim

import (
	"math"
	"testing"
)

// Gulv fra y=500 og ned, plattform x=100..300 ved y=500. Verdenen er 400x600.
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
		{-1, 5, true}, {5, -1, true}, {40, 5, true}, // utenfor verden = fast
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
	// 120*0.5 = 60 uten drag, litt mindre med drag
	if vy := w.Ship.VY; vy < 50 || vy > 60.5 {
		t.Errorf("VY etter 0.5 s fritt fall = %.2f, forventet ca 58", vy)
	}
}

func TestThrustLiftsShip(t *testing.T) {
	w := testWorld()
	fly(w, 200, 300, 0, 0, 0)
	run(w, 60, Input{Thrust: true})
	if w.Ship.VY >= 0 {
		t.Errorf("VY = %.2f, skipet burde stige (negativ VY)", w.Ship.VY)
	}
}

func TestFuelBurn(t *testing.T) {
	w := testWorld()
	fly(w, 200, 300, 0, 0, 0)
	run(w, 60, Input{Thrust: true}) // 1 s
	if f := w.Ship.Fuel; math.Abs(f-(1-BurnRate)) > 0.001 {
		t.Errorf("Fuel = %.4f, forventet %.4f", f, 1-BurnRate)
	}
}

func TestNoFuelNoThrust(t *testing.T) {
	w := testWorld()
	fly(w, 200, 100, 0, 0, 0)
	w.Ship.Fuel = 0
	run(w, 30, Input{Thrust: true})
	if w.Ship.Thrusting {
		t.Error("Thrusting satt uten drivstoff")
	}
	if vy := w.Ship.VY; vy < 50 {
		t.Errorf("VY = %.2f, skal være fritt fall uten drivstoff", vy)
	}
}

func TestLessFuelMeansMoreAcceleration(t *testing.T) {
	heavy, light := testWorld(), testWorld()
	fly(heavy, 200, 300, 0, 0, 0)
	fly(light, 200, 300, 0, 0, 0)
	light.Ship.Fuel = 0.2
	run(heavy, 10, Input{Thrust: true})
	run(light, 10, Input{Thrust: true})
	if light.Ship.VY >= heavy.Ship.VY { // mer negativ = raskere oppover
		t.Errorf("lett VY=%.2f, tung VY=%.2f: lett skip burde akselerere mer", light.Ship.VY, heavy.Ship.VY)
	}
}

func TestRotationDirection(t *testing.T) {
	w := testWorld()
	fly(w, 200, 300, 0, 0, 0)
	run(w, 30, Input{Right: true, Thrust: true})
	if w.Ship.Angle <= 0 || w.Ship.VX <= 0 {
		t.Errorf("Right: angle=%.2f vx=%.2f, begge burde være positive", w.Ship.Angle, w.Ship.VX)
	}
}

func TestSoftLanding(t *testing.T) {
	w := testWorld()
	fly(w, 200, 500-ShipRear-0.5-3, 0, 10, 0)
	run(w, 120, Input{})
	if w.Ship.State != Landed {
		t.Fatalf("state = %v, forventet Landed", w.Ship.State)
	}
	if want := 500 - ShipRear - 0.5; math.Abs(w.Ship.Y-want) > 1e-9 {
		t.Errorf("Y = %v, forventet %v", w.Ship.Y, want)
	}
}

func TestHardLandingCrashes(t *testing.T) {
	w := testWorld()
	fly(w, 200, 400, 0, 0, 0) // faller ca 90 px, for fort
	run(w, 300, Input{})
	if w.Ship.State != Crashed {
		t.Errorf("state = %v, forventet Crashed", w.Ship.State)
	}
}

func TestMissingPadCrashes(t *testing.T) {
	w := testWorld()
	fly(w, 350, 500-ShipRear-0.5-3, 0, 10, 0) // sakte og rett, men utenfor plattformen
	run(w, 120, Input{})
	if w.Ship.State != Crashed {
		t.Errorf("state = %v, forventet Crashed", w.Ship.State)
	}
}

func TestTiltedLandingCrashes(t *testing.T) {
	w := testWorld()
	fly(w, 200, 500-ShipRear-0.5-6, 0, 10, 0.6)
	run(w, 120, Input{})
	if w.Ship.State != Crashed {
		t.Errorf("state = %v, forventet Crashed", w.Ship.State)
	}
}

func TestRefuelOnPad(t *testing.T) {
	w := testWorld()
	w.Ship.Fuel = 0.5
	run(w, 60, Input{})
	if f := w.Ship.Fuel; math.Abs(f-(0.5+RefuelRate)) > 0.001 {
		t.Errorf("Fuel = %.4f, forventet %.4f", f, 0.5+RefuelRate)
	}
}

func TestLiftoffFromPad(t *testing.T) {
	w := testWorld()
	run(w, 30, Input{Thrust: true})
	if w.Ship.State != Flying || w.Ship.Y >= 500-ShipRear-0.5 {
		t.Errorf("state=%v y=%.2f, skipet burde ha løftet", w.Ship.State, w.Ship.Y)
	}
}

// Testverden med én krystall på (200,400). Plattformen er Pads[0].
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
		t.Errorf("NGot=%d Got=%v, krystallen burde være tatt", w.NGot, w.Got)
	}
}

func TestFarCrystalNotPicked(t *testing.T) {
	w := crystalWorld()
	fly(w, 200, 300, 0, 0, 0) // 100 px unna
	w.Step(Input{})
	if w.NGot != 0 {
		t.Errorf("NGot=%d, krystallen er for langt unna", w.NGot)
	}
}

func TestWinNeedsAllCrystalsAndStartPad(t *testing.T) {
	// Lander uten krystall: ingen seier.
	w := crystalWorld()
	fly(w, 200, 500-ShipRear-0.5-3, 0, 10, 0)
	run(w, 120, Input{})
	if w.Ship.State != Landed || w.Won {
		t.Errorf("state=%v won=%v, burde lande uten å vinne", w.Ship.State, w.Won)
	}

	// Tar krystallen, så lander: seier.
	w = crystalWorld()
	fly(w, 200, 400, 0, 0, 0)
	w.Step(Input{})
	w.Ship.X, w.Ship.Y, w.Ship.VX, w.Ship.VY = 200, 500-ShipRear-0.5-3, 0, 10
	run(w, 120, Input{})
	if !w.Won {
		t.Errorf("state=%v won=%v, burde ha vunnet", w.Ship.State, w.Won)
	}
}

func TestTimerStartsAtLiftoff(t *testing.T) {
	w := crystalWorld()
	run(w, 30, Input{}) // står på plattformen
	if w.Time != 0 || w.Started {
		t.Errorf("Time=%v Started=%v før avgang", w.Time, w.Started)
	}
	run(w, 30, Input{Thrust: true})
	if !w.Started || w.Time < 0.4 {
		t.Errorf("Time=%v Started=%v etter 0,5 s flyging", w.Time, w.Started)
	}
}

func TestResetClearsProgress(t *testing.T) {
	w := crystalWorld()
	fly(w, 200, 400, 0, 0, 0)
	w.Started = true
	run(w, 5, Input{})
	w.Reset()
	if w.NGot != 0 || w.Got[0] || w.Time != 0 || w.Started || w.Won {
		t.Errorf("Reset nullstilte ikke: NGot=%d Got=%v Time=%v Started=%v Won=%v",
			w.NGot, w.Got, w.Time, w.Started, w.Won)
	}
}
