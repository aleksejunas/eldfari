package game

import "time"

// fpsMeter tracks the slowest frame during the last second, for the FPS
// overlay (key I). A single slow frame is what you feel as a hitch, and the
// average FPS hides it.
type fpsMeter struct {
	last  time.Time     // when the previous frame was drawn
	since time.Time     // start of the current one-second window
	worst time.Duration // slowest frame so far in this window
	shown time.Duration // slowest frame in the previous window: what the overlay shows
}

// frame is called once per drawn frame.
func (m *fpsMeter) frame(now time.Time) {
	if !m.last.IsZero() {
		m.worst = max(m.worst, now.Sub(m.last))
	}
	m.last = now
	if m.since.IsZero() {
		m.since = now
	}
	if now.Sub(m.since) >= time.Second {
		m.shown, m.worst, m.since = m.worst, 0, now
	}
}
