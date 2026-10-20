package game

import (
	"log"
	"strings"
)

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

// renamedLevels maps old best-time keys to the level's current name, so
// records survive a rename (the levels got Old Norse names).
var renamedLevels = map[string]string{
	"crystals/Grotta":    "crystals/Hellir",
	"crystals/Passasjen": "crystals/Gangr",
	"crystals/Sprekken":  "crystals/Rifa",
	"crystals/Dryppet":   "crystals/Dropi",
	"crystals/Langveien": "crystals/Langvegr",
	"crystals/Sjakta":    "crystals/Gjá",
	"crystals/Dypet":     "crystals/Djúp",
	"cargo/Kranen":       "cargo/Vindáss",
	"cargo/Lageret":      "cargo/Búr",
	"cargo/Hylla":        "cargo/Hilla",
}

// migrateOldKeys keeps old records: times saved before there were modes
// ("Grotta") move to the crystals mode ("crystals/Grotta"), and renamed
// levels move to their new name ("crystals/Hellir"). If both exist, the
// newer key wins.
func migrateOldKeys(times map[string]float64) {
	for k, v := range times {
		if strings.Contains(k, "/") {
			continue
		}
		delete(times, k)
		if _, ok := times["crystals/"+k]; !ok {
			times["crystals/"+k] = v
		}
	}
	for old, now := range renamedLevels {
		v, ok := times[old]
		if !ok {
			continue
		}
		delete(times, old)
		if _, ok := times[now]; !ok {
			times[now] = v
		}
	}
}
