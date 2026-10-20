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
- [x] Run it on 10, 20 and 70: Gangr is 144 px at its narrowest (under
      the pillar), not ~180, so the ramp below starts from there
- [x] Ramp (narrowest gap, route length):
  - 10 Hellir (was Grotta): open, 272 px, 6.6k px
  - 20 Gangr (was Passasjen): 144 px (6.9x), 6.9k px
  - [x] 30 Rifa: one 130 px crack between two halls, five pads:
        128 px (6.1x), 8.1k px
  - [x] 40 Dropi: short low passage (250 px long) + stalactites:
        112 px (5.3x), 8.1k px
  - [x] 50 Langvegr: long snake, refuel halfway: 104 px (5.0x), 14.2k px
  - [x] 60 Gjá: one wide sloping shaft (168 px) as practice, then a low
        passage: 96 px (4.6x), 8.5k px
  - 70 Djúp (was Dypet): narrow passages, sloping tunnel, long route:
    79 px (3.7x), 8.3k px
- [ ] Every new level is play-tested by someone else (via the web version).
      Check especially fuel on Langvegr

## 5. Cargo on a rope (Thrust style) (done)

1. [x] Substeps in the physics: 4 per frame (`Substeps` in `sim/world.go`)
2. [x] Rope as a hard distance constraint: slack allowed, never stretches;
       tested that the distance never exceeds L
3. [x] Cargo mass (0.8): pulled by gravity, pulls on the ship, slows it
4. [x] Collision for the cargo: bounces a little and slides, no damage
5. [x] Pick-up and delivery: hooks on within L (80 px); delivered when it
       rests calmly on the goal pad
6. [x] Cargo mode: `levels/cargo/`, own loader (`"cargo"`, `"goal"` pad
       name), `sim.Mode` with a `switch`
7. [x] Cargo levels: 10 Vindáss (over a hill), 20 Búr (down a shaft into
       a cellar, out through a tunnel), 30 Hilla (onto a high shelf)
8. [x] Temporary key `M` to switch mode (replaced by the menu)
- [x] Made playable after the first test ("impossible"): hook range 120 px
      with a winch and a ring showing it, cargo mass 0.4, springy damped
      rope, swing damping; an autopilot test delivers every cargo level
- [ ] Play-test the cargo levels again; tune the constants in
      `sim/cargo.go`
- [ ] `cmd/levelstat` only measures crystals levels so far

## 6. Mode chooser (done)

- [x] A menu listing modes and levels with best times (`game/menu.go`):
      left/right mode, up/down level, Enter to play, `M`/`Esc` back. The
      game starts in it. Replaces the temporary mode key

## 7. Gamepad support

Full plan: [gamepad.md](gamepad.md)

- [ ] Keyboard and gamepad input merged in one `readInput()`
- [ ] Tested on real hardware (xbox/playstation pad, Steam Deck)

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
- [x] Old Norse level names (Grotta -> Hellir, Kranen -> Vindáss, ...);
      `renamedLevels` in `game/besttimes.go` moves old records
