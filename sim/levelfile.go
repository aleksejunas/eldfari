package sim

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"sort"
	"strings"
)

// Banefilformat (levels/*.json). Koordinater er piksler, y peker nedover.
//
//	{
//	  "name": "Grotta",
//	  "size": [3200, 1400],
//	  "polygons": [{"name": "gulv", "points": [[0,1100], [320,1100], ...]}],
//	  "pads":     [{"name": "start", "x0": 100, "x1": 240, "y": 1100}],
//	  "crystals": [[600, 700]]
//	}
//
// pads[0] er startplattformen: skipet spawner der, og man vinner ved å lande
// der med alle krystaller tatt. Alt utenfor verdenen regnes som fast stoff.
// Polygonene fylles even-odd samlet, så kantene deres må ikke krysse hverandre.

const (
	minPadWidth      = 60 // px
	padClearance     = 60 // fri luft over en plattform, px
	crystalClearance = 30 // fri luft rundt en krystall, px
	maxWorldSize     = 8192
)

type polyFile struct {
	Name   string       `json:"name"`
	Points [][2]float64 `json:"points"`
}

type padFile struct {
	Name string  `json:"name"`
	X0   float64 `json:"x0"`
	X1   float64 `json:"x1"`
	Y    float64 `json:"y"`
}

type levelFile struct {
	Name     string       `json:"name"`
	Size     [2]int       `json:"size"`
	Polygons []polyFile   `json:"polygons"`
	Pads     []padFile    `json:"pads"`
	Crystals [][2]float64 `json:"crystals"`
}

// LoadDir leser alle *.json i fsys (sortert på filnavn) og validerer dem.
func LoadDir(fsys fs.FS) ([]*Level, error) {
	names, err := fs.Glob(fsys, "*.json")
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, errors.New("fant ingen banefiler (*.json)")
	}
	sort.Strings(names)
	var out []*Level
	for _, n := range names {
		data, err := fs.ReadFile(fsys, n)
		if err != nil {
			return nil, err
		}
		l, err := ParseLevel(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", n, err)
		}
		if l.Name == "" {
			l.Name = strings.TrimSuffix(n, ".json")
		}
		out = append(out, l)
	}
	return out, nil
}

// ParseLevel leser én banefil og validerer den. Ukjente felt er en feil,
// så skrivefeil i nøkler blir oppdaget i stedet for å bli ignorert.
func ParseLevel(data []byte) (*Level, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var f levelFile
	if err := dec.Decode(&f); err != nil {
		return nil, fmt.Errorf("ugyldig JSON: %w", err)
	}
	w, h := f.Size[0], f.Size[1]
	if w < 100 || h < 100 || w > maxWorldSize || h > maxWorldSize {
		return nil, fmt.Errorf("size %v må være mellom 100 og %d i begge retninger", f.Size, maxWorldSize)
	}
	if len(f.Pads) == 0 {
		return nil, errors.New("banen trenger minst én plattform (pads[0] er start)")
	}

	polys := make([][]Pt, len(f.Polygons))
	polyNames := make([]string, len(f.Polygons))
	for i, p := range f.Polygons {
		polyNames[i] = p.Name
		for _, q := range p.Points {
			polys[i] = append(polys[i], Pt{q[0], q[1]})
		}
	}
	pads := make([]Pad, len(f.Pads))
	padNames := make([]string, len(f.Pads))
	for i, p := range f.Pads {
		pads[i] = Pad{X0: p.X0, X1: p.X1, Y: p.Y}
		padNames[i] = p.Name
	}
	start := Pt{X: (pads[0].X0 + pads[0].X1) / 2, Y: pads[0].Y - ShipRear - 0.5}

	l := NewLevel(w, h, polys, pads, start)
	l.Name = f.Name
	l.polyNames, l.padNames = polyNames, padNames
	for _, c := range f.Crystals {
		l.Crystals = append(l.Crystals, Pt{c[0], c[1]})
	}
	if err := l.Validate(); err != nil {
		return nil, err
	}
	return l, nil
}

func nameOr(names []string, i int) string {
	if i < len(names) && names[i] != "" {
		return fmt.Sprintf("%q", names[i])
	}
	return fmt.Sprintf("#%d", i)
}

func cross(o, a, b Pt) float64 { return (a.X-o.X)*(b.Y-o.Y) - (a.Y-o.Y)*(b.X-o.X) }

