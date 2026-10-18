// Package levels embeds the level files (*.json) in the binary.
// Drop a new level in this folder and it is picked up automatically
// (levels are sorted by file name, so use an NN- prefix).
package levels

import "embed"

//go:embed *.json
var FS embed.FS
