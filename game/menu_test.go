package game

import (
	"testing"

	"eldfari/sim"
)

func TestMenuMoveWrapsAndRemembersLevelPerMode(t *testing.T) {
	var m menu
	counts := []int{7, 3} // crystals, cargo

	m.move(0, -1, counts) // up from the first level: wraps to the last
	if mode, lvl := m.selected(); mode != sim.ModeCrystals || lvl != 6 {
		t.Fatalf("got %v level %d, want crystals level 6", mode, lvl)
	}

	m.move(1, 0, counts) // to cargo
	m.move(0, 4, counts) // 4 down in 3 levels: wraps to level 1
	if mode, lvl := m.selected(); mode != sim.ModeCargo || lvl != 1 {
		t.Fatalf("got %v level %d, want cargo level 1", mode, lvl)
	}

	m.move(1, 0, counts) // right from the last mode: wraps to crystals
	if mode, lvl := m.selected(); mode != sim.ModeCrystals || lvl != 6 {
		t.Errorf("got %v level %d, want crystals level 6 (remembered)", mode, lvl)
	}
}
