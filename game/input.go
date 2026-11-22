package game

import (
	"log"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"eldfari/sim"
)

// Input: the keyboard and gamepads together. Pads use Ebitengine's standard
// layout, so PS4, Switch and Xbox controllers give the same buttons. Names
// here follow Xbox: A is the bottom face button (PS4 ×), B the right one
// (PS4 ○), Y the top one (PS4 △), RT the right trigger (PS4 R2).
//
// Flying: left stick or d-pad turns, RT thrust, A fire, B bomb. Player 1
// has the first pad, player 2 the second; the keyboard works as well.

const (
	stickOn   = 0.3 // the stick starts turning past this
	stickOff  = 0.2 // and stops below this, so it doesn't flicker
	menuStick = 0.5 // the stick moves the menu one step past this

	// minPadButtons tells a real controller from the extra devices some
	// controllers bring. On Linux a PS4 controller also shows up as its
	// motion sensors (0 buttons, an axis that reads as a half-pressed
	// trigger) and its touchpad (4 buttons, a stick stuck at -1), with the
	// same name and standard layout as the controller itself.
	minPadButtons = 6
)

// extraMappings fixes controllers that Ebitengine's database maps wrong
// (SDL gamecontrollerdb format). Only buttons measured on the real controller
// are listed; the rest are left unmapped rather than guessed.
//
// PDP Afterglow Wireless Deluxe (Switch) over Bluetooth ("Lic Pro
// Controller"): without this, ZR/ZL read as RB/LB and + as RT, so + gave
// thrust and ZR did nothing. Over the USB cable it sends no input at all.
// Nintendo's X is the top button (standard y) and Y the left one (x). Not
// measured yet: -, d-pad up/down, the right stick.
const extraMappings = "050000006f0e00008601000001000000,Lic Pro Controller," +
	"a:b0,b:b1,x:b2,y:b3,leftshoulder:b4,rightshoulder:b5,lefttrigger:b6,righttrigger:b7,start:b9," +
	"dpleft:b19,dpright:b17,leftx:a0,lefty:a1,platform:Linux,\n"

func init() {
	if _, err := ebiten.UpdateStandardGamepadLayoutMappings(extraMappings); err != nil {
		log.Printf("gamepad mappings: %v", err)
	}
}

// padState is what one pad does right now.
type padState struct {
	stickX float64 // left stick, -1 left .. 1 right
	// RT pressed. Ebitengine's pressed state, not the 0..1 value: a PDP
	// Switch controller has digital triggers that read 0.5 at rest.
	rt                  bool
	dpadLeft, dpadRight bool
	a, b                bool
}

// stickState remembers whether the stick is turning, for the dead zone with
// hysteresis.
type stickState struct{ left, right bool }

// padInput turns a pad's state into ship input. Turning stays digital, like
// the keyboard, so neither has an advantage.
func padInput(st padState, stick *stickState) sim.Input {
	stick.left = st.stickX < -stickOn || stick.left && st.stickX < -stickOff
	stick.right = st.stickX > stickOn || stick.right && st.stickX > stickOff
	return sim.Input{
		Left:   stick.left || st.dpadLeft,
		Right:  stick.right || st.dpadRight,
		Thrust: st.rt,
		Fire:   st.a,
		Bomb:   st.b,
	}
}

func readPad(id ebiten.GamepadID) padState {
	pressed := func(b ebiten.StandardGamepadButton) bool { return ebiten.IsStandardGamepadButtonPressed(id, b) }
	return padState{
		stickX:    ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal),
		rt:        pressed(ebiten.StandardGamepadButtonFrontBottomRight),
		dpadLeft:  pressed(ebiten.StandardGamepadButtonLeftLeft),
		dpadRight: pressed(ebiten.StandardGamepadButtonLeftRight),
		a:         pressed(ebiten.StandardGamepadButtonRightBottom),
		b:         pressed(ebiten.StandardGamepadButtonRightRight),
	}
}

var warnedPads = map[ebiten.GamepadID]bool{}

// pads returns the connected controllers, in a fixed order: the first is
// player 1's, the second player 2's. Other devices are skipped (and logged
// once).
func pads() []ebiten.GamepadID {
	var out []ebiten.GamepadID
	for _, id := range ebiten.AppendGamepadIDs(nil) {
		if isController(ebiten.IsStandardGamepadLayoutAvailable(id), ebiten.GamepadButtonCount(id)) {
			out = append(out, id)
		} else if !warnedPads[id] {
			warnedPads[id] = true
			log.Printf("ignoring %q: no standard layout or too few buttons to be a controller", ebiten.GamepadName(id))
		}
	}
	slices.Sort(out)
	return out
}

// isController reports whether a gamepad device is a real controller, not
// the motion sensors or touchpad that come along with some.
func isController(standardLayout bool, buttons int) bool {
	return standardLayout && buttons >= minPadButtons
}

// readInput combines the keyboard and a pad for a player: 0 or 1 in the
// two-player modes, -1 in the one-player modes (all keys and the first pad).
func (g *Game) readInput(player int, ids []ebiten.GamepadID) sim.Input {
	key := ebiten.IsKeyPressed
	var in sim.Input
	switch player {
	case 0:
		in = sim.Input{Left: key(ebiten.KeyA), Right: key(ebiten.KeyD), Thrust: key(ebiten.KeyW),
			Fire: key(ebiten.KeyShiftLeft), Bomb: key(ebiten.KeySpace)}
	case 1:
		in = sim.Input{Left: key(ebiten.KeyArrowLeft), Right: key(ebiten.KeyArrowRight), Thrust: key(ebiten.KeyArrowUp),
			Fire: key(ebiten.KeyControlRight), Bomb: key(ebiten.KeyAltRight)}
	default:
		in = sim.Input{
			Left:   key(ebiten.KeyArrowLeft) || key(ebiten.KeyA),
			Right:  key(ebiten.KeyArrowRight) || key(ebiten.KeyD),
			Thrust: key(ebiten.KeyArrowUp) || key(ebiten.KeyW) || key(ebiten.KeySpace),
		}
	}
	i := max(player, 0)
	if i >= len(ids) {
		return in
	}
	p := padInput(readPad(ids[i]), &g.sticks[i])
	return sim.Input{
		Left: in.Left || p.Left, Right: in.Right || p.Right, Thrust: in.Thrust || p.Thrust,
		Fire: in.Fire || p.Fire, Bomb: in.Bomb || p.Bomb,
	}
}

// padJustPressed reports whether any pad just pressed the button.
func padJustPressed(ids []ebiten.GamepadID, b ebiten.StandardGamepadButton) bool {
	for _, id := range ids {
		if inpututil.IsStandardGamepadButtonJustPressed(id, b) {
			return true
		}
	}
	return false
}

// stickSteps turns the left sticks into single steps for the menu: one step
// each time a stick is pushed past half way.
func (g *Game) stickSteps(ids []ebiten.GamepadID) (dx, dy int) {
	var x, y float64
	for _, id := range ids {
		x += ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal)
		y += ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical)
	}
	step := func(v float64) int {
		switch {
		case v > menuStick:
			return 1
		case v < -menuStick:
			return -1
		}
		return 0
	}
	now := [2]int{step(x), step(y)}
	if now[0] != g.menuStick[0] {
		dx = now[0]
	}
	if now[1] != g.menuStick[1] {
		dy = now[1]
	}
	g.menuStick = now
	return dx, dy
}
