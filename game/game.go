// Package game is the Ebitengine front end: window, input, camera, drawing and
// HUD. All game logic lives in sim/.
package game

import (
	"image/color"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"raketti/levels"
	"raketti/sim"
)

const (
	startW = 1280 // initial window size; the window can be resized freely
	startH = 800
)

// Zoom = window pixels per world pixel. Z cycles through these.
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

	w, h    int     // screen size in real pixels (set in Layout)
	scale   float64 // g.h / startH: the HUD and zoom are designed for a startH tall screen
	zoomIdx int

	tick    int    // counts Update calls, used for animation
	mode    string // folder in levels/; only "crystals" for now
	version string
	wasWon  bool
}

func newGame() *Game {
	ls, err := sim.LoadDir(levels.Crystals)
	if err != nil {
		log.Fatalf("level files:\n%v", err)
	}
	g := &Game{
		levels:  ls,
		best:    loadBestTimes(),
		w:       startW,
		h:       startH,
		scale:   1,
		zoomIdx: 2, // 2x
		version: version(),
		mode:    "crystals",
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

// bestKey is the key for the current level's best time: mode and level name.
func (g *Game) bestKey() string { return g.mode + "/" + g.world.Level.Name }

// zoom returns screen pixels per world pixel.
func (g *Game) zoom() float64 { return zoomSteps[g.zoomIdx] * g.scale }

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
		g.best.record(g.bestKey(), g.world.Time)
	}
	g.wasWon = g.world.Won

	s := &g.world.Ship
	if s.Thrusting {
		for i := 0; i < 2; i++ {
			g.parts = append(g.parts, exhaust(s))
		}
	}
	if s.Refuelling && g.tick%3 == 0 {
		g.parts = append(g.parts, refuelSpark(s))
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
	ln := func(x0, y0, x1, y1, w float64, c color.RGBA) {
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

	// Arrows on a ring around the ship towards targets that are off screen:
	// the nearest crystal, and the pad to head for.
	vw, vh := float64(g.w)/z, float64(g.h)/z
	arrow := func(tx, ty float64, c color.RGBA) {
		if tx > g.camX && tx < g.camX+vw && ty > g.camY && ty < g.camY+vh {
			return
		}
		a := math.Atan2(ty-s.Y, tx-s.X)
		dx, dy := math.Cos(a), math.Sin(a)
		tipX, tipY := s.X+dx*34, s.Y+dy*34
		bx, by := tipX-dx*10, tipY-dy*10
		lx, ly := bx-dy*5, by+dx*5
		rx, ry := bx+dy*5, by-dx*5
		ln(tipX, tipY, lx, ly, 2, c)
		ln(lx, ly, rx, ry, 2, c)
		ln(rx, ry, tipX, tipY, 2, c)
	}
	if s.State != sim.Crashed && !g.world.Won {
		if c, ok := g.world.NearestCrystal(); ok {
			arrow(c.X, c.Y, colGem)
		}
		pad := g.world.GoalPad()
		arrow((pad.X0+pad.X1)/2, pad.Y, colPad)
	}

	// Particles are tiny, so they skip antialiasing: Ebitengine then draws
	// them all in one batch.
	for _, p := range g.parts {
		dx, dy := math.Cos(p.ang)*p.length/2, math.Sin(p.ang)*p.length/2
		x0, y0 := (p.x-dx-g.camX)*z, (p.y-dy-g.camY)*z
		x1, y1 := (p.x+dx-g.camX)*z, (p.y+dy-g.camY)*z
		c := fade(p.col, p.life/p.maxLife)
		vector.StrokeLine(screen, float32(x0), float32(y0), float32(x1), float32(y1), float32(1.5*z), c, false)
	}

	g.drawHUD(screen)
}

// Layout draws at the screen's real resolution (window size × the screen's
// scale factor), so lines and text stay sharp on HiDPI screens. Everything is
// then scaled by the screen height, so the game looks the same at any size.
func (g *Game) Layout(outsideW, outsideH int) (int, int) {
	dsf := ebiten.Monitor().DeviceScaleFactor()
	g.w = int(float64(max(outsideW, 320)) * dsf)
	g.h = int(float64(max(outsideH, 200)) * dsf)
	g.scale = float64(g.h) / startH
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

// Run opens the window and runs the game until the window is closed.
func Run() error {
	ebiten.SetWindowSize(startW, startH)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Raketti")
	log.Printf("Raketti %s", version())
	return ebiten.RunGame(newGame())
}
