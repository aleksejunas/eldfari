package game

import (
	"runtime/debug"
	"time"
)

// version describes the git commit the binary was built from, e.g.
// "d4cc02d 2026-10-19 20:46 UTC". Go stamps this into the binary by itself when
// it is built with `go build` in a git checkout. `go run` and tests get "dev".
func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	return formatVersion(info.Settings)
}

func formatVersion(settings []debug.BuildSetting) string {
	var rev, at string
	modified := false
	for _, s := range settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.time":
			at = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}
	if rev == "" {
		return "dev"
	}
	v := rev[:min(7, len(rev))]
	if t, err := time.Parse(time.RFC3339, at); err == nil {
		v += " " + t.UTC().Format("2006-01-02 15:04 UTC")
	}
	if modified {
		v += " +changes"
	}
	return v
}
