package sim

import (
	"testing"

	"eldfari/levels"
)

// An open arena with a floor at y=500, a start pad for each player and three
// crystals.
func raceArena() *Race {
	floor := []Pt{{0, 500}, {1000, 500}, {1000, 600}, {0, 600}}
	l := NewLevel(1000, 600, [][]Pt{floor}, []Pad{{50, 150, 500}, {850, 950, 500}}, Pt{100, 500 - ShipRear - 0.5})
	l.Mode = ModeRace
	l.Crystals = []Pt{{300, 300}, {500, 300}, {700, 300}}
	return NewRace(l)
}

func step2(r *Race, frames int) {
	for range frames {
		r.Step([2]Input{})
	}
}

func TestShippedRaceLevelsLoad(t *testing.T) {
	ls, err := LoadDir(levels.Race, ModeRace)
	if err != nil {
		t.Fatalf("the race levels are not valid:\n%v", err)
	}
	for _, l := range ls {
		if len(l.Crystals)%2 == 0 {
			t.Errorf("%s: %d crystals; use an odd number so a race can't end in a draw", l.Name, len(l.Crystals))
		}
		r := NewRace(l)
		step2(r, 60)
		if r.Players[0].Ship.State != Landed || r.Players[1].Ship.State != Landed || r.Winner >= 0 {
			t.Errorf("%s: after 1 s without input both should stand on their pads", l.Name)
		}
	}
}

func TestCrystalTakenByOneIsGoneForBoth(t *testing.T) {
	r := raceArena()
	fly(r.Players[0].World, 300, 300, 0, 0, 0)
	step2(r, 1)
	if r.Taken[0] != 0 || r.Players[0].Score != 1 {
		t.Fatalf("taken %v, score %d: player 1 should have the crystal", r.Taken, r.Players[0].Score)
	}
	fly(r.Players[1].World, 300, 300, 0, 0, 0)
	step2(r, 1)
	if r.Players[1].Score != 0 || r.Taken[0] != 0 {
		t.Errorf("player 2 took a crystal that was already gone")
	}
}

func TestMostCrystalsWinsWhenAllTakenAndLanded(t *testing.T) {
	r := raceArena()
	r.Taken = []int{0, 0, 1}
	r.Players[0].Score, r.Players[1].Score = 2, 1
	fly(r.Players[1].World, 900, 300, 0, 0, 0) // still in the air
	step2(r, 1)
	if r.Winner >= 0 {
		t.Fatal("the race should go on until both have landed")
	}
	r.Players[1].spawn() // back on its pad
	step2(r, 1)
	if r.Winner != 0 {
		t.Errorf("winner %d, want player 0 with 2 crystals", r.Winner)
	}
}

func TestRaceNearestCrystalSkipsTaken(t *testing.T) {
	r := raceArena()
	fly(r.Players[1].World, 520, 300, 0, 0, 0)
	if c, ok := r.NearestCrystal(1); !ok || c != r.Level.Crystals[1] {
		t.Fatalf("nearest = %v %v, want %v", c, ok, r.Level.Crystals[1])
	}
	r.Taken[1] = 0
	if c, _ := r.NearestCrystal(1); c != r.Level.Crystals[2] {
		t.Errorf("nearest = %v, want %v once the middle one is taken", c, r.Level.Crystals[2])
	}
}

func TestEqualCrystalsIsADraw(t *testing.T) {
	r := raceArena()
	r.Level.Crystals = r.Level.Crystals[:2]
	r.Taken = []int{0, 1}
	r.Players[0].Score, r.Players[1].Score = 1, 1
	step2(r, 1)
	if r.Winner != RaceDraw {
		t.Errorf("winner %d, want a draw", r.Winner)
	}
}

func TestCrashedRacerStartsAgainOnOwnPad(t *testing.T) {
	r := raceArena()
	fly(r.Players[1].World, 500, 400, 0, 300, 0) // falling fast onto the floor
	step2(r, 30)
	if r.Players[1].Alive() {
		t.Fatal("the ship should have crashed")
	}
	step2(r, int(RespawnTime/Dt)+1)
	pad := r.Level.Pads[1]
	if s := r.Players[1].Ship; s.State != Landed || s.X != (pad.X0+pad.X1)/2 {
		t.Errorf("after %v s the ship should stand on its own pad: %+v", RespawnTime, s)
	}
	if r.Players[0].Score != 0 {
		t.Error("a crash gives no crystals to the other player")
	}
}
