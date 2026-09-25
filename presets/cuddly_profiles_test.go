package presets

import (
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestCuddlyWaveProfilesKeepOverlapsAndWrapLengths(t *testing.T) {
	digi, err := motion.CompileWaveProgram(CuddlyDigiWaveProgram()...)
	if err != nil {
		t.Fatal(err)
	}
	intro, err := motion.CompileWaveProgram(CuddlyIntroWaveProgram()...)
	if err != nil {
		t.Fatal(err)
	}
	ehhh, err := CuddlyEhhhProfile()
	if err != nil {
		t.Fatal(err)
	}
	if len(digi) != 3040 || len(intro) != 1533 || len(ehhh) != 2271 {
		t.Fatalf("profile lengths: Digi %d, intro %d, Ehhh %d", len(digi), len(intro), len(ehhh))
	}
	checks := []struct {
		name   string
		values []float64
		index  int
		want   float64
	}{
		{"Digi first hold", digi, 282, 40},
		{"Digi overlapping negative hold", digi, 390, -40},
		{"Digi crest end", digi, 510, 40 * math.Sin(9.6+148*.02)},
		{"Digi next section", digi, 511, 0},
		{"Digi short perturbation", digi, 511 + 252 + 504 + 315 + 630 + 200, 40*math.Sin(200*.01) + 2*math.Sin(200)},
		{"intro blank lead", intro, 251, 0},
		{"intro first crest", intro, 504 + 29, 200 * math.Sin(29*.05)},
		{"intro overlapping negative hold", intro, 504 + 138, -200},
		{"intro final crest", intro, 1532, 40 * math.Sin(251*.05)},
		{"Ehhh prefix copy", ehhh, 250 + 390, -20},
		{"Ehhh tail second segment", ehhh, 250 + 763 + 630 + 200, 20*math.Sin(200*.01) + 2*math.Sin(200)},
	}
	for _, check := range checks {
		if got := check.values[check.index]; math.Abs(got-check.want) > 1e-12 {
			t.Errorf("%s at %d = %.17g, want %.17g", check.name, check.index, got, check.want)
		}
	}
	// Presets return editable, independent data instead of sharing mutable terms.
	a, b := CuddlyDigiWaveProgram(), CuddlyDigiWaveProgram()
	a[0].Section.Terms[0].Amplitude = 999
	if b[0].Section.Terms[0].Amplitude != 40 {
		t.Fatal("Digi preset terms alias")
	}
}
