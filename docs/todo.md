# Raketti – todo

## 1. Crystals (in progress)

- [x] Crystals in the level (Level.Crystals), picked up by flying through them
- [x] Counter and timer in the HUD (timer starts at first liftoff)
- [x] Win by landing on the start pad with all crystals taken
- [ ] Best time (in memory first, later saved to file)
- [ ] Tests: pickup, win, timer, free air around each crystal

## 2. Cargo on a rope (Thrust style)

- [ ] Ball/crate on the ground with a "hook" distance to the ship
- [ ] Rope as a hard distance constraint (rigid first, maybe spring later)
- [ ] The ball is affected by gravity and pulls on the ship (pendulum)
- [ ] Collision for the ball against terrain
- [ ] Delivery: the ball must reach a specific pad
- [ ] Tune mass: heavy cargo = slower acceleration

## 3. Multiple game modes

- [ ] Minimal (just collect crystals)
- [ ] 2 player
- [ ]

## 4. Two players (what Turboraketti really was)

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
