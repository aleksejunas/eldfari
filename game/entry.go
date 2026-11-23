package game

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// nameEntry is the initials input for the hall of fame, like on an arcade
// machine: three letters A–Z. Either after a time that makes a level's top
// 10, or for the winner of a duel or race (winMode set).
type nameEntry struct {
	letters [3]byte
	pos     int // the letter being changed
	level   string
	time    float64
	winMode string // "duel" or "race" for a win, "" for a time
	player  int    // the winner, for the title
	age     int    // frames shown; typing is ignored at first (see updateEntry)
}

func newNameEntry(level string, t float64, last string) *nameEntry {
	e := &nameEntry{level: level, time: t, letters: [3]byte{'A', 'A', 'A'}}
	for i := 0; i < len(e.letters) && i < len(last); i++ {
		if c := last[i]; c >= 'A' && c <= 'Z' {
			e.letters[i] = c
		}
	}
	return e
}

func (e *nameEntry) name() string { return string(e.letters[:]) }

// change steps the current letter through A–Z, wrapping around.
func (e *nameEntry) change(d int) {
	e.letters[e.pos] = byte('A' + ((int(e.letters[e.pos]-'A')+d)%26+26)%26)
}

// move goes to the letter left or right, staying within the three.
func (e *nameEntry) move(d int) { e.pos = min(max(e.pos+d, 0), len(e.letters)-1) }

// next confirms the current letter and goes on. It reports true after the
// last one.
func (e *nameEntry) next() bool {
	if e.pos == len(e.letters)-1 {
		return true
	}
	e.pos++
	return false
}

// typed sets the current letter from a typed character and goes on. It
// reports true after the last one. Other characters are ignored.
func (e *nameEntry) typed(r rune) bool {
	if r >= 'a' && r <= 'z' {
		r -= 'a' - 'A'
	}
	if r < 'A' || r > 'Z' {
		return false
	}
	e.letters[e.pos] = byte(r)
	return e.next()
}

// updateEntry handles the initials input. Keyboard: type the letters, or
// up/down to change and left/right to move; Enter saves, Backspace goes
// back. Pads: d-pad or stick to change and move, A next, B back.
func (g *Game) updateEntry() {
	e := g.entry
	ids := pads()
	pad := func(b ebiten.StandardGamepadButton) bool { return padJustPressed(ids, b) }
	key := inpututil.IsKeyJustPressed
	dx, dy := g.stickSteps(ids)
	// The first half second ignores typing: a key held while landing (W)
	// would otherwise repeat into the initials at once.
	if e.age++; e.age < 30 {
		return
	}
	for _, r := range ebiten.AppendInputChars(nil) {
		if e.typed(r) {
			g.saveEntry()
			return
		}
	}
	switch {
	case key(ebiten.KeyArrowUp) || pad(ebiten.StandardGamepadButtonLeftTop) || dy < 0:
		e.change(1)
	case key(ebiten.KeyArrowDown) || pad(ebiten.StandardGamepadButtonLeftBottom) || dy > 0:
		e.change(-1)
	case key(ebiten.KeyArrowLeft) || key(ebiten.KeyBackspace) || pad(ebiten.StandardGamepadButtonLeftLeft) ||
		pad(ebiten.StandardGamepadButtonRightRight) || dx < 0:
		e.move(-1)
	case key(ebiten.KeyArrowRight) || pad(ebiten.StandardGamepadButtonLeftRight) || dx > 0:
		e.move(1)
	case key(ebiten.KeyEnter) || key(ebiten.KeyNumpadEnter):
		g.saveEntry()
	case pad(ebiten.StandardGamepadButtonRightBottom):
		if e.next() {
			g.saveEntry()
		}
	}
}

func (g *Game) saveEntry() {
	if e := g.entry; e.winMode != "" {
		g.lastWins = g.best.addWin(e.winMode, e.name())
	} else {
		g.lastRank = g.best.add(e.level, e.name(), e.time)
	}
	g.entry = nil
}

// drawEntry draws the initials input in a box in the middle of the screen.
func (g *Game) drawEntry(screen *ebiten.Image) {
	e := g.entry
	W, H := float64(g.w)/g.scale, float64(g.h)/g.scale
	small := g.face(fontText, sizeSmall)
	big := g.face(fontNum, 40)
	y := H * 0.42
	help := "type initials  ·  ↑↓ change  ·  ← → move  ·  Enter save        pad: ↑↓ ← →, A next, B back"
	g.badge(screen, W/2, y-12, g.textWidth(help, small)+40, 130)
	title := fmt.Sprintf("HALL OF FAME!  %s", fmtTime(e.time))
	if e.winMode != "" {
		title = fmt.Sprintf("PLAYER %d WINS  ·  HALL OF FAME", e.player+1)
	}
	g.print(screen, title, W/2, y, g.face(fontTitle, sizeText), text.AlignCenter, colGem)
	for i, c := range e.letters {
		x := W/2 + float64(i-1)*40
		col := colDim
		if i == e.pos {
			col = colShip
			g.rect(screen, x-14, y+76, 28, 3, colGem)
		}
		g.print(screen, string(c), x, y+26, big, text.AlignCenter, col)
	}
	g.print(screen, help, W/2, y+92, small, text.AlignCenter, colDim)
}
