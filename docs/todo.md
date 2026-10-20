# Raketti – todo

The sections are in the order they will be done.

## 1. Crystals (done)

- [x] Crystals in the level (Level.Crystals), picked up by flying through them
- [x] Counter and timer in the HUD (timer starts at first liftoff)
- [x] Win by landing on the start pad with all crystals taken
- [x] Best time, saved to file (`game/besttimes.go`)
- [x] Tests: pickup, win, timer, free air around each crystal

## 2. Share the game: playable in the browser

I want to write this code myself. Claude explains and gives code to copy,
but does not edit files.

- [x] WASM build (`GOOS=js GOARCH=wasm`) + `wasm_exec.js` + a small
      `index.html`
- [x] Host on GitHub Pages (static, a GitHub Actions workflow that builds
      and deploys)
- [ ] Best times in the browser (no config dir there: use localStorage or
      keep in memory)
- [ ] Short instructions for a non-developer: link, keys, what the goal is

## 3. Levels: one folder per mode

Mode = folder: `levels/crystals/`, `levels/cargo/`, `levels/twoplayer/`.
No "mode" field in the JSON. Each mode has its own loader and its own fields,
and `DisallowUnknownFields` stays strict, so data for one mode is an error in
another mode's levels.

- [ ] Move today's levels: `01-grotta` -> `crystals/10-grotta`,
      `02-passasjen` -> `crystals/20-passasjen`,
      `03-dypet` -> `crystals/70-dypet`
- [ ] `//go:embed crystals/*.json`. A mode's folder and embed pattern are
      added when that mode is built (`go:embed` fails on an empty folder)
- [ ] File names start with a number with gaps (10, 20, 30 ...), so new
      levels can be slotted in without renaming. Sort on the number, not as
      text, so 100 comes after 90
- [ ] N/P only steps through the levels of the current mode
- [ ] Best times keyed by mode + level name, so equal names in two modes
      don't share a record
- [ ] Update paths in README and CLAUDE.md

## 4. New crystals levels

The ship is about 21 px tall. Level 30 is around 6x ship height at its
narrowest point, and the ramp goes down towards ~4x. One new difficulty per
level.

- [ ] Analysis first: a small Go tool that computes the minimum clearance
      along the route and the route length per level, as difficulty numbers
- [ ] Run it on 10, 20 and 70 to check the ramp below against real numbers
- [ ] Ramp:
  - 10 Grotta: open (exists)
  - 20 Passasjen: passages ~180 px (exists)
  - [ ] 30 (new): one narrower passage (~130 px), more pads
  - [ ] 40 (new): short low passage (~250 px) + stalactites
  - [ ] 50 (new): longer routes, fuel planning
  - [ ] 60 (new): one wide sloping passage (160 px+) as practice
  - 70 Dypet: narrow passages, sloping tunnel, long route (exists)
- [ ] Every new level is play-tested by someone else (via the web version)

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
