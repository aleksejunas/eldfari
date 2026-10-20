package sim

import (
	"math"
	"testing"

	"raketti/levels"
)

// autopilot flies the ship through waypoints with a simple controller, using
// only the same digital input a player has (left, right, thrust). If it can
// deliver the cargo, the level is playable, and the test catches physics or
// level changes that make it impossible.
type autopilot struct {
	w         *World
	waypoints []Pt
	next      int
	launch    int // frames of full, straight thrust left after leaving a pad
}

func (a *autopilot) input(frame int) Input {
	s := &a.w.Ship
	// Leave a pad like a player does: full thrust straight up for half a
	// second, so the pulses below don't set the ship down again.
	if s.State == Landed {
		a.launch = 30
	}
	if a.launch > 0 {
		a.launch--
		return Input{Thrust: true}
	}
	if a.next < len(a.waypoints)-1 {
		wp := a.waypoints[a.next]
		if math.Hypot(wp.X-s.X, wp.Y-s.Y) < 40 && math.Hypot(s.VX, s.VY) < 60 {
			a.next++
		}
	}
	wp := a.waypoints[a.next]

	// Wanted acceleration towards the waypoint, then the thrust direction
	// that gives it against gravity (y points down).
	ax := math.Max(-60, math.Min(60, 0.6*(wp.X-s.X)-1.2*s.VX))
	ay := math.Max(-60, math.Min(60, 0.6*(wp.Y-s.Y)-1.2*s.VY))
	tx, ty := ax, ay-Gravity
	want := math.Max(-0.7, math.Min(0.7, math.Atan2(tx, -ty)))

	in := Input{Left: s.Angle > want+0.05, Right: s.Angle < want-0.05}
	mass := s.Mass()
	if a.w.Cargo.Attached {
		mass += CargoMass
	}
	// Thrust in pulses, as much of the time as the wanted force needs.
	duty := math.Hypot(tx, ty) / (Thrust / mass)
	in.Thrust = math.Abs(s.Angle-want) < 0.5 && float64(frame%10) < duty*10
	return in
}

func TestCargoLevelsAreDeliverable(t *testing.T) {
	// Waypoints for the ship: to the cargo (it hooks on), then a safe way to
	// the goal pad, ending with the cargo hanging just above it.
	routes := map[string][]Pt{
		"Vindáss": {{800, 820}, {800, 450}, {2000, 450}, {2000, 868}},
		"Búr":     {{610, 600}, {610, 1000}, {610, 1230}, {900, 1290}, {610, 1250}, {610, 500}, {1500, 480}, {2550, 480}, {2550, 818}},
		"Hilla":   {{600, 1170}, {600, 380}, {1850, 380}, {1850, 418}},
	}
	ls, err := LoadDir(levels.Cargo, ModeCargo)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range ls {
		route, ok := routes[l.Name]
		if !ok {
			t.Errorf("%s: no autopilot route, add one", l.Name)
			continue
		}
		t.Run(l.Name, func(t *testing.T) {
			w := NewWorld(l)
			ap := &autopilot{w: w, waypoints: route}
			for frame := range 60 * 120 { // two minutes
				w.Step(ap.input(frame))
				if w.Ship.State == Crashed {
					t.Fatalf("crashed after %.1f s at (%.0f,%.0f), heading for waypoint %d",
						w.Time, w.Ship.X, w.Ship.Y, ap.next)
				}
				if w.Won {
					t.Logf("delivered in %.1f s", w.Time)
					return
				}
			}
			t.Errorf("not delivered after 2 min: waypoint %d, ship (%.0f,%.0f), cargo attached %v at (%.0f,%.0f), fuel %.2f",
				ap.next, w.Ship.X, w.Ship.Y, w.Cargo.Attached, w.Cargo.X, w.Cargo.Y, w.Ship.Fuel)
		})
	}
}
