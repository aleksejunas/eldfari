//go:build !js

package game

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

// loadBestTimes reads the best times from a JSON file in the user's config
// dir, e.g. ~/.config/raketti/best.json on Linux.
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
	save := func(times map[string]float64) error { return writeBestTimes(path, times) }
	return &bestTimes{times: times, save: save}
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
