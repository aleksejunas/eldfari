# Gamepad support

Goal: play with a controller on desktop, Steam Deck and in the browser (WASM).
Keyboard keeps working exactly as before. No touch support: see README.

## Decisions

- Use Ebitengine's standard gamepad layout, not raw button numbers.
  Controllers without a standard layout are ignored (log once).
- Rotation stays digital, like the keyboard. `sim.Input` is `Left/Right/Thrust`
  and balance must not differ between keyboard and pad. Analog rotation would
  be a physics change in `sim`, so it is a separate task, later.

## Mapping

| Action         | Gamepad                    |
|----------------|----------------------------|
| Rotate         | Left stick X, d-pad L/R    |
| Thrust         | A, right trigger           |
| Restart        | Start                      |
| Next/prev level| Right/left shoulder        |
| Zoom           | Y                          |
| Fullscreen     | Back/Select                |

## Tasks

### Code (do first, no hardware needed)

- [ ] Move input out of `Update()` into `input.go`: `readInput() sim.Input`
- [ ] Read pads with `AppendGamepadIDs`, `IsStandardGamepadLayoutAvailable`,
      `IsStandardGamepadButtonPressed`, `StandardGamepadAxisValue`
- [ ] Dead zone with hysteresis (on at ~0.3, off at ~0.2), as constants
- [ ] Pure mapping function (buttons + axes -> `sim.Input`) with tests:
      dead zone, d-pad vs stick, both at once
- [ ] Hot-plug: pad disconnected mid-flight gives no input, no crash
- [ ] HUD help text follows the last used device (keyboard or pad)

### Hardware (needs a controller)

- [ ] Check stick drift does not cause ghost rotation
- [ ] Steam Deck

### WASM

- [ ] "Press a button to start" (browser hides the pad until a button press)
- [ ] Test in Chrome and Firefox
- [ ] Mobile browser with Bluetooth pad (only after desktop works)

### Later

- [ ] `readInput(player int)` so pads can be assigned per player in 2-player
      mode (structure only, no 2-player yet)
- [ ] Optional: analog rotation (needs its own design and tests in `sim`)

## Done when

- Keyboard behaviour unchanged, tests green
- Playable start to finish with only a controller
- Mapping documented in README and `CLAUDE.md`
