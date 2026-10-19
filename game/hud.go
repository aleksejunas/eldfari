package game

import (
	"bytes"
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gomedium"
	"golang.org/x/image/font/gofont/gomonobold"

	"raketti/sim"
)

// HUD layout. All sizes and positions are for a startH (800 px) tall screen and
// are multiplied by g.scale when drawn.
const (
	margin    = 16.0
	sizeSmall = 12.0 // help text, version, labels
	sizeText  = 15.0
	sizeBig   = 24.0 // time and messages
	fuelBarW  = 140.0
	fuelBarH  = 8.0
)

var colDim = color.RGBA{0x8c, 0x82, 0x7c, 0xff} // labels, help text, version

// Text uses Go Medium/Bold. Numbers that change every frame use Go Mono, where
// all digits are equally wide, so they don't jitter.
var (
	fontText  = mustFont(gomedium.TTF)
	fontTitle = mustFont(gobold.TTF)
	fontNum   = mustFont(gomonobold.TTF)
)

func mustFont(ttf []byte) *text.GoTextFaceSource {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(ttf))
	if err != nil {
		log.Fatalf("font: %v", err)
	}
	return src
}

func (g *Game) face(src *text.GoTextFaceSource, size float64) *text.GoTextFace {
	return &text.GoTextFace{Source: src, Size: size * g.scale}
}

// print draws s with its top at y. align decides whether x is the left edge,
// the centre or the right edge of the text. It returns the text width.
func (g *Game) print(dst *ebiten.Image, s string, x, y float64, face *text.GoTextFace, align text.Align, c color.Color) float64 {
	op := &text.DrawOptions{}
	op.PrimaryAlign = align
	// A dark shadow keeps the text readable on top of terrain lines.
	op.GeoM.Translate(x*g.scale+g.scale, y*g.scale+g.scale)
	op.ColorScale.ScaleWithColor(colBG)
	text.Draw(dst, s, face, op)
	op.GeoM.Translate(-g.scale, -g.scale)
	op.ColorScale.Reset()
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, face, op)
	return text.Advance(s, face) / g.scale
}

// fmtTime formats as m:ss.s
func fmtTime(t float64) string {
	m := int(t) / 60
	return fmt.Sprintf("%d:%04.1f", m, t-float64(m*60))
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	w := g.world
	s := &w.Ship
	W, H := float64(g.w)/g.scale, float64(g.h)/g.scale
	small := g.face(fontText, sizeSmall)
	normal := g.face(fontText, sizeText)
	title := g.face(fontTitle, sizeText)
	big := g.face(fontTitle, sizeBig)
	num := g.face(fontNum, sizeText)
	bigNum := g.face(fontNum, sizeBig)
	blink := g.tick/10%2 == 0

	// Top left: fuel bar, then speed. A speed is shown in the warning colour
	// when it is too fast to land.
	x := margin + g.print(screen, "FUEL ", margin, margin, small, text.AlignStart, colDim)
	fuelCol := colPad
	switch {
	case s.Refuelling && blink:
		fuelCol = colGem
	case s.Fuel < 0.2 && blink:
		fuelCol = colTerrain
	}
	barY := margin + 4
	g.rect(screen, x, barY, fuelBarW*s.Fuel, fuelBarH, fuelCol)
	g.strokeRect(screen, x, barY, fuelBarW, fuelBarH, colDim)

	y := margin + 22
	x = margin + g.print(screen, "VX ", margin, y, small, text.AlignStart, colDim)
	x += g.print(screen, fmt.Sprintf("%+4.0f   ", s.VX), x, y-2, num, text.AlignStart, speedCol(s.VX, sim.MaxLandVX))
	x += g.print(screen, "VY ", x, y, small, text.AlignStart, colDim)
	g.print(screen, fmt.Sprintf("%+4.0f", s.VY), x, y-2, num, text.AlignStart, speedCol(s.VY, sim.MaxLandVY))

	// Top centre: level name and number.
	g.print(screen, w.Level.Name, W/2, margin-2, title, text.AlignCenter, colShip)
	g.print(screen, fmt.Sprintf("LEVEL %d/%d", g.lvl+1, len(g.levels)), W/2, margin+20, small, text.AlignCenter, colDim)

	// Top right: time, crystals and best time.
	right := W - margin
	g.print(screen, fmtTime(w.Time), right, margin-6, bigNum, text.AlignEnd, colShip)
	total := len(w.Level.Crystals)
	gemCol := colShip
	if w.NGot == total {
		gemCol = colGem
	}
	cw := g.print(screen, fmt.Sprintf("%d/%d", w.NGot, total), right, margin+26, num, text.AlignEnd, gemCol)
	g.diamond(screen, right-cw-12, margin+35, colGem)
	if b := g.best.get(w.Level.Name); b > 0 {
		g.print(screen, "BEST "+fmtTime(b), right, margin+48, small, text.AlignEnd, colDim)
	}

	// Centre: a message about what to do now.
	msg, sub, msgCol := "", "", colShip
	switch {
	case w.Won:
		msg, sub, msgCol = "YOU WIN!  "+fmtTime(w.Time), "R retry  ·  N next level", colGem
	case s.State == sim.Crashed:
		msg, sub, msgCol = "CRASHED", "R restart", colTerrain
	case w.NGot == total && total > 0:
		msg, sub, msgCol = "ALL CRYSTALS", "land on the start pad", colGem
	case s.State == sim.Landed:
		msg, sub = "LANDED", "thrust to lift off"
	}
	if msg != "" {
		my := H * 0.22
		g.print(screen, msg, W/2, my, big, text.AlignCenter, msgCol)
		g.print(screen, sub, W/2, my+34, normal, text.AlignCenter, colDim)
	}

	// Bottom: key help on the left, version on the right.
	help := fmt.Sprintf("A/D turn  ·  W thrust  ·  R restart  ·  N/P level  ·  Z zoom %gx  ·  F fullscreen", zoomSteps[g.zoomIdx])
	y = H - margin - sizeSmall
	hw := g.print(screen, help, margin, y, small, text.AlignStart, colDim)
	if margin+hw+24+text.Advance(g.version, small)/g.scale > right {
		y -= sizeSmall + 6 // no room next to the help text: one line up
	}
	g.print(screen, g.version, right, y, small, text.AlignEnd, colDim)
}

// speedCol is the warning colour when v is too fast to land.
func speedCol(v, limit float64) color.Color {
	if math.Abs(v) > limit {
		return colTerrain
	}
	return colShip
}

func (g *Game) rect(dst *ebiten.Image, x, y, w, h float64, c color.Color) {
	k := g.scale
	vector.FillRect(dst, float32(x*k), float32(y*k), float32(w*k), float32(h*k), c, true)
}

func (g *Game) strokeRect(dst *ebiten.Image, x, y, w, h float64, c color.Color) {
	k := g.scale
	vector.StrokeRect(dst, float32(x*k), float32(y*k), float32(w*k), float32(h*k), float32(k), c, true)
}

// diamond draws a small crystal icon centred on (x, y).
func (g *Game) diamond(dst *ebiten.Image, x, y float64, c color.Color) {
	k := g.scale
	r, hw := 7.0, 5.0
	pts := [][2]float64{{x, y - r}, {x + hw, y}, {x, y + r}, {x - hw, y}}
	for i, a := range pts {
		b := pts[(i+1)%len(pts)]
		line(dst, a[0]*k, a[1]*k, b[0]*k, b[1]*k, 1.5*k, c)
	}
}
