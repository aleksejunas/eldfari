// Raketti: vector remake of a gravity game in the style of Thrust/Turboraketti.
// All game logic lives in sim/. This file is only window, input and drawing.
package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"raketti/levels"
	"raketti/sim"
)

const (
	startW = 1280 // initial window size; the window can be resized freely
	startH = 800

	// The HUD text (ebitenutil.DebugPrint) uses a fixed 6x16 px font, so it is drawn
	// to a small image and scaled up with the nearest filter.
	hudScale = 2
)

// Zoom = screen pixels per world pixel. Z cycles through these.
var zoomSteps = []float64{1, 1.5, 2, 3, 4}

var (
	colBG      = color.RGBA{0x2b, 0x26, 0x26, 0xff}
	colTerrain = color.RGBA{0xe8, 0x83, 0x6b, 0xff}
	colPad     = color.RGBA{0xe8, 0xa8, 0x5a, 0xff}
	colShip    = color.RGBA{0xdc, 0xd7, 0xd2, 0xff}
	colGem     = color.RGBA{0xf5, 0xd6, 0x8a, 0xff}
)

type Game struct {
	levels []*sim.Level
	lvl    int // index into levels
	best   *bestTimes

	world      *sim.World
	parts      []particle
	camX, camY float64 // top-left corner of the view, in world coordinates
	prev       sim.State

	w, h    int // logical screen size = window size (set in Layout)
	zoomIdx int
	hud     *ebiten.Image

	tick   int // counts Update calls, used for animation
	wasWon bool
}

func newGame() *Game {
	ls, err := sim.LoadDir(levels.FS)
	if err != nil {
		log.Fatalf("level files:\n%v", err)
	}
	g := &Game{
		levels:  ls,
		best:    loadBestTimes(),
		w:       startW,
		h:       startH,
		zoomIdx: 2, // 2x
	}
	g.setLevel(0)
	return g
}

// setLevel switches level (with wrap-around) and restarts.
func (g *Game) setLevel(i int) {
	n := len(g.levels)
	g.lvl = ((i % n) + n) % n
	g.world = sim.NewWorld(g.levels[g.lvl])
	g.reset()
}

func (g *Game) zoom() float64 { return zoomSteps[g.zoomIdx] }

func (g *Game) reset() {
	g.world.Reset()
	g.parts = nil
	g.prev = sim.Landed
	g.wasWon = false
	g.camX, g.camY = g.cameraTarget()
}

// clampCam keeps the view inside the world. If the world is smaller than the view,
// it is centred instead.
func clampCam(target, view, world float64) float64 {
	if world <= view {
		return (world - view) / 2
	}
	return math.Max(0, math.Min(target, world-view))
}

func (g *Game) cameraTarget() (float64, float64) {
	s := &g.world.Ship
	l := g.world.Level
	z := g.zoom()
	vw, vh := float64(g.w)/z, float64(g.h)/z
	return clampCam(s.X-vw/2, vw, float64(l.W)), clampCam(s.Y-vh/2, vh, float64(l.H))
}

func (g *Game) Update() error {
	g.tick++
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.reset()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		g.setLevel(g.lvl + 1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.setLevel(g.lvl - 1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		g.zoomIdx = (g.zoomIdx + 1) % len(zoomSteps)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	in := sim.Input{
		Left:   ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA),
		Right:  ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD),
		Thrust: ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeySpace),
	}
	g.world.Step(in)

	if g.world.Won && !g.wasWon {
		g.best.record(g.world.Level.Name, g.world.Time)
	}
	g.wasWon = g.world.Won

	s := &g.world.Ship
	if s.Thrusting {
		for i := 0; i < 2; i++ {
			g.parts = append(g.parts, exhaust(s))
		}
	}
	if s.State == sim.Crashed && g.prev != sim.Crashed {
		g.parts = append(g.parts, explosion(s)...)
	}
	g.prev = s.State

	alive := g.parts[:0]
	for i := range g.parts {
		p := g.parts[i]
		p.update()
		if p.life > 0 {
			alive = append(alive, p)
		}
	}
	g.parts = alive

	tx, ty := g.cameraTarget()
	g.camX += (tx - g.camX) * 0.12
	g.camY += (ty - g.camY) * 0.12
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(colBG)
	l := g.world.Level
	s := &g.world.Ship
	z := g.zoom()

	// ln draws a world line segment: offset by the camera, scaled by zoom.
	// Line width follows zoom, but a bit weaker so it doesn't get chunky.
	ln := func(x0, y0, x1, y1, w float64, c color.Color) {
		line(screen, (x0-g.camX)*z, (y0-g.camY)*z, (x1-g.camX)*z, (y1-g.camY)*z, w*z*0.75, c)
	}

	for _, poly := range l.Polys {
		for i := range poly {
			a, b := poly[i], poly[(i+1)%len(poly)]
			ln(a.X, a.Y, b.X, b.Y, 2, colTerrain)
		}
	}
	for _, pad := range l.Pads {
		ln(pad.X0, pad.Y, pad.X1, pad.Y, 4, colPad)
	}

	// Crystals: a diamond that pulses and "spins" by changing its width.
	for i, c := range l.Crystals {
		if g.world.Got[i] {
			continue
		}
		t := float64(g.tick)
		r := 11 + 1.5*math.Sin(t/12+float64(i))
		hw := 2 + 8*math.Abs(math.Cos(t/30+float64(i)))
		ln(c.X, c.Y-r, c.X+hw, c.Y, 2, colGem)
		ln(c.X+hw, c.Y, c.X, c.Y+r, 2, colGem)
		ln(c.X, c.Y+r, c.X-hw, c.Y, 2, colGem)
		ln(c.X-hw, c.Y, c.X, c.Y-r, 2, colGem)
		ln(c.X, c.Y-r, c.X, c.Y+r, 1, colGem)
	}

	if s.State != sim.Crashed {
		v := s.Points()
		for i := range v {
			a, b := v[i], v[(i+1)%len(v)]
			ln(a.X, a.Y, b.X, b.Y, 2, colShip)
		}
		if s.Thrusting {
			length := 8 + rand.Float64()*8
			a := s.ToWorld(-4, sim.ShipRear)
			b := s.ToWorld(4, sim.ShipRear)
			tip := s.ToWorld(0, sim.ShipRear+length)
			ln(a.X, a.Y, tip.X, tip.Y, 2, colPad)
			ln(b.X, b.Y, tip.X, tip.Y, 2, colPad)
		}
	}

	for _, p := range g.parts {
		dx, dy := math.Cos(p.ang)*p.length/2, math.Sin(p.ang)*p.length/2
		c := fade(colTerrain, p.life/p.maxLife)
		ln(p.x-dx, p.y-dy, p.x+dx, p.y+dy, 2, c)
	}

	g.drawHUD(screen)
}

