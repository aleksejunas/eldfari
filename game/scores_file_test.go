//go:build !js

package game

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScoresSaveAndLoadAgain(t *testing.T) {
	path := filepath.Join(t.TempDir(), "eldfari", "scores.json") // dir doesn't exist yet
	s := newScores()
	s.save = func(data []byte) error { return writeFile(path, data) }
	s.add("crystals/Hellir", "ROL", 40)
	s.add("crystals/Hellir", "KAI", 35)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	again, err := decodeScores(data)
	if err != nil {
		t.Fatal(err)
	}
	if top := again.Top["crystals/Hellir"]; len(top) != 2 || top[0].Name != "KAI" || again.Last != "KAI" {
		t.Errorf("loaded %+v, last %q", top, again.Last)
	}
}

func TestBadScoresFileGivesEmptyScores(t *testing.T) {
	s, err := decodeScores([]byte("{not json"))
	if err == nil || s == nil || len(s.Top) != 0 || s.Top == nil {
		t.Errorf("scores=%+v err=%v, want empty, usable scores and an error", s, err)
	}
	old, _ := decodeScores([]byte(`{"crystals/Hellir": 92.4}`)) // the old best-time file
	if old.best("crystals/Hellir") != 0 {
		t.Error("the old best-time format should just give empty scores")
	}
}
