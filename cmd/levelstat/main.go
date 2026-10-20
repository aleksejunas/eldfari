// Command levelstat prints difficulty numbers for the crystals levels: the
// narrowest gap the route must pass and the route length.
//
//	go run ./cmd/levelstat                 # the embedded levels
//	go run ./cmd/levelstat -png /tmp/maps  # also write a picture per level
//	go run ./cmd/levelstat path/to/dir     # levels in another folder
package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"eldfari/levels"
	"eldfari/sim"
)

const shipHeight = 21.0 // px, nose to rear

func main() {
	pngDir := flag.String("png", "", "write a picture of each level and its route to this folder")
	flag.Parse()

	fsys := levels.Crystals
	if flag.NArg() > 0 {
		fsys = os.DirFS(flag.Arg(0))
	}
	ls, err := sim.LoadDir(fsys, sim.ModeCrystals)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%-12s %9s %7s %8s %8s %8s %5s\n", "level", "gap px", "ships", "route", "longest", "crystal", "pads")
	for _, l := range ls {
		g, st := analyse(l)
		if !st.ok {
			fmt.Printf("%-12s some crystal can't be reached with %.0f px clearance\n", l.Name, shipRadius)
			continue
		}
		fmt.Printf("%-12s %9.0f %6.1fx %8.0f %8.0f %8d %5d   narrowest at (%.0f,%.0f)\n",
			l.Name, st.gap, st.gap/shipHeight, st.route, st.longestLeg, len(l.Crystals), len(l.Pads),
			st.gapAt.X, st.gapAt.Y)
		if *pngDir != "" {
			if err := writePNG(*pngDir, l, g, st); err != nil {
				log.Fatal(err)
			}
		}
	}
}

// writePNG draws one pixel per grid cell: terrain, places the ship doesn't
// fit, the route, crystals, pads and a ring around the narrowest gap.
func writePNG(dir string, l *sim.Level, g *grid, st stats) error {
	img := image.NewRGBA(image.Rect(0, 0, g.w, g.h))
	for i, c := range g.clear {
		col := color.RGBA{0x2b, 0x26, 0x26, 0xff}
		switch {
		case c == 0:
			col = color.RGBA{0xe8, 0x83, 0x6b, 0xff}
		case c < shipRadius:
			col = color.RGBA{0x5a, 0x3a, 0x36, 0xff}
		}
		img.Set(i%g.w, i/g.w, col)
	}
	for _, i := range st.path {
		img.Set(i%g.w, i/g.w, color.RGBA{0xdc, 0xd7, 0xd2, 0xff})
	}
	for _, p := range l.Pads {
		for x := p.X0; x <= p.X1; x += cell {
			i := g.index(sim.Pt{X: x, Y: p.Y - 1})
			img.Set(i%g.w, i/g.w, color.RGBA{0xe8, 0xa8, 0x5a, 0xff})
		}
	}
	dot := func(p sim.Pt, r int, col color.RGBA) {
		i := g.index(p)
		cx, cy := i%g.w, i/g.w
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				if d := dx*dx + dy*dy; d <= r*r && d >= (r-1)*(r-1) || r <= 2 && d <= r*r {
					img.Set(cx+dx, cy+dy, col)
				}
			}
		}
	}
	for _, c := range l.Crystals {
		dot(c, 2, color.RGBA{0xf5, 0xd6, 0x8a, 0xff})
	}
	dot(st.gapAt, 8, color.RGBA{0xff, 0x40, 0x40, 0xff})

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, strings.ToLower(l.Name)+".png"), buf.Bytes(), 0o644)
}
