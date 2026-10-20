package game

import (
	"testing"
	"time"
)

func TestFPSMeterShowsSlowestFrameOfLastSecond(t *testing.T) {
	var m fpsMeter
	now := time.Unix(0, 0)
	step := func(d time.Duration) {
		now = now.Add(d)
		m.frame(now)
	}

	m.frame(now)
	for range 30 {
		step(16 * time.Millisecond)
	}
	step(80 * time.Millisecond) // one hitch
	for range 40 {
		step(16 * time.Millisecond)
	}
	// More than a second has passed: the hitch is what's shown.
	if m.shown != 80*time.Millisecond {
		t.Errorf("shown = %v, want 80ms", m.shown)
	}

	for range 70 {
		step(16 * time.Millisecond)
	}
	// The next second had no hitch.
	if m.shown != 16*time.Millisecond {
		t.Errorf("shown = %v, want 16ms after a smooth second", m.shown)
	}
}
