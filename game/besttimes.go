package game

import "log"

// bestTimes holds the best time per level in seconds, keyed by mode and level
// name ("crystals/Grotta").
// Where they are stored depends on the platform: a JSON file on desktop
// (besttimes_file.go) and localStorage in the browser (besttimes_js.go).
type bestTimes struct {
	times map[string]float64
	save  func(map[string]float64) error // nil = keep in memory only
}

// get returns the best time for the level, or 0 if there is none.
func (b *bestTimes) get(level string) float64 { return b.times[level] }

// record stores t if it beats the old time, and saves all times.
func (b *bestTimes) record(level string, t float64) {
	if old, ok := b.times[level]; ok && old <= t {
		return
	}
	b.times[level] = t
	if b.save == nil {
		return
	}
	if err := b.save(b.times); err != nil {
		log.Printf("saving best times: %v", err)
	}
}
