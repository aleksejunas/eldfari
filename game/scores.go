package game

import (
	"encoding/json"
	"log"
	"slices"
)

// topSize is how many times each level's high-score list keeps.
const topSize = 10

// score is one entry in a high-score list: three initials and a time in
// seconds.
type score struct {
	Name string  `json:"name"`
	Time float64 `json:"time"`
}

// scores holds a top 10 per level, keyed by mode and level name
// ("crystals/Hellir"), fastest first, and the initials last used. Where they
// are stored depends on the platform: a JSON file on desktop (scores_file.go)
// and localStorage in the browser (scores_js.go).
type scores struct {
	Top  map[string][]score `json:"top"`
	Last string             `json:"last"` // initials offered for the next entry

	save func([]byte) error // nil = keep in memory only
}

func newScores() *scores { return &scores{Top: map[string][]score{}, Last: "AAA"} }

// decodeScores reads saved scores. Bad data gives empty scores and an error.
func decodeScores(data []byte) (*scores, error) {
	s := newScores()
	if len(data) == 0 {
		return s, nil
	}
	if err := json.Unmarshal(data, s); err != nil || s.Top == nil {
		return newScores(), err
	}
	return s, nil
}

// best returns the fastest time for the level, or 0 if there is none.
func (s *scores) best(level string) float64 {
	if top := s.Top[level]; len(top) > 0 {
		return top[0].Time
	}
	return 0
}

// qualifies reports whether t makes the level's top 10.
func (s *scores) qualifies(level string, t float64) bool {
	top := s.Top[level]
	return len(top) < topSize || t < top[len(top)-1].Time
}

// add puts a time in the level's top 10, remembers the initials and saves.
// It returns the place in the list (0 = first), or -1 if it didn't make it.
func (s *scores) add(level, name string, t float64) int {
	if !s.qualifies(level, t) {
		return -1
	}
	top := s.Top[level]
	i, _ := slices.BinarySearchFunc(top, t, func(e score, t float64) int {
		if e.Time <= t { // an equal time keeps the older entry first
			return -1
		}
		return 1
	})
	top = slices.Insert(top, i, score{Name: name, Time: t})
	if len(top) > topSize {
		top = top[:topSize]
	}
	s.Top[level], s.Last = top, name
	s.store()
	return i
}

func (s *scores) store() {
	if s.save == nil {
		return
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err == nil {
		err = s.save(append(data, '\n'))
	}
	if err != nil {
		log.Printf("saving scores: %v", err)
	}
}
