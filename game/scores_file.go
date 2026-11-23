//go:build !js

package game

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

// loadScores reads the high scores from a JSON file in the user's config dir,
// e.g. ~/.config/eldfari/scores.json on Linux.
func loadScores() *scores {
	dir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("scores will not be saved: %v", err)
		return newScores()
	}
	path := filepath.Join(dir, "eldfari", "scores.json")
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Printf("ignoring scores: %v", err)
	}
	s, err := decodeScores(data)
	if err != nil {
		log.Printf("ignoring scores in %s: %v", path, err)
	}
	s.save = func(data []byte) error { return writeFile(path, data) }
	return s
}

// writeFile writes to a temp file and renames it, so a crash mid-write can't
// leave a half-written file behind.
func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
