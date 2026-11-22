# eldfari

Go + Ebitengine remake of Turboraketti (Amiga): gravity, landing, fuel,
crystals. Two players come later.

## Commands

```sh
go run .          # play
go test ./...     # must be green before commit
go vet ./...
go run ./cmd/levelstat   # narrowest gap + route length per level
```

## Architecture

- `sim/` is deterministic and does **not** import Ebitengine. All game logic
  lives here and is tested without a window. Fixed time step `Dt = 1/60`, one
  `World.Step` per `Update`.
- `game/` is the Ebitengine front end: window, input, camera, drawing and HUD
  (`game.go`), particles (`fx.go`, cosmetic only, uses `math/rand`, never
  affects `sim`), best times and version info. Root `main.go` only calls
  `game.Run()`.
- New game logic belongs in `sim/` with tests, not in `game/`.
- Game modes: `sim.Mode` (crystals, cargo, duel, race) on each `Level`;
  `World.Step` switches on it. Cargo mode lives in `sim/cargo.go` (rope,
  cargo, delivery) and `game/cargo.go` (drawing). Duel (`sim/duel.go`) and
  race (`sim/race.go`) give each player a World sharing the Level;
  `game/twoplayer.go` draws the split screen for both. The menu
  (`game/menu.go`, `M`/`Esc`) chooses mode and level; the game starts in it.
- Input: `game/input.go` merges the keyboard and gamepads
  (`readInput(player)`); the pure `padInput` mapping is tested. A controller
  that maps wrong: measure it with `go run ./cmd/padlisten` and add a line to
  `extraMappings`.
- Browser build: `web/index.html` + GitHub Actions (`.github/workflows/`)
  deploys to GitHub Pages on every push to `master`.

## Physics (constants in `sim/world.go`)

- Gravity 120 px/s², acceleration = Thrust 360 / mass,
  mass = 1 + 0.6·fuel, drag 0.15 /s.
- Landing: every contact point touching terrain must be on the same pad, and
  speed and angle must be gentle (`MaxLandVY`, `MaxLandVX`, `MaxLandAngle`,
  see `gentle()`). Otherwise crash.
- Crystals win: collect all crystals and land on the start pad (`Pads[0]`).
  The timer starts at first liftoff.
- Pads refuel the ship while it stands on them.
- 4 substeps per frame (`Substeps`), so fast movement and the rope stay
  stable.
- Cargo (`sim/cargo.go`): hooks on within 120 px (a dashed ring shows it),
  a winch reels the rope in to 80 px. The rope can go slack; tight, it is a
  stiff damped spring (max 10 px stretch), and the swing is damped a little.
  Cargo mass 0.4. Bounces without damage; delivered when it rests calmly on
  the goal pad. `TestCargoLevelsAreDeliverable` flies every cargo level with
  a simple autopilot: a new cargo level needs a route there.

## Terrain

Polygons filled even-odd, rasterized to a bit mask for pixel collision
(`sim/level.go`). Polygon edges must not cross each other (that makes holes).
Everything outside the world is solid.

## Level format

- JSON in one folder per game mode (`levels/crystals/`, `levels/cargo/`),
  embedded in `levels/levels.go`. No mode field in the JSON: the folder is
  the mode. A new mode's folder is added to the embed pattern when the mode
  is built (`go:embed` fails on an empty folder).
- File names start with a number with gaps: `10-hellir.json`. `LoadDir` sorts
  numerically (100 after 90) and rejects missing or duplicate numbers.
- Parsed and validated in `sim/levelfile.go`: unknown fields are errors, and
  all problems are reported together (crossing edges, narrow/floating pads,
  clearance above pads and around crystals, spawn). `pads[0]` is start and
  goal. `TestShippedLevelsLoadAndValidate` checks them all.
- Level names are Old Norse; file names are ASCII. Crystals levels, easy to
  hard: 10 Hellir, 20 Gangr, 30 Rifa, 40 Dropi, 50 Langvegr, 60 Gjá, 70
  Djúp. The narrowest gap goes from ~7x ship height (Gangr, 144 px) down to
  ~4x (Djúp, 79 px). Check a new
  or changed level with `go run ./cmd/levelstat`. If a level is too hard,
  tune it. Level names are proper names and stay Norwegian.
- Cargo levels add `"cargo": [x, y]` and `"goal": "pad name"` instead of
  crystals, easy to hard: 10 Vindáss, 20 Búr, 30 Hilla, 40 Grjót, 50
  Langferð, 60 Ljóri, 70 Myrkr. `cmd/levelstat` measures them
  with the cargo hanging below the ship (gap 200–424 px).
- Best times are keyed by mode and level name (`crystals/Hellir`). Renaming
  a level loses its records, which is fine.

## Conventions

- Tests green before commit.
- Simple, readable code over clever code. No unnecessary complexity or
  abstractions.
- Code, comments, error messages and docs are in English. Conversation with
  the user is Norwegian.
- Markdown lines are at most 80 characters (markdownlint MD013). Use lists
  instead of tables when rows would get longer.
- Browser/WASM code (`web/`, workflow, browser storage): the user writes it
  themselves. Explain and give code to copy, do not edit those files.

## Working style (save tokens)

- Read `NOTES.md` first; don't explore the repo for an overview. Ask if
  something is unclear.
- Search (`rg`/`grep`) before reading; read only the relevant lines. Don't
  read files over ~200 lines without reason, and don't re-read a file after
  editing it. Don't dump level JSON; use `jq`.
- Limit output: `go build ./... 2>&1 | tail -30`,
  `go test ./... 2>&1 | tail -40`. No `-v` unless asked or a test fails.
- Don't run the game to check it; the user tests gameplay.
- Short answers: no recap of what was just done, no repeated code. Show a
  diff or file names, not whole files. For choices: 2–3 lines per option,
  then a recommendation. No subagents unless asked.
- After each finished step or decision, rewrite `NOTES.md` (max ~40 lines:
  Now, Next, Decisions, Known problems) and commit with a message that says
  why. When a task is done: update `NOTES.md`, commit, and say "ferdig, bra
  tidspunkt for ny økt".

## Roadmap

Details and checkboxes in `docs/todo.md`, where the sections are in the order
they will be done:

1. Done: browser build with intro page and best times in localStorage;
   levels in one folder per mode.
2. Done: crystals levels 30–60 and `cmd/levelstat`; they still need
   play-testing by someone else.
3. Done: cargo on a rope (substeps, rope, cargo mode, 3 levels); needs
   play-testing.
4. Done: menu to choose mode and level.
5. Next: gamepad (`docs/gamepad.md`).
6. Two players in split screen.
7. Later: in-game level editor.
