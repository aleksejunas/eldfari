package game

import (
	"image/color"
	"math"
	"math/rand"

	"eldfari/sim"
)

// particle is used for exhaust streaks, fuel sparks and wreckage: a short line segment
// with position, velocity, angle and lifetime.
type particle struct {
	x, y, vx, vy  float64
	ang, spin     float64
	length        float64
	life, maxLife float64
	grav          float64
	col           color.RGBA
}

func (p *particle) update(l *sim.Level) {
	px, py := p.x, p.y
	p.vy += p.grav * sim.Dt
	p.x += p.vx * sim.Dt
	p.y += p.vy * sim.Dt
	p.ang += p.spin * sim.Dt
	p.life -= sim.Dt
	// Falling wreckage lands on the terrain: back out, bounce a little and
	// lose most of the speed and spin, so it comes to rest on the ground.
	if p.grav > 0 && l.Solid(p.x, p.y) {
		p.x, p.y = px, py
		p.vx, p.vy, p.spin = p.vx*0.4, -p.vy*0.25, p.spin*0.3
	}
}

// exhaust returns an exhaust particle behind the ship, opposite the nose.
func exhaust(s *sim.Ship) particle {
	pos := s.ToWorld(rand.Float64()*4-2, sim.ShipRear+1)
	sin, cos := math.Sincos(s.Angle)
	speed := 140 + rand.Float64()*60
	side := (rand.Float64() - 0.5) * 50
	vx := s.VX - sin*speed + cos*side
	vy := s.VY + cos*speed + sin*side
	life := 0.25 + rand.Float64()*0.2
	return particle{
		x: pos.X, y: pos.Y, vx: vx, vy: vy,
		ang: math.Atan2(vy, vx), length: 3,
		life: life, maxLife: life,
		col: colTerrain,
	}
}

// explosion breaks the ship apart: its three sides tumble away with the
// ship's speed, small pieces fly out, and a short burst of sparks marks the
// hit. Pieces fall with real gravity and come to rest on the terrain.
func explosion(s *sim.Ship) []particle {
	var out []particle
	v := s.Points()
	for i := range v {
		a, b := v[i], v[(i+1)%len(v)]
		mx, my := (a.X+b.X)/2, (a.Y+b.Y)/2
		out = append(out, particle{
			x: mx, y: my,
			vx:  s.VX*0.3 + (mx-s.X)*4 + (rand.Float64()-0.5)*40,
			vy:  s.VY*0.3 + (my-s.Y)*4 - 40 - rand.Float64()*40,
			ang: math.Atan2(b.Y-a.Y, b.X-a.X), spin: (rand.Float64() - 0.5) * 8,
			length: math.Hypot(b.X-a.X, b.Y-a.Y),
			life:   2.5, maxLife: 2.5,
			grav: sim.Gravity, col: colShip,
		})
	}
	for range 8 {
		a := rand.Float64() * 2 * math.Pi
		speed := 40 + rand.Float64()*90
		life := 1.4 + rand.Float64()*0.8
		out = append(out, particle{
			x: s.X, y: s.Y,
			vx: s.VX*0.3 + math.Cos(a)*speed, vy: s.VY*0.3 + math.Sin(a)*speed,
			ang: rand.Float64() * 2 * math.Pi, spin: (rand.Float64() - 0.5) * 14,
			length: 3 + rand.Float64()*4,
			life:   life, maxLife: life,
			grav: sim.Gravity, col: colTerrain,
		})
	}
	for range 14 {
		a := rand.Float64() * 2 * math.Pi
		speed := 120 + rand.Float64()*160
		life := 0.2 + rand.Float64()*0.25
		vx, vy := math.Cos(a)*speed, math.Sin(a)*speed
		out = append(out, particle{
			x: s.X, y: s.Y, vx: vx, vy: vy,
			ang: math.Atan2(vy, vx), length: 4,
			life: life, maxLife: life,
			col: colGem,
		})
	}
	return out
}

// refuelSpark returns a spark that rises from the pad up through the ship
func refuelSpark(s *sim.Ship) particle {
	life := 0.3 + rand.Float64()*0.2
	return particle{
		x:   s.X + rand.Float64()*16 - 8,
		y:   s.Y + sim.ShipRear + 0.5, // the pad surface under the ship
		vy:  -40 - rand.Float64()*30,
		ang: -math.Pi / 2, length: 3,
		life: life, maxLife: life,
		col: colGem,
	}

}
