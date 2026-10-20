//go:build !js

package game

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBestTimesMissingFile(t *testing.T) {
	times, err := readBestTimes(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil || len(times) != 0 {
		t.Errorf("times=%v err=%v, expected an empty map and no error", times, err)
	}
}

func TestBestTimesCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "best.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	times, err := readBestTimes(path)
	if err == nil || times == nil || len(times) != 0 {
		t.Errorf("times=%v err=%v, expected an empty map and an error", times, err)
	}
}

func TestBestTimesRecordAndReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "raketti", "best.json") // dir doesn't exist yet
	save := func(times map[string]float64) error { return writeBestTimes(path, times) }
	b := &bestTimes{times: map[string]float64{}, save: save}
	b.record("crystals/Grotta", 40)
	b.record("crystals/Grotta", 50) // slower: ignored
	b.record("crystals/Grotta", 35) // faster: kept
	b.record("crystals/Dypet", 90)

	times, err := readBestTimes(path)
	if err != nil {
		t.Fatal(err)
	}
	if times["crystals/Grotta"] != 35 || times["crystals/Dypet"] != 90 || len(times) != 2 {
		t.Errorf("reloaded %v, expected Grotta=35 Dypet=90", times)
	}
	if b.get("crystals/Passasjen") != 0 {
		t.Error("a level without a time should give 0")
	}
}
