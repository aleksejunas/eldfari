package sim

import (
	"fmt"
	"strings"
	"testing"

	"raketti/levels"
)

func TestShippedLevelsLoadAndValidate(t *testing.T) {
	ls, err := LoadDir(levels.FS)
	if err != nil {
		t.Fatalf("the levels are not valid:\n%v", err)
	}
	if len(ls) == 0 {
		t.Fatal("no levels")
	}
	for _, l := range ls {
		if l.Name == "" {
			t.Error("level without a name")
		}
	}
}

// Every shipped level: the ship starts at rest on the start pad and refuels.
func TestShippedLevelsSpawnAndRefuel(t *testing.T) {
	ls, err := LoadDir(levels.FS)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range ls {
		w := NewWorld(l)
		if w.Ship.State != Landed {
			t.Errorf("%s: the ship should start Landed", l.Name)
		}
		w.Ship.Fuel = 0.5
		run(w, 180, Input{}) // 3 s: more than enough to fill the tank, so min(1, ...) takes over
		if w.Ship.State != Landed || w.Ship.Fuel != 1 {
			t.Errorf("%s: state=%v fuel=%v after 3 s without input", l.Name, w.Ship.State, w.Ship.Fuel)
		}
	}
}

// A small, valid level: floor at y=500, pad 100..300, one crystal.
// The tests below change one thing at a time and check that the error is found.
func levelJSON(polyExtra, pad, crystals string) string {
	return fmt.Sprintf(`{
  "name": "test",
  "size": [800, 600],
  "polygons": [
    {"name": "floor", "points": [[0,500],[800,500],[800,600],[0,600]]}%s
  ],
  "pads": [%s],
  "crystals": %s
}`, polyExtra, pad, crystals)
}

const (
	okPad      = `{"name": "start", "x0": 100, "x1": 300, "y": 500}`
	okCrystals = `[[400, 300]]`
)

func TestParseValidLevel(t *testing.T) {
	l, err := ParseLevel([]byte(levelJSON("", okPad, okCrystals)))
	if err != nil {
		t.Fatalf("should be valid: %v", err)
	}
	if l.Name != "test" || len(l.Crystals) != 1 || len(l.Pads) != 1 {
		t.Errorf("unexpected content: %+v", l)
	}
}

func TestParseRejectsBadLevels(t *testing.T) {
	cases := []struct {
		name string
		json string
		want string // substring of the error message
	}{
		{
			"crossing polygons",
			levelJSON(`, {"name": "cross", "points": [[500,450],[600,450],[600,550],[500,550]]}`, okPad, okCrystals),
			"crosses",
		},
		{
			"pad too narrow",
			levelJSON("", `{"name": "narrow", "x0": 100, "x1": 140, "y": 500}`, okCrystals),
			"wide",
		},
		{
			"floating pad",
			levelJSON("", `{"name": "floating", "x0": 100, "x1": 300, "y": 480}`, okCrystals),
			"not flat",
		},
		{
			"crystal inside terrain",
			levelJSON("", okPad, `[[400, 550]]`),
			"inside the terrain",
		},
		{
			"crystal close to terrain",
			levelJSON("", okPad, `[[400, 480]]`),
			"within",
		},
		{
			"no crystals",
			levelJSON("", okPad, `[]`),
			"crystals",
		},
		{
			"terrain right above the pad",
			levelJSON(`, {"name": "roof", "points": [[100,450],[300,450],[300,430],[100,430]]}`, okPad, okCrystals),
			"above it",
		},
		{
			"unknown field",
			strings.Replace(levelJSON("", okPad, okCrystals), `"name": "test",`, `"name": "test", "pdas": [],`, 1),
			"pdas",
		},
		{
			"no pads",
			levelJSON("", ``, okCrystals),
			"pad",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := ParseLevel([]byte(c.json))
			if err == nil {
				t.Fatal("should have failed")
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error is missing %q:\n%v", c.want, err)
			}
		})
	}
}

// All problems should be reported together, not just the first one.
func TestValidateReportsAllProblems(t *testing.T) {
	bad := levelJSON("", `{"name": "narrow", "x0": 100, "x1": 140, "y": 500}`, `[[400, 550]]`)
	_, err := ParseLevel([]byte(bad))
	if err == nil {
		t.Fatal("should have failed")
	}
	for _, want := range []string{"wide", "inside the terrain"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error is missing %q:\n%v", want, err)
		}
	}
}
