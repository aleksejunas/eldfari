package game

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"eldfari/sim"
)

// modeOrder is the order of the modes in the menu.
var modeOrder = []sim.Mode{sim.ModeCrystals, sim.ModeCargo}

// modeHelp is a short description of each mode, in lines that fit the
// menu's left column.
var modeHelp = map[sim.Mode][]string{
	sim.ModeCrystals: {"Collect every crystal,", "then land on the start pad."},
	sim.ModeCargo:    {"Hook the cargo and lower it", "onto the marked pad."},
}

// menu is the mode and level chooser: which mode is selected, and which
// level in each mode.
type menu struct {
	mode  int         // index in modeOrder
	level map[int]int // selected level per mode
}

// move changes the selected mode by dMode and the level by dLevel. Both wrap
// around. counts is the number of levels in each mode (by index in
// modeOrder).
func (m *menu) move(dMode, dLevel int, counts []int) {
	if m.level == nil {
		m.level = map[int]int{}
	}
	n := len(counts)
	m.mode = ((m.mode+dMode)%n + n) % n
	c := counts[m.mode]
	m.level[m.mode] = ((m.level[m.mode]+dLevel)%c + c) % c
}

// selected returns the selected mode and level index.
func (m *menu) selected() (sim.Mode, int) { return modeOrder[m.mode], m.level[m.mode] }

// openMenu shows the menu with the current mode and level selected.
func (g *Game) openMenu() {
	for i, mode := range modeOrder {
		if mode == g.mode {
			g.menu.mode = i
		}
	}
	g.menu.move(0, 0, g.levelCounts())
	g.menu.level[g.menu.mode] = g.lvl
	g.inMenu = true
}

func (g *Game) levelCounts() []int {
	counts := make([]int, len(modeOrder))
	for i, mode := range modeOrder {
		counts[i] = len(g.levels[mode])
	}
	return counts
}

func (g *Game) updateMenu() {
	pressed := func(keys ...ebiten.Key) bool {
		for _, k := range keys {
			if inpututil.IsKeyJustPressed(k) {
				return true
			}
		}
		return false
	}
	counts := g.levelCounts()
	switch {
	case pressed(ebiten.KeyArrowLeft, ebiten.KeyA):
		g.menu.move(-1, 0, counts)
	case pressed(ebiten.KeyArrowRight, ebiten.KeyD):
		g.menu.move(1, 0, counts)
	case pressed(ebiten.KeyArrowUp, ebiten.KeyW):
		g.menu.move(0, -1, counts)
	case pressed(ebiten.KeyArrowDown, ebiten.KeyS):
		g.menu.move(0, 1, counts)
	case pressed(ebiten.KeyEnter, ebiten.KeyNumpadEnter, ebiten.KeySpace):
		g.mode, g.lvl = g.menu.selected()
		g.setLevel(g.lvl)
		g.inMenu = false
	}
}

// drawMenu draws the title, the modes on the left and the levels of the
// selected mode on the right, with best times.
func (g *Game) drawMenu(screen *ebiten.Image) {
	screen.Fill(colBG)
	W, H := float64(g.w)/g.scale, float64(g.h)/g.scale
	small := g.face(fontText, sizeSmall)
	normal := g.face(fontText, sizeText)
	title := g.face(fontTitle, sizeText)
	big := g.face(fontTitle, 40)
	num := g.face(fontNum, sizeText)

	left := W/2 - 300
	g.print(screen, "eldfari", left, H*0.14, big, text.AlignStart, colShip)

	selMode, selLevel := g.menu.selected()
	top := H*0.14 + 80

	// Modes.
	g.print(screen, "MODE", left, top, small, text.AlignStart, colDim)
	for i, mode := range modeOrder {
		y := top + 28 + float64(i)*30
		col, face := colDim, normal
		if mode == selMode {
			col, face = colGem, title
			g.print(screen, "›", left-18, y, face, text.AlignStart, col)
		}
		g.print(screen, strings.ToUpper(modeNames[mode]), left, y, face, text.AlignStart, col)
	}
	for i, line := range modeHelp[selMode] {
		y := top + 28 + float64(len(modeOrder))*30 + 16 + float64(i)*18
		g.print(screen, line, left, y, small, text.AlignStart, colDim)
	}

	// Levels of the selected mode.
	lx := left + 260
	g.print(screen, "LEVEL", lx, top, small, text.AlignStart, colDim)
	g.print(screen, "BEST", lx+300, top, small, text.AlignEnd, colDim)
	for i, l := range g.levels[selMode] {
		y := top + 28 + float64(i)*26
		col := colShip
		if i == selLevel {
			col = colGem
			g.print(screen, "›", lx-18, y, title, text.AlignStart, col)
		}
		g.print(screen, fmt.Sprintf("%d  %s", i+1, l.Name), lx, y, normal, text.AlignStart, col)
		if b := g.best.get(modeNames[selMode] + "/" + l.Name); b > 0 {
			g.print(screen, fmtTime(b), lx+300, y, num, text.AlignEnd, col)
		}
	}

	help := "← → mode  ·  ↑ ↓ level  ·  Enter play  ·  M or Esc: back to this menu"
	g.print(screen, help, W/2, H-margin-sizeSmall, small, text.AlignCenter, colDim)
}
