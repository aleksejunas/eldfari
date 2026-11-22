# eldfari – todo

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
7. [x] Cargo levels, easy to hard:
       - 10 Vindáss: over a hill
       - 20 Búr: down a shaft into a cellar, out through a tunnel
       - 30 Hilla: onto a high shelf
       - 40 Grjót: slalom between stalactites and stalagmites
       - 50 Langferð: long tunnel with a lower part, refuel pad halfway
       - 60 Ljóri: S-shaped chimney, deliver at the top
       - 70 Myrkr: down into a pit, 200 px tunnel with the cargo, up a
         shaft to a small ledge
8. [x] Temporary key `M` to switch mode (replaced by the menu)
- [x] Made playable after the first test ("impossible"): hook range 120 px
      with a winch and a ring showing it, cargo mass 0.4, springy damped
      rope, swing damping; an autopilot test delivers every cargo level
- [ ] Play-test the cargo levels again; tune the constants in
      `sim/cargo.go`
- [x] `cmd/levelstat` measures cargo levels too: the leg with the cargo
      needs room for the cargo hanging 80 px below the ship

## 6. Mode chooser (done)

- [x] A menu listing modes and levels with best times (`game/menu.go`):
      left/right mode, up/down level, Enter to play, `M`/`Esc` back. The
      game starts in it. Replaces the temporary mode key

## 7. Gamepad support

Full plan: [gamepad.md](gamepad.md)

- [x] Keyboard and gamepad input merged in `readInput(player)`; one pad
      per player; RT thrust, A fire, B bomb
- [x] Tested on real hardware: PS4 (USB and Bluetooth), PDP Switch pad
      (Bluetooth only; own mapping in `extraMappings`)
- [ ] Joy-Cons: measure with `go run ./cmd/padlisten`, add to `extraMappings`
- [ ] More controllers: measure each new one the same way
- [ ] Check the PS4 pad in the web version (browser mapping differs)

## 8. Two players (what Turboraketti really was)

Duel mode first (playable), then a race mode.

- [x] `levels/duel/`: pads[0] and pads[1] are the players' bases (Hólmganga,
      Vígvöllr)
- [x] Two ships: `sim.Duel` gives each player a World sharing the Level;
      keys WASD + S fire, arrows + down fire (gamepad: `readInput(player)`
      from gamepad.md, later)
- [x] Split screen: two cameras rendering to separate offscreen Images
- [x] Shooting and projectiles
- [x] Hits, damage (4 hits), death and respawn at base; first to 5 wins
- [x] Repair and refuel while landed on the own base
- [x] Bombs: drop with the ship's speed, fall, blast radius 45 px, two
      hits destroy; weapons on fingers that are free while flying
- [ ] Play-test the duel; tune `sim/duel.go` (bullet speed, damage, reload)
- [x] Race mode: both race for the same crystals (one taken is gone for
      both); when all are taken and both have landed, most crystals wins.
      Levels Kappreið and Gullhellir (symmetric, odd crystal count)
- [ ] Play-test the race
- [ ] Base station: choose weapons and fuel
- [ ] Mass = ship + fuel + ammo (the trade-off between load and speed)
- [ ] Liquid (drag per pixel) and more levels

## Small and useful (done)

- [x] Version info in the HUD (git commit + time), to see which build the web
      version is running (`game/version.go`)
- [x] Arrows towards the nearest crystal and pad (start pad once all
      crystals are taken)
- [x] Visual effect while refuelling
- [x] Levels as data (JSON or image) instead of code
- [x] Old Norse level names (Grotta -> Hellir, Kranen -> Vindáss, ...)
- [x] The game is called eldfari (was Raketti), repo and page too
