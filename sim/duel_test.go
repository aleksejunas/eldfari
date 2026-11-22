package sim

import (
	"math"
	"testing"

	"eldfari/levels"
)

// An open arena with a floor at y=500 and a base at each end.
func duelArena() *Duel {
	floor := []Pt{{0, 500}, {1000, 500}, {1000, 600}, {0, 600}}
	l := NewLevel(1000, 600, [][]Pt{floor}, []Pad{{50, 150, 500}, {850, 950, 500}}, Pt{100, 500 - ShipRear - 0.5})
	l.Mode = ModeDuel
	return NewDuel(l)
}

// shoot puts a bullet from player 0 right next to player 1, flying right, so
// it hits in the next step.
func shoot(d *Duel) {
	s := &d.Players[1].Ship
	d.Bullets = append(d.Bullets, Bullet{X: s.X - 4, Y: s.Y, VX: BulletSpeed, Life: 1, Owner: 0})
}

// target puts player 1 in the air, away from its base (where it would be
// repaired between hits).
func target(d *Duel) { fly(d.Players[1].World, 500, 300, 0, 0, 0) }

func TestShippedDuelLevelsLoad(t *testing.T) {
	ls, err := LoadDir(levels.Duel, ModeDuel)
	if err != nil {
		t.Fatalf("the duel levels are not valid:\n%v", err)
	}
	for _, l := range ls {
		d := NewDuel(l)
		run3(d, 60)
		for i, p := range d.Players {
			if p.Ship.State != Landed || p.Score != 0 {
				t.Errorf("%s: player %d after 1 s without input: %v, score %d", l.Name, i, p.Ship.State, p.Score)
			}
		}
	}
}

func TestPlayersSpawnOnTheirBases(t *testing.T) {
	d := duelArena()
	for i, p := range d.Players {
		pad := d.Level.Pads[i]
		if p.Ship.State != Landed || p.Ship.X != (pad.X0+pad.X1)/2 || p.Health != 1 {
			t.Errorf("player %d: %+v, want landed on pad %v with full health", i, p.Ship, pad)
		}
	}
}

func TestFireFromTheNoseWithReload(t *testing.T) {
	d := duelArena()
	fly(d.Players[0].World, 300, 200, 0, 0, 0) // nose up
	d.Step([2]Input{{Fire: true}, {}})
	if len(d.Bullets) != 1 || d.Bullets[0].VY > -BulletSpeed+10 || d.Bullets[0].Y > 200-10 {
		t.Fatalf("bullets %+v, want one flying up from the nose", d.Bullets)
	}
	d.Step([2]Input{{Fire: true}, {}})
	if len(d.Bullets) != 1 {
		t.Errorf("%d bullets: the gun should reload before the next shot", len(d.Bullets))
	}
	run2 := func(frames int) {
		for range frames {
			d.Step([2]Input{{Fire: true}, {}})
		}
	}
	run2(int(ReloadTime / Dt))
	if len(d.Bullets) != 2 {
		t.Errorf("%d bullets after the reload time, want 2", len(d.Bullets))
	}
}

func TestBulletsDamageThenDestroy(t *testing.T) {
	d := duelArena()
	target(d)
	p := &d.Players[1]
	shoot(d)
	run3(d, 1)
	if math.Abs(p.Health-(1-BulletDamage)) > 1e-9 || len(d.Bullets) != 0 {
		t.Fatalf("health %.2f, %d bullets: one hit should take %.2f", p.Health, len(d.Bullets), BulletDamage)
	}
	for range 3 {
		shoot(d)
		run3(d, 1)
	}
	if p.Alive() || d.Players[0].Score != 1 {
		t.Fatalf("alive %v, score %d: four hits should destroy the ship and give a point", p.Alive(), d.Players[0].Score)
	}
	run3(d, int(RespawnTime/Dt)+1)
	pad := d.Level.Pads[1]
	if !p.Alive() || p.Health != 1 || p.Ship.X != (pad.X0+pad.X1)/2 {
		t.Errorf("after %v s the ship should be back on its base with full health: %+v", RespawnTime, p.Ship)
	}
}

func TestOwnBulletsDontHit(t *testing.T) {
	d := duelArena()
	shoot(d)
	d.Bullets[0].Owner = 1 // player 1's own bullet
	run3(d, 10)
	if d.Players[1].Health != 1 {
		t.Errorf("health %.2f: a player's own bullet must not hit them", d.Players[1].Health)
	}
}

func TestBulletStopsAtTerrain(t *testing.T) {
	d := duelArena()
	d.Bullets = []Bullet{{X: 500, Y: 496, VY: BulletSpeed, Life: 1, Owner: 0}}
	d.Step([2]Input{})
	if len(d.Bullets) != 0 {
		t.Error("a bullet flying into the floor should be gone")
	}
}

func TestCrashGivesTheOtherAPoint(t *testing.T) {
	d := duelArena()
	fly(d.Players[0].World, 500, 400, 0, 300, 0) // falling fast onto the floor
	run3(d, 30)
	if d.Players[0].Alive() || d.Players[1].Score != 1 {
		t.Errorf("alive %v, score %d: a crash should give the other player a point",
			d.Players[0].Alive(), d.Players[1].Score)
	}
}

func TestRepairOnOwnBase(t *testing.T) {
	d := duelArena()
	d.Players[1].Health = 0.5
	run3(d, 70)
	if d.Players[1].Health != 1 {
		t.Errorf("health %.2f after 1 s on the own base, want 1", d.Players[1].Health)
	}
}

func TestFirstToDuelScoreWins(t *testing.T) {
	d := duelArena()
	d.Players[0].Score = DuelScore - 1
	target(d)
	d.Players[1].Health = BulletDamage
	shoot(d)
	run3(d, 1)
	if d.Winner != 0 {
		t.Fatalf("winner %d, want player 0", d.Winner)
	}
	d.Step([2]Input{{Fire: true}, {Thrust: true}})
	if len(d.Bullets) != 0 || d.Players[0].Ship.State != Landed {
		t.Error("nothing should move once the duel is won")
	}
}

func run3(d *Duel, frames int) {
	for range frames {
		d.Step([2]Input{})
	}
}
