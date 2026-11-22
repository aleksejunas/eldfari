package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"eldfari/sim"
)

// Two-player modes (duel and race) in the game: two players on one keyboard,
// split screen. Player 1 (left) flies with A/D/W, player 2 (right) with the
// arrow keys. In a duel, S and arrow down fire.

var (
	colP2     = color.RGBA{0x8a, 0xc6, 0xd0, 0xff}
	playerCol = [2]color.RGBA{colShip, colP2}
)

// duelView is each player's half of the screen: a camera and an image to
// draw into.
type duelView struct {
	camX, camY float64
	img        *ebiten.Image
	prev       sim.State
}

// players returns the two players in a duel or a race, or nil in the
// one-player modes.
func (g *Game) players() *[2]sim.Player {
	switch {
	case g.duel != nil:
		return &g.duel.Players
	case g.race != nil:
		return &g.race.Players
	}
	return nil
}

func (g *Game) resetTwoPlayer() {
	if g.duel != nil {
		g.duel.Reset()
	} else {
		g.race.Reset()
	}
	for i := range g.views {
		g.views[i].prev = sim.Landed
		g.views[i].camX, g.views[i].camY = g.playerCamTarget(i)
	}
}

func (g *Game) updateTwoPlayer() {
	key := ebiten.IsKeyPressed
	in := [2]sim.Input{
		{Left: key(ebiten.KeyA), Right: key(ebiten.KeyD), Thrust: key(ebiten.KeyW), Fire: key(ebiten.KeyS)},
		{Left: key(ebiten.KeyArrowLeft), Right: key(ebiten.KeyArrowRight), Thrust: key(ebiten.KeyArrowUp), Fire: key(ebiten.KeyArrowDown)},
	}
	if g.duel != nil {
		g.duel.Step(in)
	} else {
		g.race.Step(in)
	}
	ps := g.players()
	for i := range ps {
		v := &g.views[i]
		g.shipEffects(&ps[i].Ship, &v.prev)
		tx, ty := g.playerCamTarget(i)
		v.camX += (tx - v.camX) * 0.12
		v.camY += (ty - v.camY) * 0.12
	}
	g.updateParticles()
}

// playerCamTarget centres player i's ship in its half of the screen.
func (g *Game) playerCamTarget(i int) (float64, float64) {
	s, l, z := &g.players()[i].Ship, g.world.Level, g.zoom()
	vw, vh := float64(g.w/2)/z, float64(g.h)/z
	return clampCam(s.X-vw/2, vw, float64(l.W)), clampCam(s.Y-vh/2, vh, float64(l.H))
}

// drawTwoPlayer draws each player's view into its half of the screen, with
// its own fuel (and health in a duel), then the score and messages across
// both.
func (g *Game) drawTwoPlayer(screen *ebiten.Image) {
	screen.Fill(colBG)
	half := g.w / 2
	for i := range g.views {
		v := &g.views[i]
		if v.img == nil || v.img.Bounds().Dx() != half || v.img.Bounds().Dy() != g.h {
			if v.img != nil {
				v.img.Deallocate()
			}
			v.img = ebiten.NewImage(half, g.h)
		}
		v.img.Fill(colBG)
		g.drawScene(v.img, v.camX, v.camY, i)
		g.drawPlayerHUD(v.img, i)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(i*half), 0)
		screen.DrawImage(v.img, op)
	}
	line(screen, float64(half), 0, float64(half), float64(g.h), 2*g.scale, colDim)
	g.drawTwoPlayerHUD(screen)
}

// drawPlayerHUD draws player i's fuel (and health in a duel) in the top left
// of its half, and a countdown while it waits to start again.
func (g *Game) drawPlayerHUD(dst *ebiten.Image, i int) {
	p := &g.players()[i]
	small := g.face(fontText, sizeSmall)
	x := margin + g.print(dst, "FUEL   ", margin, margin, small, text.AlignStart, colDim)
	g.rect(dst, x, margin+4, fuelBarW*p.Ship.Fuel, fuelBarH, colPad)
	g.strokeRect(dst, x, margin+4, fuelBarW, fuelBarH, colDim)
	if g.duel != nil {
		y := margin + 20
		g.print(dst, "HEALTH ", margin, y, small, text.AlignStart, colDim)
		healthCol := playerCol[i]
		if p.Health <= sim.BulletDamage {
			healthCol = colTerrain
		}
		g.rect(dst, x, y+4, fuelBarW*p.Health, fuelBarH, healthCol)
		g.strokeRect(dst, x, y+4, fuelBarW, fuelBarH, colDim)
	}
	if !p.Alive() && g.winner() < 0 {
		W, H := float64(dst.Bounds().Dx())/g.scale, float64(dst.Bounds().Dy())/g.scale
		g.print(dst, fmt.Sprintf("back in %.0f", p.Respawn+0.5), W/2, H*0.3, g.face(fontTitle, sizeText), text.AlignCenter, colDim)
	}
}

// winner is the duel's or race's Winner: -1 while it goes on.
func (g *Game) winner() int {
	if g.duel != nil {
		return g.duel.Winner
	}
	return g.race.Winner
}

// drawTwoPlayerHUD draws the score in the middle at the top, the winner and
// the keys.
func (g *Game) drawTwoPlayerHUD(screen *ebiten.Image) {
	W, H := float64(g.w)/g.scale, float64(g.h)/g.scale
	ps := g.players()
	bigNum := g.face(fontNum, sizeBig)
	normal := g.face(fontText, sizeText)
	g.print(screen, fmt.Sprint(ps[0].Score), W/2-14, margin-6, bigNum, text.AlignEnd, playerCol[0])
	g.print(screen, fmt.Sprint(ps[1].Score), W/2+14, margin-6, bigNum, text.AlignStart, playerCol[1])

	label := fmt.Sprintf("%s  ·  FIRST TO %d", g.world.Level.Name, sim.DuelScore)
	help := "P1: A/D turn · W thrust · S fire        P2: ← → turn · ↑ thrust · ↓ fire        R restart · M menu"
	if g.race != nil {
		label = fmt.Sprintf("%s  ·  %d CRYSTALS LEFT", g.world.Level.Name, g.race.Left())
		help = "P1: A/D turn · W thrust        P2: ← → turn · ↑ thrust        R restart · M menu"
	}
	g.print(screen, label, W/2, margin+30, g.face(fontText, sizeSmall), text.AlignCenter, colDim)

	msg, sub, col := "", "", colShip
	switch w := g.winner(); {
	case w == sim.RaceDraw:
		msg, sub = "DRAW!", "R rematch  ·  N next level  ·  M menu"
	case w >= 0:
		msg, sub, col = fmt.Sprintf("PLAYER %d WINS!", w+1), "R rematch  ·  N next level  ·  M menu", playerCol[w]
	case g.race != nil && g.race.Left() == 0:
		msg, sub, col = "ALL CRYSTALS TAKEN", "land on a pad", colGem
	}
	if msg != "" {
		g.print(screen, msg, W/2, H*0.22, g.face(fontTitle, sizeBig), text.AlignCenter, col)
		g.print(screen, sub, W/2, H*0.22+34, normal, text.AlignCenter, colDim)
	}
	g.print(screen, help, W/2, H-margin-sizeSmall, g.face(fontText, sizeSmall), text.AlignCenter, colDim)
}
