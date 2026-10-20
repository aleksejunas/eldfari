package game

import (
	"reflect"
	"testing"
)

func TestMigrateOldKeys(t *testing.T) {
	times := map[string]float64{
		"Grotta":          40, // old key: moves to crystals/
		"Dypet":           90, // old key, but a newer record exists: dropped
		"crystals/Dypet":  80,
		"cargo/Something": 30, // already has a mode: untouched
	}
	migrateOldKeys(times)
	want := map[string]float64{
		"crystals/Grotta": 40,
		"crystals/Dypet":  80,
		"cargo/Something": 30,
	}
	if !reflect.DeepEqual(times, want) {
		t.Errorf("got %v, want %v", times, want)
	}
}
