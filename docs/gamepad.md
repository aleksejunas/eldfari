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

Xbox names (PS4: A = ×, B = ○, Y = △, RT = R2, LB/RB = L1/R1). Thrust and
weapons are on different fingers, so you can do both at once.

- Turn: left stick X or d-pad left/right
- Thrust: RT
- Fire: A; bomb: B (duel)
- Menu: d-pad or stick moves, A or Start plays
- In game: Start = menu, Back = restart, LB/RB = previous/next level,
  Y = zoom
- Player 1 has the first pad, player 2 the second; the keyboard still works

## Tasks

### Code (do first, no hardware needed)

- [x] Input in `game/input.go`: `readInput(player)`, keyboard OR pad
- [x] Read pads with `AppendGamepadIDs`, `IsStandardGamepadLayoutAvailable`,
      `IsStandardGamepadButtonPressed`, `StandardGamepadAxisValue`
- [x] Dead zone with hysteresis (on at 0.3, off at 0.2), as constants
- [x] Pure mapping function `padInput` with tests: dead zone, hysteresis,
      d-pad, trigger, fire/bomb
- [x] One pad per player in duel and race; menu with d-pad or stick
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
