package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"eldfari/sim"
)

// Duel mode in the game: two players on one keyboard, split screen. Player 1
// (left) flies with A/D/W and fires with S, player 2 (right) with the arrow
// keys and fires with arrow down.

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

func (g *Game) resetDuel() {
	g.duel.Reset()
	for i := range g.views {
		g.views[i].prev = sim.Landed
		g.views[i].camX, g.views[i].camY = g.duelCamTarget(i)
	}
}

func (g *Game) updateDuel() {
	key := ebiten.IsKeyPressed
	g.duel.Step([2]sim.Input{
		{Left: key(ebiten.KeyA), Right: key(ebiten.KeyD), Thrust: key(ebiten.KeyW), Fire: key(ebiten.KeyS)},
		{Left: key(ebiten.KeyArrowLeft), Right: key(ebiten.KeyArrowRight), Thrust: key(ebiten.KeyArrowUp), Fire: key(ebiten.KeyArrowDown)},
	})
	for i := range g.duel.Players {
		v := &g.views[i]
		g.shipEffects(&g.duel.Players[i].Ship, &v.prev)
		tx, ty := g.duelCamTarget(i)
		v.camX += (tx - v.camX) * 0.12
		v.camY += (ty - v.camY) * 0.12
	}
	g.updateParticles()
}

// duelCamTarget centres player i's ship in its half of the screen.
func (g *Game) duelCamTarget(i int) (float64, float64) {
	s, l, z := &g.duel.Players[i].Ship, g.duel.Level, g.zoom()
	vw, vh := float64(g.w/2)/z, float64(g.h)/z
	return clampCam(s.X-vw/2, vw, float64(l.W)), clampCam(s.Y-vh/2, vh, float64(l.H))
}

// drawDuel draws each player's view into its half of the screen, with its
// own fuel and health, then the score and messages across both.
func (g *Game) drawDuel(screen *ebiten.Image) {
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
	g.drawDuelHUD(screen)
}

// drawPlayerHUD draws player i's fuel and health in the top left of its half.
func (g *Game) drawPlayerHUD(dst *ebiten.Image, i int) {
	p := &g.duel.Players[i]
	small := g.face(fontText, sizeSmall)
	x := margin + g.print(dst, "FUEL   ", margin, margin, small, text.AlignStart, colDim)
	g.rect(dst, x, margin+4, fuelBarW*p.Ship.Fuel, fuelBarH, colPad)
	g.strokeRect(dst, x, margin+4, fuelBarW, fuelBarH, colDim)
	y := margin + 20
	g.print(dst, "HEALTH ", margin, y, small, text.AlignStart, colDim)
	healthCol := playerCol[i]
	if p.Health <= sim.BulletDamage {
		healthCol = colTerrain
	}
	g.rect(dst, x, y+4, fuelBarW*p.Health, fuelBarH, healthCol)
	g.strokeRect(dst, x, y+4, fuelBarW, fuelBarH, colDim)
	if !p.Alive() && g.duel.Winner < 0 {
		W, H := float64(dst.Bounds().Dx())/g.scale, float64(dst.Bounds().Dy())/g.scale
		g.print(dst, fmt.Sprintf("back in %.0f", p.Respawn+0.5), W/2, H*0.3, g.face(fontTitle, sizeText), text.AlignCenter, colDim)
	}
}

// drawDuelHUD draws the score in the middle at the top, the winner and the
// keys.
func (g *Game) drawDuelHUD(screen *ebiten.Image) {
	W, H := float64(g.w)/g.scale, float64(g.h)/g.scale
	bigNum := g.face(fontNum, sizeBig)
	g.print(screen, fmt.Sprint(g.duel.Players[0].Score), W/2-14, margin-6, bigNum, text.AlignEnd, playerCol[0])
	g.print(screen, fmt.Sprint(g.duel.Players[1].Score), W/2+14, margin-6, bigNum, text.AlignStart, playerCol[1])
	label := fmt.Sprintf("%s  ·  FIRST TO %d", g.duel.Level.Name, sim.DuelScore)
	g.print(screen, label, W/2, margin+30, g.face(fontText, sizeSmall), text.AlignCenter, colDim)

	if w := g.duel.Winner; w >= 0 {
		g.print(screen, fmt.Sprintf("PLAYER %d WINS!", w+1), W/2, H*0.22, g.face(fontTitle, sizeBig), text.AlignCenter, playerCol[w])
		g.print(screen, "R rematch  ·  N next level  ·  M menu", W/2, H*0.22+34, g.face(fontText, sizeText), text.AlignCenter, colDim)
	}
	help := "P1: A/D turn · W thrust · S fire        P2: ← → turn · ↑ thrust · ↓ fire        R restart · M menu"
	g.print(screen, help, W/2, H-margin-sizeSmall, g.face(fontText, sizeSmall), text.AlignCenter, colDim)
}
