package sim

import "testing"

// placeAt puts a power-up of the kind right where player 0 is.
func placeAt(d *Duel, kind PowerUpKind) {
	s := &d.Players[0].Ship
	d.PowerUp = &PowerUp{X: s.X, Y: s.Y, Kind: kind, Life: PowerUpLife}
}

func TestPowerUpAppearsRarelyWithRoom(t *testing.T) {
	d := duelArena()
	run3(d, int(PowerUpMin/Dt)-1)
	if d.PowerUp != nil {
		t.Fatalf("a power-up appeared before %v s", PowerUpMin)
	}
	run3(d, int((PowerUpMax-PowerUpMin)/Dt)+2)
	pu := d.PowerUp
	if pu == nil {
		t.Fatalf("no power-up after %v s", PowerUpMax)
	}
	if !d.Level.hasRoom(pu.X, pu.Y, powerUpRoom) {
		t.Errorf("power-up at (%.0f,%.0f) has no room around it", pu.X, pu.Y)
	}
	run3(d, int(PowerUpLife/Dt)+1)
	if d.PowerUp == pu {
		t.Errorf("the power-up should disappear after %v s", PowerUpLife)
	}
}

func TestShieldBlocksHitsAndRunsOut(t *testing.T) {
	d := duelArena()
	target(d) // player 1 in the air
	fly(d.Players[0].World, 500, 200, 0, 0, 0)
	d.PowerUp = &PowerUp{X: 500, Y: 300, Kind: Shield, Life: PowerUpLife}
	run3(d, 1)
	p := &d.Players[1]
	if d.PowerUp != nil || p.ShieldLeft <= 0 {
		t.Fatalf("power-up %v, shield %.1f: player 1 should have taken the shield", d.PowerUp, p.ShieldLeft)
	}
	shoot(d)
	run3(d, 1)
	if p.Health != 1 {
		t.Errorf("health %.2f: the shield should block the hit", p.Health)
	}
	p.ShieldLeft = Dt / 2 // about to run out
	run3(d, 1)
	shoot(d)
	run3(d, 1)
	if p.Health == 1 {
		t.Error("after the shield runs out, hits should hurt again")
	}
}

func TestOneShotDestroysWithOneHit(t *testing.T) {
	d := duelArena()
	fly(d.Players[0].World, 300, 200, 0, 0, 0)
	placeAt(d, OneShot)
	run3(d, 1)
	if d.Players[0].OneShotLeft <= 0 {
		t.Fatal("player 0 should have taken the one-shot power-up")
	}
	target(d)
	shoot(d) // from player 0
	run3(d, 1)
	if d.Players[1].Alive() || d.Players[0].Score != 1 {
		t.Errorf("alive %v, score %d: one hit should destroy the ship", d.Players[1].Alive(), d.Players[0].Score)
	}
}

func TestPowerUpsAreLostOnDeath(t *testing.T) {
	d := duelArena()
	p := &d.Players[1]
	p.ShieldLeft, p.OneShotLeft = PowerUpTime, PowerUpTime
	p.wreck()
	run3(d, int(RespawnTime/Dt)+1)
	if p.ShieldLeft != 0 || p.OneShotLeft != 0 {
		t.Errorf("shield %.1f, one-shot %.1f: a new ship starts without power-ups", p.ShieldLeft, p.OneShotLeft)
	}
}
