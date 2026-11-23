package game

import "testing"

func TestNameEntry(t *testing.T) {
	e := newNameEntry("x", 40, "ROL")
	if e.name() != "ROL" {
		t.Fatalf("name %q, want the last initials ROL", e.name())
	}
	e.change(-1) // R -> Q
	e.move(1)
	e.change(13) // O -> B (wraps past Z)
	e.move(5)    // stays on the last letter
	e.change(1)  // L -> M
	if e.name() != "QBM" || e.pos != 2 {
		t.Errorf("name %q at %d, want QBM at 2", e.name(), e.pos)
	}

	e = newNameEntry("x", 40, "")
	if e.typed('k') || e.typed('1') || e.typed('a') {
		t.Fatal("not done before the third letter")
	}
	if !e.typed('i') || e.name() != "KAI" {
		t.Errorf("name %q, want KAI and done after three letters", e.name())
	}
}
