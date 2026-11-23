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
// arrow keys. In a duel the weapons are on fingers that are free while
// flying: player 1 fires with left Shift (little finger) and bombs with Space
// (thumb), player 2 fires with right Ctrl and bombs with right Alt (thumb).

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

func (g *Game) updateTwoPlayer(ids []ebiten.GamepadID) {
	in := [2]sim.Input{g.readInput(0, ids), g.readInput(1, ids)}
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

// drawPlayerHUD draws player i's fuel (and health in a duel) and keys in
// its half: player 1 at the left edge, player 2 mirrored at the right edge,
// so neither runs into the score in the middle. Also a countdown while the
// player waits to start again.
func (g *Game) drawPlayerHUD(dst *ebiten.Image, i int) {
	p := &g.players()[i]
	W, H := float64(dst.Bounds().Dx())/g.scale, float64(dst.Bounds().Dy())/g.scale
	small := g.face(fontText, sizeSmall)
	labelW := g.textWidth("HEALTH  ", small) // same for both labels, so the bars line up
	x := margin
	if i == 1 {
		x = W - margin - labelW - fuelBarW
	}
	bar := func(y float64, label string, v float64, col color.RGBA) {
		g.print(dst, label, x, y, small, text.AlignStart, colDim)
		g.rect(dst, x+labelW, y+4, fuelBarW*v, fuelBarH, col)
		g.strokeRect(dst, x+labelW, y+4, fuelBarW, fuelBarH, colDim)
	}
	bar(margin, "FUEL", p.Ship.Fuel, colPad)
	if g.duel != nil {
		col := playerCol[i]
		if p.Health <= sim.BulletDamage {
			col = colTerrain
		}
		bar(margin+20, "HEALTH", p.Health, col)
	}

	keys := [2]string{"A/D turn  ·  W thrust", "← → turn  ·  ↑ thrust"}
	if g.duel != nil {
		keys = [2]string{"A/D turn  ·  W thrust  ·  Shift fire  ·  Space bomb", "← → turn  ·  ↑ thrust  ·  Ctrl fire  ·  Alt bomb"}
	}
	if i == 0 {
		g.print(dst, keys[0], margin, H-margin-sizeSmall, small, text.AlignStart, colDim)
	} else {
		g.print(dst, keys[1], W-margin, H-margin-sizeSmall, small, text.AlignEnd, colDim)
	}

	if p.Waiting() {
		g.print(dst, "thrust or turn to go on", W/2, H*0.3, g.face(fontTitle, sizeText), text.AlignCenter, colDim)
	}
	if !p.Alive() && g.winner() < 0 {
		g.print(dst, fmt.Sprintf("back in %.0f", p.Respawn+0.5), W/2, H*0.3, g.face(fontTitle, sizeText), text.AlignCenter, colDim)
	}
}

// textWidth is the width of s in window pixels (like the HUD positions).
func (g *Game) textWidth(s string, face *text.GoTextFace) float64 {
	return text.AdvanceAt(s, len(s), face) / g.scale
}

// badge draws a box in the background colour with a thin frame, centred on
// x, so text on top of the divider between the halves stays readable.
func (g *Game) badge(dst *ebiten.Image, x, y, w, h float64) {
	g.rect(dst, x-w/2, y, w, h, colBG)
	g.strokeRect(dst, x-w/2, y, w, h, colDim)
}

// winner is the duel's or race's Winner: -1 while it goes on.
func (g *Game) winner() int {
	if g.duel != nil {
		return g.duel.Winner
	}
	return g.race.Winner
}

// drawTwoPlayerHUD draws what belongs to both players in the middle, on
// boxes over the divider: the score and level at the top, the winner or a
// message, and the common keys at the bottom.
func (g *Game) drawTwoPlayerHUD(screen *ebiten.Image) {
	W, H := float64(g.w)/g.scale, float64(g.h)/g.scale
	ps := g.players()
	bigNum := g.face(fontNum, sizeBig)
	small := g.face(fontText, sizeSmall)
	normal := g.face(fontText, sizeText)

	label := fmt.Sprintf("%s  ·  FIRST TO %d", g.world.Level.Name, sim.DuelScore)
	if g.race != nil {
		label = fmt.Sprintf("%s  ·  %d CRYSTALS LEFT", g.world.Level.Name, g.race.Left())
	}
	g.badge(screen, W/2, margin-10, g.textWidth(label, small)+32, 66)
	g.print(screen, fmt.Sprint(ps[0].Score), W/2-14, margin-6, bigNum, text.AlignEnd, playerCol[0])
	g.print(screen, ":", W/2, margin-6, bigNum, text.AlignCenter, colDim)
	g.print(screen, fmt.Sprint(ps[1].Score), W/2+14, margin-6, bigNum, text.AlignStart, playerCol[1])
	g.print(screen, label, W/2, margin+32, small, text.AlignCenter, colDim)

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
		big := g.face(fontTitle, sizeBig)
		w := max(g.textWidth(msg, big), g.textWidth(sub, normal)) + 48
		g.badge(screen, W/2, H*0.22-14, w, 82)
		g.print(screen, msg, W/2, H*0.22, big, text.AlignCenter, col)
		g.print(screen, sub, W/2, H*0.22+34, normal, text.AlignCenter, colDim)
	}

	common := "R restart  ·  M menu"
	g.badge(screen, W/2, H-margin-sizeSmall-8, g.textWidth(common, small)+28, sizeSmall+16)
	g.print(screen, common, W/2, H-margin-sizeSmall, small, text.AlignCenter, colDim)
}
