# Working notes

Rewrite this file instead of appending. Max ~40 lines.

## Now

- Nothing in progress. Last done: gamepad support (one pad per player).
  Duel, race and pads are not run by Claude: need play-testing.

## Next

1. Play-test the duel; tune bullet speed, damage, reload in `sim/duel.go`.
2. Play-test the race (Kappreið, Gullhellir).
3. Play-test the cargo levels: difficulty order, and tune `CargoMass`,
   `RopeLength`, `HookRange` in `sim/cargo.go`.
4. Play-test crystals levels 30–60, especially fuel on Langvegr.
5. Test the gamepads (PS4, Switch over Bluetooth): stick drift, unplugging
   mid-flight, the menu.

## Decisions

- `sim/` has no Ebitengine import: game logic is testable without a window.
- One folder per mode in `levels/`, no mode field in the JSON.
- Cargo hooks on within 120 px with a winch: 80 px was too fiddly to hook.
- Cargo mass 0.4 and a springy damped rope: 0.8 and a rigid rope were
  unplayable in the first test.
- Old Norse level names; losing best times on a rename is fine.
- Duel and race: each player a World sharing the Level, `sim.Player` shared
  (no interface yet). Duel: dying gives the other a point. Race: crystals
  are shared, odd count per level so it can't end in a draw.
- Duel weapons on fingers free while flying: P1 LShift fire, Space bomb;
  P2 RCtrl fire, RAlt bomb. S/↓ shared a finger with thrust.
- Gamepad: RT thrust, A fire, B bomb (user's choice); turning stays digital
  so pad and keyboard are equal. First pad = player 1.
- Pads need 6+ buttons: on Linux a PS4 pad is also its motion sensors and
  touchpad (same name), which gave stuck thrust and turning.
- Bombs as a second weapon: aiming the nose while flying was too hard.

## Known problems

- On a cargo win the ship freezes in mid-air under "DELIVERED!".
- The web version lags on power saver: WASM is single-threaded and the
  antialiased lines are the main cost per frame.