// segmentsCross: sant hvis linjestykkene a-b og c-d krysser hverandre på tvers.
func segmentsCross(a, b, c, d Pt) bool {
	d1, d2 := cross(a, b, c), cross(a, b, d)
	d3, d4 := cross(c, d, a), cross(c, d, b)
	return d1*d2 < 0 && d3*d4 < 0
}

// Validate samler alle problemer i banen (ikke bare det første), slik at du kan
// rette alt på én gang.
func (l *Level) Validate() error {
	var errs []error
	addf := func(format string, a ...any) { errs = append(errs, fmt.Errorf(format, a...)) }
	inWorld := func(x, y float64) bool {
		return x >= 0 && x <= float64(l.W) && y >= 0 && y <= float64(l.H)
	}

	if len(l.Polys) == 0 {
		addf("banen har ingen polygoner")
	}
	for i, p := range l.Polys {
		if len(p) < 3 {
			addf("polygon %s har bare %d punkter (minst 3)", nameOr(l.polyNames, i), len(p))
		}
		for j, pt := range p {
			if !inWorld(pt.X, pt.Y) {
				addf("polygon %s punkt %d (%.0f,%.0f) er utenfor verdenen", nameOr(l.polyNames, i), j, pt.X, pt.Y)
			}
		}
	}

	// Kanter som krysser hverandre ødelegger even-odd-fyllingen (hull i terrenget).
	type edge struct {
		poly, idx int
		a, b      Pt
	}
	var edges []edge
	for i, p := range l.Polys {
		for j := range p {
			edges = append(edges, edge{i, j, p[j], p[(j+1)%len(p)]})
		}
	}
	for i := range edges {
		for j := i + 1; j < len(edges); j++ {
			ei, ej := edges[i], edges[j]
			if ei.poly == ej.poly {
				n := len(l.Polys[ei.poly])
				if ej.idx == ei.idx+1 || (ei.idx == 0 && ej.idx == n-1) {
					continue // naboer deler et hjørne
				}
			}
			if segmentsCross(ei.a, ei.b, ej.a, ej.b) {
				addf("kant %d i polygon %s krysser kant %d i polygon %s",
					ei.idx, nameOr(l.polyNames, ei.poly), ej.idx, nameOr(l.polyNames, ej.poly))
			}
		}
	}

	for i, p := range l.Pads {
		name := nameOr(l.padNames, i)
		if p.X1-p.X0 < minPadWidth {
			addf("plattform %s er bare %.0f px bred (minst %d)", name, p.X1-p.X0, minPadWidth)
		}
		if !inWorld(p.X0, p.Y) || !inWorld(p.X1, p.Y) {
			addf("plattform %s ligger utenfor verdenen", name)
			continue
		}
		for x := p.X0; x <= p.X1; x += 4 {
			if !l.Solid(x, p.Y) || l.Solid(x, p.Y-1) {
				addf("plattform %s er ikke flat og solid ved x=%.0f (terrengets topp må ligge på y=%.0f)", name, x, p.Y)
				break
			}
		}
	clear:
		for x := p.X0; x <= p.X1; x += 4 {
			for y := p.Y - padClearance; y < p.Y; y += 2 {
				if l.Solid(x, y) {
					addf("plattform %s har terreng innenfor %d px over seg ved (%.0f,%.0f)", name, padClearance, x, y)
					break clear
				}
			}
		}
	}

	if len(l.Crystals) == 0 {
		addf("banen har ingen krystaller, så den kan ikke vinnes")
	}
	for i, c := range l.Crystals {
		if !inWorld(c.X, c.Y) {
			addf("krystall %d (%.0f,%.0f) er utenfor verdenen", i, c.X, c.Y)
			continue
		}
		if l.Solid(c.X, c.Y) {
			addf("krystall %d (%.0f,%.0f) ligger inni terrenget", i, c.X, c.Y)
			continue
		}
		for a := 0.0; a < 2*math.Pi; a += math.Pi / 8 {
			if l.Solid(c.X+crystalClearance*math.Cos(a), c.Y+crystalClearance*math.Sin(a)) {
				addf("krystall %d (%.0f,%.0f) har terreng innenfor %d px", i, c.X, c.Y, crystalClearance)
				break
			}
		}
	}

	spawn := Ship{X: l.Start.X, Y: l.Start.Y}
	for _, p := range spawn.contactPoints() {
		if l.Solid(p.X, p.Y) {
			addf("skipet starter inni terrenget (kontaktpunkt %.0f,%.0f)", p.X, p.Y)
			break
		}
	}

	return errors.Join(errs...)
}
