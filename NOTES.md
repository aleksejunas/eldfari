# Working notes

Rewrite this file instead of appending. Max ~40 lines.

## Now

- Nothing in progress. Last done: cargo mode has seven levels (40–70 new).

## Next

1. Play-test the cargo levels: difficulty order, and tune `CargoMass`,
   `RopeLength`, `HookRange` in `sim/cargo.go`.
2. Play-test crystals levels 30–60, especially fuel on Langvegr.
3. `cmd/levelstat` for cargo levels (gap with the cargo hanging, route).
4. Gamepad (`docs/gamepad.md`).
5. Two players in split screen.

## Decisions

- `sim/` has no Ebitengine import: game logic is testable without a window.
- One folder per mode in `levels/`, no mode field in the JSON.
- Cargo hooks on within 120 px with a winch: 80 px was too fiddly to hook.
- Cargo mass 0.4 and a springy damped rope: 0.8 and a rigid rope were
  unplayable in the first test.
- An autopilot test flies every cargo level, so each level is proven doable.
- Old Norse level names; losing best times on a rename is fine.
- Commit and push after each green step without asking.

## Known problems

- On a cargo win the ship freezes in mid-air under "DELIVERED!".
- The web version lags on power saver: WASM is single-threaded and the
  antialiased lines are the main cost per frame.
- `cmd/levelstat` only measures crystals levels.