// fmtTime formats as m:ss.s
func fmtTime(t float64) string {
	m := int(t) / 60
	return fmt.Sprintf("%d:%04.1f", m, t-float64(m*60))
}

// drawHUD draws the text to a small image (hudScale times smaller than the screen)
// and scales it up, since the DebugPrint font is tiny otherwise.
func (g *Game) drawHUD(screen *ebiten.Image) {
	hw, hh := g.w/hudScale, g.h/hudScale
	if g.hud == nil || g.hud.Bounds().Dx() != hw || g.hud.Bounds().Dy() != hh {
		g.hud = ebiten.NewImage(hw, hh)
	}
	g.hud.Clear()

	w := g.world
	s := &w.Ship
	total := len(w.Level.Crystals)

	ebitenutil.DebugPrintAt(g.hud, "FUEL", 8, 4)
	line(g.hud, 40, 13, 40+60*s.Fuel, 13, 5, colPad)
	ebitenutil.DebugPrintAt(g.hud, fmt.Sprintf("VX %+4.0f  VY %+4.0f", s.VX, s.VY), 8, 20)
	ebitenutil.DebugPrintAt(g.hud, fmt.Sprintf("LEVEL %d/%d  %s", g.lvl+1, len(g.levels), w.Level.Name), 8, 36)

	stats := fmt.Sprintf("CRYSTALS %d/%d   TIME %s", w.NGot, total, fmtTime(w.Time))
	if b := g.best.get(w.Level.Name); b > 0 {
		stats += "   BEST " + fmtTime(b)
	}
	ebitenutil.DebugPrintAt(g.hud, stats, hw-len(stats)*6-8, 4)

	msg := ""
	switch {
	case w.Won:
		msg = "YOU WIN! TIME " + fmtTime(w.Time) + " - R retry, N next level"
	case s.State == sim.Crashed:
		msg = "CRASHED - press R"
	case w.NGot == total && total > 0:
		msg = "ALL CRYSTALS - LAND ON THE START PAD"
	case s.State == sim.Landed:
		msg = "LANDED - thrust to lift off"
	}
	if msg != "" {
		ebitenutil.DebugPrintAt(g.hud, msg, (hw-len(msg)*6)/2, 52)
	}
	help := fmt.Sprintf("A/D rotate | W/space thrust | R restart | N/P level | Z zoom %.1fx | F fullscreen", g.zoom())
	ebitenutil.DebugPrintAt(g.hud, help, 8, hh-20)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(hudScale, hudScale)
	op.Filter = ebiten.FilterNearest
	screen.DrawImage(g.hud, op)
}

// Layout: logical size follows the window, so nothing is stretched or blurred.
func (g *Game) Layout(outsideW, outsideH int) (int, int) {
	g.w, g.h = max(outsideW, 320), max(outsideH, 200)
	return g.w, g.h
}

func line(dst *ebiten.Image, x0, y0, x1, y1, w float64, c color.Color) {
	vector.StrokeLine(dst, float32(x0), float32(y0), float32(x1), float32(y1), float32(w), c, true)
}

// fade scales a colour towards transparent. Ebitengine uses premultiplied alpha.
func fade(c color.RGBA, a float64) color.RGBA {
	a = math.Max(0, math.Min(1, a))
	return color.RGBA{
		R: uint8(float64(c.R) * a),
		G: uint8(float64(c.G) * a),
		B: uint8(float64(c.B) * a),
		A: uint8(255 * a),
	}
}

func main() {
	ebiten.SetWindowSize(startW, startH)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Raketti")
	if err := ebiten.RunGame(newGame()); err != nil {
		log.Fatal(err)
	}
}
