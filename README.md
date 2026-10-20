# eldfari

Vector-style gravity cave flyer, inspired by Turboraketti (Amiga, Heikki
Kosola). No graphics or levels are copied from the original. Only the idea.

The spark came from [Quattroids](https://github.com/28allday/Quattroids)
by 28allday, a vector arcade shooter for the Omarchy shell. It brought
Turboraketti back to mind, and its clean vector look is the visual
inspiration for eldfari.

Written in Go, using [Ebitengine](https://ebitengine.org/).

It is still at an early stage. More levels, a level editor and several game
modes are on the way.

## Play in the browser

**<https://aleksejunas.github.io/eldfari/>**

Fly through the caves, collect every crystal and land gently back on the pad
you started from. Land on any pad to refuel. Too fast or too tilted and you
crash. Your time starts at liftoff.

Cargo mode: fly inside the dashed ring around the cargo, it hooks on to a
rope by itself, and lower it gently onto the marked pad. The cargo swings,
so move smoothly.

Keys: `A`/`D` or arrows = rotate, `W`/up/space = thrust, `R` = restart,
`N`/`P` = next/previous level, `M`/`Esc` = menu (mode and level), `Z` = zoom,
`F` = fullscreen, `I` = FPS and
slowest frame (to spot hitches).

If some keys do nothing, the browser is probably catching them: Vivaldi's
single-key shortcuts or a Vim-style extension (Vimium etc.). Turn those off
for the site, or use another browser. In the browser, best times are saved
in localStorage.

## Run locally

```sh
go run .
go test ./...    # physics, collision, landing and level validation
```

Best times are saved to `~/.config/eldfari/best.json` (per mode and level
name).

Linux needs the usual Ebitengine libraries (X11/GL/ALSA). See
<https://ebitengine.org/en/documents/install.html> if the build complains.

## Browser build

Every push to `master` runs `.github/workflows/pages.yml`: it tests `sim/`,
builds `web/eldfari.wasm`, copies `wasm_exec.js` from the Go toolchain and
deploys `web/` to GitHub Pages.

To test the browser build locally:

```sh
GOOS=js GOARCH=wasm go build -o web/eldfari.wasm .
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/
python3 -m http.server -d web 8080    # open http://localhost:8080
```

`wasm_exec.js` must come from the same Go version that built the `.wasm`.
Both are build output and ignored by git.

## Layout

- `sim/world.go`: ship, physics (gravity, thrust/mass, drag), collision,
  landing, crystals, win. All constants at the top.
- `sim/level.go`: terrain as polygons, rasterized to an invisible bitmask.
  Pads.
- `sim/cargo.go`: cargo mode: rope, cargo physics, delivery.
- `sim/levelfile.go`: JSON level format per mode, loading and validation.
- `cmd/levelstat`: prints difficulty numbers per level (narrowest gap, route
  length); `-png dir` draws each level with its route.
- `levels/crystals/*.json`, `levels/cargo/*.json`: the levels, one folder per
  game mode, embedded
  in the binary and sorted by the number at the start of the file name.
- `main.go`: starts the game, nothing else.
- `game/game.go`: Ebitengine: window, input, camera, drawing.
- `game/cargo.go`: draws the rope, the cargo and the goal pad.
- `game/hud.go`: the HUD (fuel, speed, time, crystals, messages), drawn with
  the Go fonts and scaled by the window height.
- `game/fx.go`: exhaust, wreckage and refuel sparks (line segments).
- `game/besttimes.go`: best time per level, saved as JSON in the user's
  config dir (memory only in the browser).
- `game/version.go`: git commit and time the binary was built from, shown in
  the HUD.
- `web/index.html`: page that loads `wasm_exec.js` and starts `eldfari.wasm`.
- `.github/workflows/pages.yml`: builds the WASM version and deploys it to
  GitHub Pages.

The idea: `sim` knows nothing about Ebitengine, so it is deterministic and
testable without a window. Fixed time step (60 Hz), one `Step` per `Update`.

## Physics in short

- Engine acceleration = `Thrust / mass`, and
  mass = `BaseMass + FuelMass * fuel`. A full tank makes the ship slower.
  That is the core of the Turboraketti feel.
- Landing requires all hit points on the same pad, `|VY| <= 80`,
  `|VX| <= 50` and an angle within 0.5 rad of straight up. Otherwise you
  crash.
- Pads refuel the ship while it stands on them.
- Win by collecting all crystals and landing on the start pad. The timer
  starts at first liftoff.
- All numbers are guesses. Play and tune them in `sim/world.go`.

## Known limitations

- No substeps: above roughly 600 px/s the ship could in theory pass through
  thin walls.
- Collision is point tests (corners + edge midpoints), not full
  line-vs-mask.
- The ship crashes into the world border (everything outside the level is
  solid).
