package sim

import (
	"fmt"
	"strings"
	"testing"
	"testing/fstest"

	"raketti/levels"
)

func TestShippedLevelsLoadAndValidate(t *testing.T) {
	ls, err := LoadDir(levels.Crystals, ModeCrystals)
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

func TestShippedCargoLevelsLoadAndValidate(t *testing.T) {
	ls, err := LoadDir(levels.Cargo, ModeCargo)
	if err != nil {
		t.Fatalf("the cargo levels are not valid:\n%v", err)
	}
	for _, l := range ls {
		w := NewWorld(l)
		run(w, 60, Input{})
		if w.Ship.State != Landed || w.Cargo.Attached || w.Won {
			t.Errorf("%s: after 1 s without input: ship %v, cargo attached %v, won %v",
				l.Name, w.Ship.State, w.Cargo.Attached, w.Won)
		}
	}
}

// Every shipped level: the ship starts at rest on the start pad and refuels.
func TestShippedLevelsSpawnAndRefuel(t *testing.T) {
	ls, err := LoadDir(levels.Crystals, ModeCrystals)
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
	l, err := ParseLevel([]byte(levelJSON("", okPad, okCrystals)), ModeCrystals)
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
			_, err := ParseLevel([]byte(c.json), ModeCrystals)
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
	_, err := ParseLevel([]byte(bad), ModeCrystals)
	if err == nil {
		t.Fatal("should have failed")
	}
	for _, want := range []string{"wide", "inside the terrain"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error is missing %q:\n%v", want, err)
		}
	}
}

// Levels are sorted by the number at the start of the file name, not as text.
func TestLoadDirSortsByNumber(t *testing.T) {
	level := func(name string) *fstest.MapFile {
		json := strings.Replace(levelJSON("", okPad, okCrystals), `"name": "test"`, `"name": "`+name+`"`, 1)
		return &fstest.MapFile{Data: []byte(json)}
	}
	fsys := fstest.MapFS{
		"100-c.json": level("c"),
		"20-b.json":  level("b"),
		"10-a.json":  level("a"),
	}
	ls, err := LoadDir(fsys, ModeCrystals)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, l := range ls {
		got = append(got, l.Name)
	}
	if strings.Join(got, " ") != "a b c" {
		t.Errorf("order = %v, want [a b c]", got)
	}
}

func TestLoadDirRejectsBadFileNames(t *testing.T) {
	data := &fstest.MapFile{Data: []byte(levelJSON("", okPad, okCrystals))}
	cases := []struct {
		name string
		fsys fstest.MapFS
		want string
	}{
		{"no number", fstest.MapFS{"grotta.json": data}, "must start with a number"},
		{"same number", fstest.MapFS{"10-a.json": data, "10-b.json": data}, "same number"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := LoadDir(c.fsys, ModeCrystals)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v, want it to contain %q", err, c.want)
			}
		})
	}
}

// A small, valid cargo level: floor at y=500, start pad 100..250, goal pad
// "depot" 500..650, cargo resting on the floor at x=350.
func cargoJSON(polyExtra, cargo, goal string) string {
	return fmt.Sprintf(`{
  "name": "cargo test",
  "size": [800, 600],
  "polygons": [
    {"name": "floor", "points": [[0,500],[800,500],[800,600],[0,600]]}%s
  ],
  "pads": [
    {"name": "start", "x0": 100, "x1": 250, "y": 500},
    {"name": "depot", "x0": 500, "x1": 650, "y": 500}
  ],
  "cargo": %s,
  "goal": %s
}`, polyExtra, cargo, goal)
}

const (
	okCargo = `[350, 491.5]`
	okGoal  = `"depot"`
)

func TestParseValidCargoLevel(t *testing.T) {
	l, err := ParseLevel([]byte(cargoJSON("", okCargo, okGoal)), ModeCargo)
	if err != nil {
		t.Fatalf("should be valid: %v", err)
	}
	if l.Mode != ModeCargo || l.Goal != 1 || l.CargoStart != (Pt{350, 491.5}) {
		t.Errorf("mode=%v goal=%d cargo=%v", l.Mode, l.Goal, l.CargoStart)
	}
	w := NewWorld(l)
	if w.Cargo.X != 350 || w.Cargo.Attached {
		t.Errorf("the world should start with the cargo resting at its start: %+v", w.Cargo)
	}
}

func TestParseRejectsBadCargoLevels(t *testing.T) {
	cases := []struct {
		name string
		json string
		want string
	}{
		{"cargo inside the terrain", cargoJSON("", `[350, 520]`, okGoal), "overlaps the terrain"},
		{"cargo floating", cargoJSON("", `[350, 400]`, okGoal), "rest on the ground"},
		{
			"no room above the cargo",
			cargoJSON(`, {"name": "roof", "points": [[300,470],[400,470],[400,450],[300,450]]}`, okCargo, okGoal),
			"no room above",
		},
		{"goal is not a pad", cargoJSON("", okCargo, `"nowhere"`), "not the name of a pad"},
		{"goal is the start pad", cargoJSON("", okCargo, `"start"`), "start pad"},
		{"cargo starts on the goal", cargoJSON("", `[575, 491.5]`, okGoal), "starts on the goal pad"},
		{
			"crystals in a cargo level",
			strings.Replace(cargoJSON("", okCargo, okGoal), `"goal"`, `"crystals": [[400, 300]], "goal"`, 1),
			"crystals",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := ParseLevel([]byte(c.json), ModeCargo)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v, want it to contain %q", err, c.want)
			}
		})
	}
}

func TestCrystalsLevelRejectsCargoFields(t *testing.T) {
	json := strings.Replace(levelJSON("", okPad, okCrystals), `"crystals"`, `"cargo": [350, 491.5], "crystals"`, 1)
	_, err := ParseLevel([]byte(json), ModeCrystals)
	if err == nil || !strings.Contains(err.Error(), "cargo") {
		t.Errorf("err = %v, a crystals level must not have a cargo", err)
	}
}
