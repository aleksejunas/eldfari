package main

import (
	"math"
	"math/rand"

	"raketti/sim"
)

// particle brukes både til eksosstriper og vrakdeler: et kort linjestykke
// med posisjon, fart, vinkel og levetid.
type particle struct {
	x, y, vx, vy  float64
	ang, spin     float64
	length        float64
	life, maxLife float64
	grav          float64
}

func (p *particle) update() {
	p.vy += p.grav * sim.Dt
	p.x += p.vx * sim.Dt
	p.y += p.vy * sim.Dt
	p.ang += p.spin * sim.Dt
	p.life -= sim.Dt
}

// exhaust gir en eksospartikkel bak skipet, mot motsatt retning av nesa.
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
	}
}

// explosion gir vrakdeler som flyr utover og roterer.
func explosion(s *sim.Ship) []particle {
	out := make([]particle, 0, 9)
	for i := 0; i < 9; i++ {
		a := rand.Float64() * 2 * math.Pi
		speed := 30 + rand.Float64()*80
		life := 1.2 + rand.Float64()*0.6
		out = append(out, particle{
			x: s.X, y: s.Y,
			vx: math.Cos(a) * speed, vy: math.Sin(a) * speed,
			ang: rand.Float64() * 2 * math.Pi, spin: (rand.Float64() - 0.5) * 12,
			length: 5 + rand.Float64()*7,
			life:   life, maxLife: life,
			grav: 40,
		})
	}
	return out
}
