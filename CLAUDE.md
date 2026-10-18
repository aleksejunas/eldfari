# Raketti

Go + Ebitengine remake of Turboraketti (Amiga): gravity, landing, fuel, crystals.
Two players come later.

## Commands

```sh
go run .          # play
go test ./...     # must be green before commit
go vet ./...
```

## Architecture

- `sim/` is deterministic and does **not** import Ebitengine. All game logic lives here and is
  tested without a window. Fixed time step `Dt = 1/60`, one `World.Step` per `Update`.
- `main.go` is only window, input, camera, drawing and HUD. `fx.go` is particles (cosmetic only,
  uses `math/rand`, never affects `sim`).
- New game logic belongs in `sim/` with tests, not in `main.go`.

## Physics (constants in `sim/world.go`)

- Gravity 120 px/s², acceleration = Thrust 360 / mass, mass = 1 + 0.6·fuel, drag 0.15 /s.
- Landing: every contact point touching terrain must be on the same pad, and speed and angle must
  be gentle (`MaxLandVY`, `MaxLandVX`, `MaxLandAngle`, see `gentle()`). Otherwise crash.
- Win: collect all crystals and land on the start pad (`Pads[0]`). The timer starts at first liftoff.
- Pads refuel the ship while it stands on them.

## Terrain

Polygons filled even-odd, rasterized to a bool mask for pixel collision (`sim/level.go`).
Polygon edges must not cross each other (that makes holes). Everything outside the world is solid.

## Level format

- JSON in `levels/`, embedded with `//go:embed *.json`, loaded sorted by file name (`NN-name.json`).
- Parsed and validated in `sim/levelfile.go`: unknown fields are errors, and all problems are
  reported together (crossing edges, narrow/floating pads, clearance above pads and around
  crystals, spawn). `pads[0]` is start and goal. `TestShippedLevelsLoadAndValidate` checks them all.
- Level 1 Grotta (intro), 2 Passasjen (gentle step up), 3 Dypet (hard, narrow passages).
  If a level is too hard, tune it. Level names are proper names and stay Norwegian.

## Conventions

- Tests green before commit.
- Simple, readable code over clever code. No unnecessary complexity or abstractions.
- Code, comments, error messages and docs are in English. Conversation with the user is Norwegian.

## Roadmap (in this order)

1. Small things: arrow or minimap towards the nearest crystal and pad, save best time to file,
   physics substeps, fuel effect.
2. WASM build hosted on GitHub Pages, so a non-developer friend can test. The user writes this
   code themselves: explain and give code to copy, do not edit files for it.
3. Cargo on a rope, Thrust style.
4. Two players in split screen: shooting, base, weapons, liquid drag, more levels.
5. In-game level editor.
