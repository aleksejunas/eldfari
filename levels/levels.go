// Package levels embeds the level files in the binary, one folder per game
// mode. Drop a new level in its mode's folder and it is picked up
// automatically. File names start with a number with gaps (10-hellir.json,
// 20-gangr.json, ...), and levels are sorted by that number.
//
// A mode's folder is added to the embed pattern when the mode is built,
// because embedding fails on a folder without files.
package levels

import (
	"embed"
	"io/fs"
)

//go:embed crystals/*.json cargo/*.json
var files embed.FS

// Crystals and Cargo hold the levels for each game mode.
var (
	Crystals = sub("crystals")
	Cargo    = sub("cargo")
)

func sub(dir string) fs.FS {
	f, err := fs.Sub(files, dir)
	if err != nil {
		panic(err) // only happens if dir is not a valid path
	}
	return f
}
