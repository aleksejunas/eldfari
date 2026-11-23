# Working notes

Rewrite this file instead of appending. Max ~40 lines.

## Now

- Nothing in progress. Last done: power-ups in the duel (shield, one-shot).
  Duel, race and pads are not run by Claude: need play-testing.

## Next

1. Play-test the duel; tune bullet speed, damage, reload in `sim/duel.go`.
2. Play-test the race (Kappreið, Gullhellir).
3. Play-test the cargo levels: difficulty order, and tune `CargoMass`,
   `RopeLength`, `HookRange` in `sim/cargo.go`.
4. Play-test crystals levels 30–60, especially fuel on Langvegr.
5. Joy-Cons and more controllers: `go run ./cmd/padlisten`, then a line in
   `extraMappings` (game/input.go).

## Decisions

- Cargo: hook within 120 px with a winch, mass 0.4, springy rope (the first
  version was unplayable).
- Duel and race: each player a World sharing the Level, `sim.Player` shared
  (no interface yet). Duel: dying gives the other a point. Race: crystals
  are shared, odd count per level so it can't end in a draw.
- Ships colliding: both crash, nobody scores (ramming must not pay).
- Power-ups: seeded RNG in sim (tests repeatable, game seeds per duel);
  lost on death; no arrow to them, finding one is luck.
  Initials: A–Z and '-' only, to keep the hall of fame tidy.
- Race crash: start again in the air at the last safe place (80 px room,
  120 px below), hanging still until input. No free trip home, no penalty.
- Duel weapons on fingers free while flying: P1 LShift fire, Space bomb;
  P2 RCtrl fire, RAlt bomb. S/↓ shared a finger with thrust.
- Gamepad: RT thrust, A fire, B bomb (user's choice); turning stays digital
  so pad and keyboard are equal. First pad = player 1.
- Pads need 6+ buttons: on Linux a PS4 pad is also its motion sensors and
  touchpad (same name), which gave stuck thrust and turning.
- PDP Switch pad: no input over USB (Bluetooth only); own mapping in
  `game/input.go` (extraMappings). RT uses the pressed state.

## Known problems

- On a cargo win the ship freezes in mid-air under "DELIVERED!".
- The web version lags on power saver: WASM is single-threaded and the
  antialiased lines are the main cost per frame.
