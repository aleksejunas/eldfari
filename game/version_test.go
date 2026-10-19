package game

import (
	"runtime/debug"
	"testing"
)

func TestFormatVersion(t *testing.T) {
	rev := debug.BuildSetting{Key: "vcs.revision", Value: "d4cc02d42f0f5ea84fbccdd17fd2b5922f1492fe"}
	at := debug.BuildSetting{Key: "vcs.time", Value: "2026-10-19T20:41:35Z"}
	cases := []struct {
		name     string
		settings []debug.BuildSetting
		want     string
	}{
		{"no git info", nil, "dev"},
		{"clean", []debug.BuildSetting{rev, at, {Key: "vcs.modified", Value: "false"}}, "d4cc02d 2026-10-19 20:41 UTC"},
		{"uncommitted changes", []debug.BuildSetting{rev, at, {Key: "vcs.modified", Value: "true"}}, "d4cc02d 2026-10-19 20:41 UTC +changes"},
	}
	for _, c := range cases {
		if got := formatVersion(c.settings); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}
