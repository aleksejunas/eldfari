// Package levels builds the mapfiles (*.json) in the binary
// Put new levels in this folder, and it will automatically (sort on filename)

package levels

import "embed"

//go:embed *.json
var FS embed.FS
