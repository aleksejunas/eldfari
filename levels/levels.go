// Package levels embeds the level files in the binary, one folder per game
// mode. Drop a new level in its mode's folder and it is picked up
// automatically. File names start with a number with gaps (10-grotta.json,
// 20-passasjen.json, ...), and levels are sorted by that number.
//
// A mode's folder is added to the embed pattern when the mode is built,
// because embedding fails on a folder without files.
package levels

import (
	"embed"
	"io/fs"
)

//go:embed crystals/*.json
var files embed.FS

// Crystals holds the levels for the crystals mode.
var Crystals = sub("crystals")

func sub(dir string) fs.FS {
	f, err := fs.Sub(files, dir)
	if err != nil {
		panic(err) // only happens if dir is not a valid path
	}
	return f
}
