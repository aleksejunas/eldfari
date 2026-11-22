// Command padlisten shows what a game controller sends, raw and in
// Ebitengine's standard layout, to add a mapping for a new controller to
// extraMappings in game/input.go. It runs for 30 seconds; press one button at
// a time and move each stick.
//
//	go run ./cmd/padlisten
package main

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

var names = map[ebiten.StandardGamepadButton]string{
	ebiten.StandardGamepadButtonRightBottom: "A(bottom)", ebiten.StandardGamepadButtonRightRight: "B(right)",
	ebiten.StandardGamepadButtonRightLeft: "X(left)", ebiten.StandardGamepadButtonRightTop: "Y(top)",
	ebiten.StandardGamepadButtonFrontTopLeft: "LB", ebiten.StandardGamepadButtonFrontTopRight: "RB",
	ebiten.StandardGamepadButtonFrontBottomLeft: "LT", ebiten.StandardGamepadButtonFrontBottomRight: "RT",
	ebiten.StandardGamepadButtonCenterLeft: "Back", ebiten.StandardGamepadButtonCenterRight: "Start",
	ebiten.StandardGamepadButtonLeftTop: "dpad-up", ebiten.StandardGamepadButtonLeftBottom: "dpad-down",
	ebiten.StandardGamepadButtonLeftLeft: "dpad-left", ebiten.StandardGamepadButtonLeftRight: "dpad-right",
}

type listen struct {
	start time.Time
	last  map[ebiten.GamepadID]string
}

func (l *listen) Update() error {
	if l.start.IsZero() {
		l.start = time.Now()
		l.last = map[ebiten.GamepadID]string{}
		for _, id := range ebiten.AppendGamepadIDs(nil) {
			fmt.Printf("pad %d: %q sdl=%s standard=%v buttons=%d axes=%d\n", id, ebiten.GamepadName(id),
				ebiten.GamepadSDLID(id), ebiten.IsStandardGamepadLayoutAvailable(id),
				ebiten.GamepadButtonCount(id), ebiten.GamepadAxisCount(id))
		}
		fmt.Println("Press one button at a time and move each stick.")
	}
	for _, id := range ebiten.AppendGamepadIDs(nil) {
		s := fmt.Sprintf("pad %d raw buttons:", id)
		for b := range ebiten.GamepadButtonCount(id) {
			if ebiten.IsGamepadButtonPressed(id, ebiten.GamepadButton(b)) {
				s += fmt.Sprintf(" %d", b)
			}
		}
		s += "  raw axes:"
		for a := range ebiten.GamepadAxisCount(id) {
			if v := ebiten.GamepadAxisValue(id, ebiten.GamepadAxisType(a)); v > 0.5 || v < -0.5 {
				s += fmt.Sprintf(" %d=%.1f", a, v)
			}
		}
		s += "  standard:"
		for b := range ebiten.StandardGamepadButtonMax + 1 {
			if n, ok := names[b]; ok && ebiten.IsStandardGamepadButtonPressed(id, b) {
				s += " " + n
			}
		}
		if s != l.last[id] {
			fmt.Println(s)
			l.last[id] = s
		}
	}
	if time.Since(l.start) > 30*time.Second {
		return ebiten.Termination
	}
	return nil
}

func (l *listen) Draw(*ebiten.Image)         {}
func (l *listen) Layout(w, h int) (int, int) { return 320, 240 }

func main() {
	ebiten.SetWindowSize(320, 240)
	ebiten.SetWindowTitle("padlisten")
	if err := ebiten.RunGame(&listen{}); err != nil {
		panic(err)
	}
}
