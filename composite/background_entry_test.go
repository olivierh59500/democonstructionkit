package composite

import "testing"

func TestBackgroundEntrySuppressesPrecedingTilesUntilFirstImageArrives(t *testing.T) {
	for _, sample := range []struct {
		origin, minimum, period float64
		entry                   bool
		want                    float64
	}{
		{328, 0, 800, true, 0},
		{1, 0, 800, true, 0},
		{0, 0, 800, true, 800},
		{-3, 0, 800, true, 800},
		{328, 0, 800, false, 800},
		{5, 4, 32, true, 0},
		{4, 4, 32, true, 32},
	} {
		if got := backgroundEntryPeriod(sample.origin, sample.minimum, sample.period, sample.entry); got != sample.want {
			t.Fatalf("entry %+v gives period %v, want %v", sample, got, sample.want)
		}
	}
}
