# Raketti – todo

The sections are in the order they will be done.

## 1. Crystals (done)

- [x] Crystals in the level (Level.Crystals), picked up by flying through them
- [x] Counter and timer in the HUD (timer starts at first liftoff)
- [x] Win by landing on the start pad with all crystals taken
- [x] Best time, saved to file (`game/besttimes.go`)
- [x] Tests: pickup, win, timer, free air around each crystal

## 2. Share the game: playable in the browser (done)

I want to write this code myself. Claude explains and gives code to copy,
but does not edit files.

- [x] WASM build (`GOOS=js GOARCH=wasm`) + `wasm_exec.js` + a small
      `index.html`
- [x] Host on GitHub Pages (static, a GitHub Actions workflow that builds
      and deploys)
- [x] Best times in the browser (localStorage, `game/besttimes_js.go`)
- [x] Short instructions for a non-developer: link, keys, what the goal is

## 3. Levels: one folder per mode (done)

Mode = folder: `levels/crystals/`, `levels/cargo/`, `levels/twoplayer/`.
No "mode" field in the JSON. Each mode has its own loader and its own fields,
and `DisallowUnknownFields` stays strict, so data for one mode is an error in
another mode's levels.

- [x] Move today's levels: `01-grotta` -> `crystals/10-grotta`,
      `02-passasjen` -> `crystals/20-passasjen`,
      `03-dypet` -> `crystals/70-dypet`
- [x] `//go:embed crystals/*.json`. A mode's folder and embed pattern are
      added when that mode is built (`go:embed` fails on an empty folder)
- [x] File names start with a number with gaps (10, 20, 30 ...), so new
      levels can be slotted in without renaming. Sort on the number, not as
      text, so 100 comes after 90
- [x] N/P only steps through the levels of the current mode
- [x] Best times keyed by mode + level name, so equal names in two modes
      don't share a record
- [x] Update paths in README and CLAUDE.md

## 4. New crystals levels

The ship is about 21 px tall. One new difficulty per level, and the narrowest
gap goes down from ~7x ship height (20) to ~4x (70).

- [x] Analysis first: `go run ./cmd/levelstat` prints the narrowest gap the
      route must pass and the route length per level (`-png dir` draws them)
- [x] Run it on 10, 20 and 70: Passasjen is 144 px at its narrowest (under
      the pillar), not ~180, so the ramp below starts from there
- [x] Ramp (narrowest gap, route length):
  - 10 Grotta: open, 272 px, 6.6k px
  - 20 Passasjen: 144 px (6.9x), 6.9k px
  - [x] 30 Sprekken: one 130 px crack between two halls, five pads:
        128 px (6.1x), 8.1k px
  - [x] 40 Dryppet: short low passage (250 px long) + stalactites:
        112 px (5.3x), 8.1k px
  - [x] 50 Langveien: long snake, refuel halfway: 104 px (5.0x), 14.2k px
  - [x] 60 Sjakta: one wide sloping shaft (168 px) as practice, then a low
        passage: 96 px (4.6x), 8.5k px
  - 70 Dypet: narrow passages, sloping tunnel, long route: 79 px (3.7x),
    8.3k px
- [ ] Every new level is play-tested by someone else (via the web version).
      Check especially fuel on Langveien

## 5. Gamepad support

Full plan: [gamepad.md](gamepad.md)

- [ ] Keyboard and gamepad input merged in one `readInput()`
- [ ] Tested on real hardware (xbox/playstation pad, Steam Deck)

## 6. Cargo on a rope (Thrust style)

1. [ ] Substeps in the physics (a rigid rope gets unstable at 60 Hz and can
       tunnel through walls)
2. [ ] Pendulum without mass and without collision: the rope is a hard
       distance constraint (rigid first, maybe spring later), with a test
       that the distance never exceeds L
3. [ ] Cargo mass: the cargo is pulled by gravity and pulls on the ship, and
       heavy cargo = slower acceleration
4. [ ] Collision for the cargo against terrain
5. [ ] Pick-up and delivery: the cargo rests on the ground and is attached
       automatically when the ship is within L (60–100 px); it must be
       delivered to the goal pad
6. [ ] Cargo mode: `levels/cargo/` with its own loader (cargo start
       position, goal pad), and a mode enum in `sim` with a `switch` (no
       interface until two players show what is shared)
7. [ ] Cargo levels designed for the rope: no narrow sloping passages,
       playable with a pendulum hanging below the ship
8. [ ] Temporary key to switch mode until the mode chooser exists

## 7. Mode chooser

- [ ] A simple screen listing modes and levels, built now that there are two
      modes to choose between. Replaces the temporary mode key

## 8. Two players (what Turboraketti really was)

- [ ] `levels/twoplayer/` with its own loader and fields (a base per player)
- [ ] Two ships in World, two input sets (WASD + arrows, or one gamepad
      each: `readInput(player)` from gamepad.md)
- [ ] Split screen: two cameras rendering to separate offscreen Images
- [ ] Shooting and projectiles
- [ ] Hits, damage, death and respawn at base
- [ ] Base station: choose weapons and fuel, repair
- [ ] Mass = ship + fuel + ammo (the trade-off between load and speed)
- [ ] Liquid (drag per pixel) and more levels

## Small and useful (done)

- [x] Version info in the HUD (git commit + time), to see which build the web
      version is running (`game/version.go`)
- [x] Arrows towards the nearest crystal and pad (start pad once all
      crystals are taken)
- [x] Visual effect while refuelling
- [x] Levels as data (JSON or image) instead of code
