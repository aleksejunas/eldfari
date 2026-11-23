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
	for range int(math.Round(1/BulletDamage)) - 1 { // the rest of the hits it takes
		shoot(d)
		run3(d, 1)
	}
	if p.Alive() || d.Players[0].Score != 1 {
		t.Fatalf("alive %v, score %d: %v hits should destroy the ship and give a point",
			p.Alive(), d.Players[0].Score, math.Round(1/BulletDamage))
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

func TestWreckDoesntThrust(t *testing.T) {
	d := duelArena()
	fly(d.Players[0].World, 500, 470, 0, 300, 0)
	for range 10 {
		d.Step([2]Input{{Thrust: true}, {}}) // thrust held into the crash
	}
	if p := d.Players[0]; p.Alive() || p.Ship.Thrusting {
		t.Errorf("alive %v, thrusting %v: a wreck must not thrust", p.Alive(), p.Ship.Thrusting)
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

func TestBombFallsAndExplodesOnTheFloor(t *testing.T) {
	d := duelArena()
	fly(d.Players[0].World, 300, 300, 50, 0, 0)
	d.Step([2]Input{{Bomb: true}, {}})
	if len(d.Bombs) != 1 {
		t.Fatalf("%d bombs, want one", len(d.Bombs))
	}
	if b := d.Bombs[0]; b.VX < 40 || b.VY <= 0 {
		t.Errorf("bomb velocity (%.0f,%.0f): it should keep the ship's speed and fall", b.VX, b.VY)
	}
	d.Step([2]Input{{Bomb: true}, {}})
	if len(d.Bombs) != 1 {
		t.Error("the bomb bay should reload before the next bomb")
	}
	run3(d, 120) // 2 s: long enough to fall 200 px
	if len(d.Bombs) != 0 || len(d.Blasts) == 0 && d.Players[0].Alive() {
		t.Errorf("%d bombs left, %d blasts: the bomb should have exploded on the floor", len(d.Bombs), len(d.Blasts))
	}
}

func TestBombBlastHurtsTheOtherShipNearby(t *testing.T) {
	d := duelArena()
	target(d)                                             // player 1 hovers at (500,300)
	d.Bombs = []Bomb{{X: 500, Y: 270, VY: 150, Owner: 0}} // faster than the falling target
	run3(d, 30)
	if want := 1 - BombDamage; d.Players[1].Health > want+1e-9 {
		t.Errorf("health %.2f, want at most %.2f after a bomb hit", d.Players[1].Health, want)
	}

	far := duelArena()
	far.Bombs = []Bomb{{X: 500, Y: 400, Owner: 0}} // explodes on the floor, far from both bases
	run3(far, 60)
	if far.Players[1].Health != 1 || far.Players[0].Health != 1 {
		t.Error("a blast far away must not hurt anyone")
	}
}

func TestOwnBombDoesntHurt(t *testing.T) {
	d := duelArena()
	fly(d.Players[0].World, 500, 300, 0, 0, 0)
	d.Bombs = []Bomb{{X: 500, Y: 320, Owner: 0}}
	d.Players[0].Ship.Fuel = 0
	run3(d, 5)
	if d.Players[0].Health != 1 {
		t.Errorf("health %.2f: your own bomb must not hurt you", d.Players[0].Health)
	}
}

func TestShipsCrashIntoEachOther(t *testing.T) {
	d := duelArena()
	fly(d.Players[0].World, 480, 300, 100, 0, 0)
	fly(d.Players[1].World, 520, 300, -100, 0, 0)
	run3(d, 20)
	for i, p := range d.Players {
		if p.Alive() || p.Score != 0 {
			t.Errorf("player %d: alive %v, score %d; both should crash and nobody score", i, p.Alive(), p.Score)
		}
	}
}

func TestBulletHitLeavesASpark(t *testing.T) {
	d := duelArena()
	target(d)
	shoot(d)
	run3(d, 1)
	if len(d.Bullets) != 0 || len(d.Hits) != 1 {
		t.Errorf("%d bullets, %d hits: the bullet should be gone and the hit shown", len(d.Bullets), len(d.Hits))
	}
}
