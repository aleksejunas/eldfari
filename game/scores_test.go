package game

import "testing"

func TestTop10KeepsTheFastest(t *testing.T) {
	s := newScores()
	const lvl = "crystals/Hellir"
	for i, tm := range []float64{50, 40, 60, 30, 45, 55, 35, 65, 70, 75} {
		if rank := s.add(lvl, "P"+string(rune('A'+i)), tm); rank < 0 {
			t.Fatalf("time %v did not get in while the list had room", tm)
		}
	}
	if s.best(lvl) != 30 || len(s.Top[lvl]) != topSize {
		t.Fatalf("best %v, %d entries", s.best(lvl), len(s.Top[lvl]))
	}
	if s.qualifies(lvl, 80) {
		t.Error("80 s is slower than all ten and should not qualify")
	}
	if rank := s.add(lvl, "NEW", 32); rank != 1 {
		t.Errorf("32 s got place %d, want 1 (second)", rank)
	}
	top := s.Top[lvl]
	if len(top) != topSize || top[len(top)-1].Time != 70 {
		t.Errorf("the slowest (75) should have dropped out: %+v", top)
	}
	if s.Last != "NEW" {
		t.Errorf("last initials %q, want NEW", s.Last)
	}
}

func TestEqualTimeGoesAfterTheOlderOne(t *testing.T) {
	s := newScores()
	s.add("x", "OLD", 40)
	if rank := s.add("x", "NEW", 40); rank != 1 {
		t.Errorf("an equal time got place %d, want 1 (after the older one)", rank)
	}
}
