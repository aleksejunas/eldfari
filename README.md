# Raketti

Vector-style gravity cave flyer, inspired by Turboraketti (Amiga, Heikki Kosola).
No graphics or levels are copied from the original. Only the idea.

## Run

```sh
go run .
go test ./...    # physics, collision, landing and level validation
```

Keys: `A`/`D` or arrows = rotate, `W`/up/space = thrust, `R` = restart, `N`/`P` = next/previous
level, `Z` = zoom, `F` = fullscreen.

Linux needs the usual Ebitengine libraries (X11/GL/ALSA). See
<https://ebitengine.org/en/documents/install.html> if the build complains.

## Layout

| File | Contents |
|---|---|
| `sim/world.go` | Ship, physics (gravity, thrust/mass, drag), collision, landing, crystals, win. All constants at the top. |
| `sim/level.go` | Terrain as polygons, rasterized to an invisible bitmask. Pads. |
| `sim/levelfile.go` | JSON level format, loading and validation. |
| `levels/*.json` | The levels, embedded in the binary and sorted by file name. |
| `main.go` | Ebitengine: input, camera, drawing, HUD. |
| `fx.go` | Exhaust and wreckage (line segments). |

The idea: `sim` knows nothing about Ebitengine, so it is deterministic and testable without a window.
Fixed time step (60 Hz), one `Step` per `Update`.

## Physics in short

- Engine acceleration = `Thrust / mass`, and mass = `BaseMass + FuelMass * fuel`.
  A full tank makes the ship slower. That is the core of the Turboraketti feel.
- Landing requires all hit points on the same pad, `|VY| <= 80`, `|VX| <= 50` and an angle
  within 0.5 rad of straight up. Otherwise you crash.
- Pads refuel the ship while it stands on them.
- Win by collecting all crystals and landing on the start pad. The timer starts at first liftoff.
- All numbers are guesses. Play and tune them in `sim/world.go`.

## Known limitations

- No substeps: above roughly 600 px/s the ship could in theory pass through thin walls.
- Collision is point tests (corners + edge midpoints), not full line-vs-mask.
- The ship crashes into the world border (everything outside the level is solid).
