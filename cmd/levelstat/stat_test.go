package main

import (
	"math"
	"testing"

	"eldfari/sim"
)

// Two rooms joined by a corridor that is 100 px tall and 400 px long.
func corridorLevel(crystals ...sim.Pt) *sim.Level {
	mountain := pts([][2]float64{{0, 0}, {1200, 0}, {1200, 600}, {0, 600}})
	cave := pts([][2]float64{
		{100, 100}, {400, 100}, {400, 250}, {800, 250}, {800, 100}, {1100, 100},
		{1100, 500}, {800, 500}, {800, 350}, {400, 350}, {400, 500}, {100, 500},
	})
	l := sim.NewLevel(1200, 600, [][]sim.Pt{mountain, cave}, []sim.Pad{{X0: 150, X1: 300, Y: 500}},
		sim.Pt{X: 225, Y: 500 - sim.ShipRear - 0.5})
	l.Crystals = crystals
	return l
}

func pts(xy [][2]float64) []sim.Pt {
	out := make([]sim.Pt, len(xy))
	for i, p := range xy {
		out[i] = sim.Pt{X: p[0], Y: p[1]}
	}
	return out
}

func TestGapIsTheCorridor(t *testing.T) {
	_, st := analyse(corridorLevel(sim.Pt{X: 950, Y: 300}))
	if !st.ok {
		t.Fatal("the crystal should be reachable")
	}
	if math.Abs(st.gap-100) > 2*cell {
		t.Errorf("gap = %.0f, want about 100", st.gap)
	}
	if st.gapAt.X < 400 || st.gapAt.X > 800 {
		t.Errorf("narrowest at %v, want inside the corridor (x 400..800)", st.gapAt)
	}
	// Out to the crystal and back: at least twice the straight distance.
	if min := 2 * math.Hypot(950-225, 300-490); st.route < min {
		t.Errorf("route = %.0f, want at least %.0f", st.route, min)
	}
}

func TestUnreachableCrystal(t *testing.T) {
	// A crystal inside the solid rock between the rooms can't be reached.
	_, st := analyse(corridorLevel(sim.Pt{X: 600, Y: 150}))
	if st.ok {
		t.Error("a crystal inside rock should make the level fail")
	}
}
