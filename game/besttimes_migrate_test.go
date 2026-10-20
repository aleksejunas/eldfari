package game

import (
	"reflect"
	"testing"
)

func TestMigrateOldKeys(t *testing.T) {
	times := map[string]float64{
		"Grotta":          40, // no mode, old name: to crystals/, then renamed
		"Dypet":           90, // no mode, but a record under the new name exists: dropped
		"crystals/Djúp":   80,
		"cargo/Kranen":    25, // old name: renamed
		"cargo/Something": 30, // not renamed: untouched
	}
	migrateOldKeys(times)
	want := map[string]float64{
		"crystals/Hellir": 40,
		"crystals/Djúp":   80,
		"cargo/Vindáss":   25,
		"cargo/Something": 30,
	}
	if !reflect.DeepEqual(times, want) {
		t.Errorf("got %v, want %v", times, want)
	}
}
