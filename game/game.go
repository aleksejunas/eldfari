// Package game is the Ebitengine front end: window, input, camera, drawing and
// HUD. All game logic lives in sim/.
package game

import (
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"eldfari/levels"
	"eldfari/sim"
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

// modeNames are the folders in levels/, also used in best-time keys and the HUD.
var modeNames = map[sim.Mode]string{sim.ModeCrystals: "crystals", sim.ModeCargo: "cargo", sim.ModeDuel: "duel", sim.ModeRace: "race"}

type Game struct {
	levels map[sim.Mode][]*sim.Level
	mode   sim.Mode
	lvl    int // index into levels[mode]
	best   *bestTimes

	world      *sim.World // in a duel: player 1's
	duel       *sim.Duel  // nil outside duel mode; see twoplayer.go
	race       *sim.Race  // nil outside race mode
	views      [2]duelView
	parts      []particle
	camX, camY float64 // top-left corner of the view, in world coordinates
	prev       sim.State

	w, h    int     // screen size in real pixels (set in Layout)
	scale   float64 // g.h / startH: the HUD and zoom are designed for a startH tall screen
	zoomIdx int

	tick    int // counts Update calls, used for animation
	version string
	wasWon  bool
	showFPS bool // FPS overlay, toggled with I
	fps     fpsMeter

	inMenu bool // the mode and level chooser is shown instead of the game
	menu   menu

	sticks    [2]stickState // per player, for the stick's dead zone
	menuStick [2]int        // stick direction last step, for menu steps
}

func newGame() *Game {
	crystals, err := sim.LoadDir(levels.Crystals, sim.ModeCrystals)
	if err != nil {
		log.Fatalf("crystals level files:\n%v", err)
	}
	cargo, err := sim.LoadDir(levels.Cargo, sim.ModeCargo)
	if err != nil {
		log.Fatalf("cargo level files:\n%v", err)
	}
	duel, err := sim.LoadDir(levels.Duel, sim.ModeDuel)
	if err != nil {
		log.Fatalf("duel level files:\n%v", err)
	}
	race, err := sim.LoadDir(levels.Race, sim.ModeRace)
	if err != nil {
		log.Fatalf("race level files:\n%v", err)
	}
	g := &Game{
		levels:  map[sim.Mode][]*sim.Level{sim.ModeCrystals: crystals, sim.ModeCargo: cargo, sim.ModeDuel: duel, sim.ModeRace: race},
		best:    loadBestTimes(),
		w:       startW,
		h:       startH,
		scale:   1,
		zoomIdx: 2, // 2x
		version: version(),
	}
	g.setLevel(0)
	g.openMenu()
	return g
}

// setLevel switches level within the current mode (with wrap-around) and
// restarts.
func (g *Game) setLevel(i int) {
	ls := g.levels[g.mode]
	n := len(ls)
	g.lvl = ((i % n) + n) % n
	g.duel, g.race = nil, nil
	switch g.mode {
	case sim.ModeDuel:
		g.duel = sim.NewDuel(ls[g.lvl])
		g.world = g.duel.Players[0].World
	case sim.ModeRace:
		g.race = sim.NewRace(ls[g.lvl])
		g.world = g.race.Players[0].World
	default:
		g.world = sim.NewWorld(ls[g.lvl])
	}
	g.reset()
}

// bestKey is the key for the current level's best time: mode and level name.
func (g *Game) bestKey() string { return modeNames[g.mode] + "/" + g.world.Level.Name }

// zoom returns screen pixels per world pixel.
func (g *Game) zoom() float64 { return zoomSteps[g.zoomIdx] * g.scale }

func (g *Game) reset() {
	g.parts = nil
	if g.players() != nil {
		g.resetTwoPlayer()
		return
	}
	g.world.Reset()
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
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	if g.inMenu {
		g.updateMenu()
		return nil
	}
	ids := pads()
	// Pads: Start = menu, Back = restart, LB/RB = previous/next level, Y = zoom.
	padKey := func(b ebiten.StandardGamepadButton) bool { return padJustPressed(ids, b) }
	if inpututil.IsKeyJustPressed(ebiten.KeyM) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) ||
		padKey(ebiten.StandardGamepadButtonCenterRight) {
		g.openMenu()
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) || padKey(ebiten.StandardGamepadButtonCenterLeft) {
		g.reset()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyN) || padKey(ebiten.StandardGamepadButtonFrontTopRight) {
		g.setLevel(g.lvl + 1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyP) || padKey(ebiten.StandardGamepadButtonFrontTopLeft) {
		g.setLevel(g.lvl - 1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) || padKey(ebiten.StandardGamepadButtonRightTop) {
		g.zoomIdx = (g.zoomIdx + 1) % len(zoomSteps)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		g.showFPS = !g.showFPS
	}
	if g.players() != nil {
		g.updateTwoPlayer(ids)
		return nil
	}

	g.world.Step(g.readInput(-1, ids))

	if g.world.Won && !g.wasWon {
		g.best.record(g.bestKey(), g.world.Time)
	}
	g.wasWon = g.world.Won

	g.shipEffects(&g.world.Ship, &g.prev)
	g.updateParticles()

	tx, ty := g.cameraTarget()
	g.camX += (tx - g.camX) * 0.12
	g.camY += (ty - g.camY) * 0.12
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.fps.frame(time.Now())
	if g.inMenu {
		g.drawMenu(screen)
		return
	}
	if g.players() != nil {
		g.drawTwoPlayer(screen)
		return
	}
	screen.Fill(colBG)
	g.drawScene(screen, g.camX, g.camY, -1)
	g.drawHUD(screen)
}

// drawScene draws the level, the ships, the arrows and the particles onto dst,
// seen from camX, camY. me is the player whose view it is in a duel (the
// arrow points to the other one), or -1 outside a duel.
func (g *Game) drawScene(dst *ebiten.Image, camX, camY float64, me int) {
	l := g.world.Level
	z := g.zoom()

	// ln draws a world line segment: offset by the camera, scaled by zoom.
	// Line width follows zoom, but a bit weaker so it doesn't get chunky.
	ln := func(x0, y0, x1, y1, w float64, c color.RGBA) {
		line(dst, (x0-camX)*z, (y0-camY)*z, (x1-camX)*z, (y1-camY)*z, w*z*0.75, c)
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
		if g.world.Got[i] || g.race != nil && g.race.Taken[i] >= 0 {
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

	if l.Mode == sim.ModeCargo {
		g.drawCargo(ln)
	}

	if ps := g.players(); ps != nil {
		for i := range ps {
			drawShip(ln, &ps[i].Ship, playerCol[i])
		}
	} else {
		drawShip(ln, &g.world.Ship, colShip)
	}
	if g.duel != nil {
		for _, b := range g.duel.Bullets {
			ln(b.X, b.Y, b.X-b.VX*0.02, b.Y-b.VY*0.02, 2, playerCol[b.Owner])
		}
		// Bombs: a small diamond. Blasts: a ring that grows and fades.
		for _, b := range g.duel.Bombs {
			c := playerCol[b.Owner]
			ln(b.X, b.Y-5, b.X+4, b.Y, 2, c)
			ln(b.X+4, b.Y, b.X, b.Y+5, 2, c)
			ln(b.X, b.Y+5, b.X-4, b.Y, 2, c)
			ln(b.X-4, b.Y, b.X, b.Y-5, 2, c)
		}
		for _, b := range g.duel.Blasts {
			k := b.Age / sim.BlastTime
			r, c := sim.BlastRadius*(0.4+0.6*k), fade(colPad, 1-k)
			const n = 16
			for j := range n {
				a0, a1 := float64(j)*2*math.Pi/n, float64(j+1)*2*math.Pi/n
				ln(b.X+r*math.Cos(a0), b.Y+r*math.Sin(a0), b.X+r*math.Cos(a1), b.Y+r*math.Sin(a1), 3, c)
			}
		}
	}

	// Arrows on a ring around the ship towards targets that are off screen:
	// the nearest crystal or the cargo, and the pad to head for. In a duel:
	// the other player. In a race: the nearest crystal left, then a pad.
	vw, vh := float64(dst.Bounds().Dx())/z, float64(dst.Bounds().Dy())/z
	s := &g.world.Ship
	if me >= 0 {
		s = &g.players()[me].Ship
	}
	arrow := func(tx, ty float64, c color.RGBA) {
		if tx > camX && tx < camX+vw && ty > camY && ty < camY+vh {
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
	switch {
	case g.duel != nil:
		if other := &g.duel.Players[1-me]; s.State != sim.Crashed && other.Alive() {
			arrow(other.Ship.X, other.Ship.Y, playerCol[1-me])
		}
	case g.race != nil:
		if s.State == sim.Crashed || g.race.Winner >= 0 {
			break
		}
		if c, ok := g.race.NearestCrystal(me); ok {
			arrow(c.X, c.Y, colGem)
		} else {
			pad := g.race.Players[me].GoalPad() // the nearest pad: time to land
			arrow((pad.X0+pad.X1)/2, pad.Y, colPad)
		}
	case s.State != sim.Crashed && !g.world.Won:
		if c, ok := g.world.NearestCrystal(); ok {
			arrow(c.X, c.Y, colGem)
		}
		if c := &g.world.Cargo; l.Mode == sim.ModeCargo && !c.Attached && !c.Delivered {
			arrow(c.X, c.Y, colGem)
		}
		pad, col := g.world.GoalPad(), colPad
		if l.Mode == sim.ModeCargo && g.world.Cargo.Attached {
			col = colGem // the goal pad has the cargo's colour
		}
		arrow((pad.X0+pad.X1)/2, pad.Y, col)
	}

	// Particles are tiny, so they skip antialiasing: Ebitengine then draws
	// them all in one batch.
	for _, p := range g.parts {
		dx, dy := math.Cos(p.ang)*p.length/2, math.Sin(p.ang)*p.length/2
		x0, y0 := (p.x-dx-camX)*z, (p.y-dy-camY)*z
		x1, y1 := (p.x+dx-camX)*z, (p.y+dy-camY)*z
		c := fade(p.col, p.life/p.maxLife)
		vector.StrokeLine(dst, float32(x0), float32(y0), float32(x1), float32(y1), float32(1.5*z), c, false)
	}
}

// shipEffects adds exhaust, refuel sparks and wreckage for a ship. prev is
// its state in the previous step, to see when it has just crashed.
func (g *Game) shipEffects(s *sim.Ship, prev *sim.State) {
	if s.Thrusting {
		for range 2 {
			g.parts = append(g.parts, exhaust(s))
		}
	}
	if s.Refuelling && g.tick%3 == 0 {
		g.parts = append(g.parts, refuelSpark(s))
	}
	if s.State == sim.Crashed && *prev != sim.Crashed {
		g.parts = append(g.parts, explosion(s)...)
	}
	*prev = s.State
}

func (g *Game) updateParticles() {
	alive := g.parts[:0]
	for i := range g.parts {
		p := g.parts[i]
		p.update(g.world.Level)
		if p.life > 0 {
			alive = append(alive, p)
		}
	}
	g.parts = alive
}

// drawShip draws a ship that isn't wrecked, with a flame while thrusting.
func drawShip(ln func(x0, y0, x1, y1, w float64, c color.RGBA), s *sim.Ship, col color.RGBA) {
	if s.State == sim.Crashed {
		return
	}
	v := s.Points()
	for i := range v {
		a, b := v[i], v[(i+1)%len(v)]
		ln(a.X, a.Y, b.X, b.Y, 2, col)
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
	ebiten.SetWindowTitle("eldfari")
	log.Printf("eldfari %s", version())
	return ebiten.RunGame(newGame())
}
