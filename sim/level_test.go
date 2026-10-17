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
		t.Fatalf("banene er ikke gyldige:\n%v", err)
	}
	if len(ls) == 0 {
		t.Fatal("ingen baner")
	}
	for _, l := range ls {
		if l.Name == "" {
			t.Error("bane uten navn")
		}
	}
}

// Hver levert bane: skipet starter stille på startplattformen og fyller tanken.
func TestShippedLevelsSpawnAndRefuel(t *testing.T) {
	ls, err := LoadDir(levels.FS)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range ls {
		w := NewWorld(l)
		if w.Ship.State != Landed {
			t.Errorf("%s: skipet skal starte Landed", l.Name)
		}
		w.Ship.Fuel = 0.5
		run(w, 180, Input{}) // 3 s: mer enn nok til å nå full tank, så min(1, ...) tar over
		if w.Ship.State != Landed || w.Ship.Fuel != 1 {
			t.Errorf("%s: state=%v fuel=%v etter 3 s uten input", l.Name, w.Ship.State, w.Ship.Fuel)
		}
	}
}

// En liten, gyldig bane: gulv på y=500, plattform 100..300, én krystall.
// Testene under endrer én ting om gangen og sjekker at feilen blir funnet.
func levelJSON(polyExtra, pad, crystals string) string {
	return fmt.Sprintf(`{
  "name": "test",
  "size": [800, 600],
  "polygons": [
    {"name": "gulv", "points": [[0,500],[800,500],[800,600],[0,600]]}%s
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
		t.Fatalf("burde være gyldig: %v", err)
	}
	if l.Name != "test" || len(l.Crystals) != 1 || len(l.Pads) != 1 {
		t.Errorf("uventet innhold: %+v", l)
	}
}

func TestParseRejectsBadLevels(t *testing.T) {
	cases := []struct {
		name string
		json string
		want string // delstreng i feilmeldingen
	}{
		{
			"kryssende polygoner",
			levelJSON(`, {"name": "kryss", "points": [[500,450],[600,450],[600,550],[500,550]]}`, okPad, okCrystals),
			"krysser",
		},
		{
			"for smal plattform",
			levelJSON("", `{"name": "smal", "x0": 100, "x1": 140, "y": 500}`, okCrystals),
			"bred",
		},
		{
			"plattform som svever",
			levelJSON("", `{"name": "svever", "x0": 100, "x1": 300, "y": 480}`, okCrystals),
			"ikke flat",
		},
		{
			"krystall inni terrenget",
			levelJSON("", okPad, `[[400, 550]]`),
			"inni terrenget",
		},
		{
			"krystall tett på terrenget",
			levelJSON("", okPad, `[[400, 480]]`),
			"innenfor",
		},
		{
			"ingen krystaller",
			levelJSON("", okPad, `[]`),
			"krystaller",
		},
		{
			"terreng rett over plattformen",
			levelJSON(`, {"name": "tak", "points": [[100,450],[300,450],[300,430],[100,430]]}`, okPad, okCrystals),
			"over seg",
		},
		{
			"ukjent felt",
			strings.Replace(levelJSON("", okPad, okCrystals), `"name": "test",`, `"name": "test", "pdas": [],`, 1),
			"pdas",
		},
		{
			"ingen plattformer",
			levelJSON("", ``, okCrystals),
			"plattform",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := ParseLevel([]byte(c.json))
			if err == nil {
				t.Fatal("burde ha gitt feil")
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("feilen mangler %q:\n%v", c.want, err)
			}
		})
	}
}

// Alle problemer skal rapporteres samlet, ikke bare det første.
func TestValidateReportsAllProblems(t *testing.T) {
	bad := levelJSON("", `{"name": "smal", "x0": 100, "x1": 140, "y": 500}`, `[[400, 550]]`)
	_, err := ParseLevel([]byte(bad))
	if err == nil {
		t.Fatal("burde ha gitt feil")
	}
	for _, want := range []string{"bred", "inni terrenget"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("feilen mangler %q:\n%v", want, err)
		}
	}
}
