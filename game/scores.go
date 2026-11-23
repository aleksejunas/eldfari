package game

import (
	"encoding/json"
	"log"
	"slices"
	"strings"
)

// topSize is how many times each level's high-score list keeps.
const topSize = 10

// score is one entry in a high-score list: three initials and a time in
// seconds.
type score struct {
	Name string  `json:"name"`
	Time float64 `json:"time"`
}

// scores is the hall of fame: a top 10 of times per level, keyed by mode and
// level name ("crystals/Hellir"), fastest first; wins per initials in the
// two-player modes, keyed by mode ("duel"); and the initials last used. Where they
// are stored depends on the platform: a JSON file on desktop (scores_file.go)
// and localStorage in the browser (scores_js.go).
type scores struct {
	Top  map[string][]score        `json:"top"`
	Wins map[string]map[string]int `json:"wins"` // mode -> initials -> wins
	Last string                    `json:"last"` // initials offered for the next entry

	save func([]byte) error // nil = keep in memory only
}

func newScores() *scores {
	return &scores{Top: map[string][]score{}, Wins: map[string]map[string]int{}, Last: "AAA"}
}

// decodeScores reads saved scores. Bad data gives empty scores and an error.
func decodeScores(data []byte) (*scores, error) {
	s := newScores()
	if len(data) == 0 {
		return s, nil
	}
	if err := json.Unmarshal(data, s); err != nil || s.Top == nil {
		return newScores(), err
	}
	if s.Wins == nil { // saved before there were wins
		s.Wins = map[string]map[string]int{}
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

// addWin counts a win for the initials in a two-player mode, remembers the
// initials, saves, and returns the number of wins they have now.
func (s *scores) addWin(mode, name string) int {
	if s.Wins[mode] == nil {
		s.Wins[mode] = map[string]int{}
	}
	s.Wins[mode][name]++
	s.Last = name
	s.store()
	return s.Wins[mode][name]
}

// winner is one line in a two-player hall of fame.
type winner struct {
	Name string
	Wins int
}

// topWins returns the ten with the most wins in the mode, most first; equal
// wins in alphabetical order.
func (s *scores) topWins(mode string) []winner {
	var out []winner
	for name, n := range s.Wins[mode] {
		out = append(out, winner{name, n})
	}
	slices.SortFunc(out, func(a, b winner) int {
		if a.Wins != b.Wins {
			return b.Wins - a.Wins
		}
		return strings.Compare(a.Name, b.Name)
	})
	return out[:min(len(out), topSize)]
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
