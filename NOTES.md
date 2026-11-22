# Working notes

Rewrite this file instead of appending. Max ~40 lines.

## Now

- Nothing in progress. Last done: race mode (two players, shared crystals,
  most wins). Duel and race are not run by Claude: need play-testing.

## Next

1. Play-test the duel; tune bullet speed, damage, reload in `sim/duel.go`.
2. Play-test the race (Kappreið, Gullhellir).
3. Play-test the cargo levels: difficulty order, and tune `CargoMass`,
   `RopeLength`, `HookRange` in `sim/cargo.go`.
4. Play-test crystals levels 30–60, especially fuel on Langvegr.
5. Gamepad (`docs/gamepad.md`).

## Decisions

- `sim/` has no Ebitengine import: game logic is testable without a window.
- One folder per mode in `levels/`, no mode field in the JSON.
- Cargo hooks on within 120 px with a winch: 80 px was too fiddly to hook.
- Cargo mass 0.4 and a springy damped rope: 0.8 and a rigid rope were
  unplayable in the first test.
- An autopilot test flies every cargo level, so each level is proven doable.
- Old Norse level names; losing best times on a rename is fine.
- Commit and push after each green step without asking.
- Duel and race: each player a World sharing the Level, `sim.Player` shared
  (no interface yet). Duel: dying gives the other a point. Race: crystals
  are shared, odd count per level so it can't end in a draw.

## Known problems

- On a cargo win the ship freezes in mid-air under "DELIVERED!".
- The web version lags on power saver: WASM is single-threaded and the
  antialiased lines are the main cost per frame.
