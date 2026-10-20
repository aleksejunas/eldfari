# Raketti

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
- Browser build: `web/index.html` + GitHub Actions (`.github/workflows/`)
  deploys to GitHub Pages on every push to `master`.

## Physics (constants in `sim/world.go`)

- Gravity 120 px/s², acceleration = Thrust 360 / mass,
  mass = 1 + 0.6·fuel, drag 0.15 /s.
- Landing: every contact point touching terrain must be on the same pad, and
  speed and angle must be gentle (`MaxLandVY`, `MaxLandVX`, `MaxLandAngle`,
  see `gentle()`). Otherwise crash.
- Win: collect all crystals and land on the start pad (`Pads[0]`). The timer
  starts at first liftoff.
- Pads refuel the ship while it stands on them.

## Terrain

Polygons filled even-odd, rasterized to a bool mask for pixel collision
(`sim/level.go`). Polygon edges must not cross each other (that makes holes).
Everything outside the world is solid.

## Level format

- JSON in one folder per game mode (`levels/crystals/` for now), embedded in
  `levels/levels.go`. No mode field in the JSON: the folder is the mode. A
  mode's folder is added to the embed pattern when the mode is built
  (`go:embed` fails on an empty folder).
- File names start with a number with gaps: `10-grotta.json`. `LoadDir` sorts
  numerically (100 after 90) and rejects missing or duplicate numbers.
- Parsed and validated in `sim/levelfile.go`: unknown fields are errors, and
  all problems are reported together (crossing edges, narrow/floating pads,
  clearance above pads and around crystals, spawn). `pads[0]` is start and
  goal. `TestShippedLevelsLoadAndValidate` checks them all.
- Crystals levels, easy to hard: 10 Grotta, 20 Passasjen, 30 Sprekken, 40
  Dryppet, 50 Langveien, 60 Sjakta, 70 Dypet. The narrowest gap goes from
  ~7x ship height (Passasjen, 144 px) down to ~4x (Dypet, 79 px). Check a new
  or changed level with `go run ./cmd/levelstat`. If a level is too hard,
  tune it. Level names are proper names and stay Norwegian.
- Best times are keyed by mode and level name (`crystals/Grotta`).

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

## Roadmap

Details and checkboxes in `docs/todo.md`, where the sections are in the order
they will be done:

1. Done: browser build with intro page and best times in localStorage;
   levels in one folder per mode.
2. Done: crystals levels 30–60 and `cmd/levelstat`; they still need
   play-testing by someone else.
3. Next: gamepad (`docs/gamepad.md`).
4. Cargo on a rope: substeps first, then pendulum, mass, collision,
   delivery, cargo mode.
5. Mode chooser (once there are two modes).
6. Two players in split screen.
7. Later: in-game level editor.
