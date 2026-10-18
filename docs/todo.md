# Raketti – todo

## 1. Crystals (in progress)

- [x] Crystals in the level (Level.Crystals), picked up by flying through them
- [x] Counter and timer in the HUD (timer starts at first liftoff)
- [x] Win by landing on the start pad with all crystals taken
- [x] Best time, saved to file (`besttimes.go`)
- [ ] Tests: pickup, win, timer, free air around each crystal

## 2. Share the game: playable in the browser

I want to write this code myself. Claude explains and gives code to copy, but does not edit files.

- [ ] WASM build (`GOOS=js GOARCH=wasm`) + `wasm_exec.js` + a small `index.html`
- [ ] Host on GitHub Pages (static, e.g. a GitHub Actions workflow that builds and deploys)
- [ ] Best times in the browser (no config dir there: use localStorage or keep in memory)
- [ ] Short instructions for a non-developer: link, keys, what the goal is

## 3. Cargo on a rope (Thrust style)

- [ ] Ball/crate on the ground with a "hook" distance to the ship
- [ ] Rope as a hard distance constraint (rigid first, maybe spring later)
- [ ] The ball is affected by gravity and pulls on the ship (pendulum)
- [ ] Collision for the ball against terrain
- [ ] Delivery: the ball must reach a specific pad
- [ ] Tune mass: heavy cargo = slower acceleration

## 4. Multiple game modes

- [ ] Minimal (just collect crystals)
- [ ] 2 player
- [ ]

## 5. Two players (what Turboraketti really was)

- [ ] Two ships in World, two input sets (WASD + arrows)
- [ ] Split screen: two cameras rendering to separate offscreen Images
- [ ] Shooting and projectiles
- [ ] Hits, damage, death and respawn at base
- [ ] Base station: choose weapons and fuel, repair
- [ ] Mass = ship + fuel + ammo (the trade-off between load and speed)
- [ ] Liquid (drag per pixel) and more levels

## Small and useful

- [ ] Minimap or arrows towards the nearest pad
- [ ] Visual effect while refuelling
- [x] Levels as data (JSON or image) instead of code
- [ ] Substeps in the physics (high speed can pass through thin walls)
- [ ] Speedometer
