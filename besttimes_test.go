package main

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
	os.WriteFile(path, []byte("{not json"), 0o644)
	times, err := readBestTimes(path)
	if err == nil || times == nil || len(times) != 0 {
		t.Errorf("times=%v err=%v, expected an empty map and an error", times, err)
	}
}

func TestBestTimesRecordAndReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "raketti", "best.json") // dir doesn't exist yet
	b := &bestTimes{path: path, times: map[string]float64{}}

	b.record("Grotta", 40)
	b.record("Grotta", 50) // slower: ignored
	b.record("Grotta", 35) // faster: kept
	b.record("Dypet", 90)

	times, err := readBestTimes(path)
	if err != nil {
		t.Fatal(err)
	}
	if times["Grotta"] != 35 || times["Dypet"] != 90 || len(times) != 2 {
		t.Errorf("reloaded %v, expected Grotta=35 Dypet=90", times)
	}
	if b.get("Passasjen") != 0 {
		t.Error("a level without a time should give 0")
	}
}
