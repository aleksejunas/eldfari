package game

import "testing"

func TestPadStickDeadZoneAndHysteresis(t *testing.T) {
	var s stickState
	steps := []struct {
		x    float64
		left bool
	}{
		{-0.1, false},  // drift inside the dead zone
		{-0.25, false}, // between off and on: not yet
		{-0.35, true},  // past stickOn: turns
		{-0.25, true},  // back between: keeps turning, no flicker
		{-0.15, false}, // below stickOff: stops
	}
	for i, st := range steps {
		if in := padInput(padState{stickX: st.x}, &s); in.Left != st.left || in.Right {
			t.Errorf("step %d (x=%v): left=%v right=%v, want left=%v", i, st.x, in.Left, in.Right, st.left)
		}
	}
	if in := padInput(padState{stickX: 0.4}, &s); !in.Right || in.Left {
		t.Errorf("stick right: %+v", in)
	}
}

func TestPadButtons(t *testing.T) {
	var s stickState
	in := padInput(padState{dpadLeft: true, trigger: 0.2}, &s)
	if !in.Left || in.Thrust {
		t.Errorf("d-pad left with RT at 0.2: %+v, want left and no thrust", in)
	}
	in = padInput(padState{trigger: 0.5, a: true, b: true}, &s)
	if !in.Thrust || !in.Fire || !in.Bomb {
		t.Errorf("RT 0.5 + A + B: %+v, want thrust, fire and bomb", in)
	}
}
