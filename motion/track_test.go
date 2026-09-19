package motion

import "testing"

func TestKeyframeClampInterpolationAndInputIsolation(t *testing.T) {
	keys := []Key[float64]{{Time: 1, Value: 10}, {Time: 3, Value: 30}, {Time: 4, Value: 0}}
	track, err := NewTrack(keys, Lerp)
	if err != nil {
		t.Fatal(err)
	}
	keys[0].Value = 999
	for _, tt := range []struct{ time, want float64 }{{-1, 10}, {2, 20}, {3, 30}, {3.5, 15}, {100, 0}} {
		if got := track.At(tt.time); got != tt.want {
			t.Fatal(tt, got)
		}
	}
	if _, err = NewTrack([]Key[float64]{{Time: 1}, {Time: 1}}, Lerp); err == nil {
		t.Fatal("accepted ambiguous key times")
	}
}
