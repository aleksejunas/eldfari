package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

// bestTimes holds the best time per level (keyed by level name, in seconds)
// and saves them as JSON, e.g. ~/.config/raketti/best.json on Linux.
type bestTimes struct {
	path  string // "" = keep in memory only (no config dir, e.g. in a browser)
	times map[string]float64
}

func loadBestTimes() *bestTimes {
	dir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("best times will not be saved: %v", err)
		return &bestTimes{times: map[string]float64{}}
	}
	path := filepath.Join(dir, "raketti", "best.json")
	times, err := readBestTimes(path)
	if err != nil {
		log.Printf("ignoring best times: %v", err)
	}
	return &bestTimes{path: path, times: times}
}

// get returns the best time for the level, or 0 if there is none.
func (b *bestTimes) get(level string) float64 { return b.times[level] }

// record stores t if it beats the old time, and saves the file.
func (b *bestTimes) record(level string, t float64) {
	if old, ok := b.times[level]; ok && old <= t {
		return
	}
	b.times[level] = t
	if b.path == "" {
		return
	}
	if err := writeBestTimes(b.path, b.times); err != nil {
		log.Printf("saving best times: %v", err)
	}
}

// readBestTimes always returns a usable map. A missing file is not an error.
func readBestTimes(path string) (map[string]float64, error) {
	times := map[string]float64{}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return times, nil
	}
	if err != nil {
		return times, err
	}
	if err := json.Unmarshal(data, &times); err != nil {
		return map[string]float64{}, err
	}
	return times, nil
}

// writeBestTimes writes to a temp file and renames it, so a crash mid-write
// can't leave a half-written file behind.
func writeBestTimes(path string, times map[string]float64) error {
	data, err := json.MarshalIndent(times, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
