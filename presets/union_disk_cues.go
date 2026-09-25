package presets

import (
	"math"

	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/timeline"
)

// UnionDiskCopierCueRanges keeps the authored gaps and strict open boundaries
// of the copy presentation. The caller supplies messages and image assets.
func UnionDiskCopierCueRanges() []timeline.CueRange {
	return []timeline.CueRange{
		{Start: 0, End: 100},
		{Start: 120, End: 360, OpenStart: true},
		{Start: 380, End: 480, OpenStart: true},
		{Start: 500, End: 740, OpenStart: true},
		{Start: 760, End: 1000, OpenStart: true},
		{Start: 1000, End: 1100, OpenStart: true},
		{Start: 1120, End: math.Inf(1), OpenStart: true},
	}
}

// UnionDiskCopierFade selects one of eight editable raster/color bank images.
func UnionDiskCopierFade() timeline.SteppedEnvelopeConfig {
	return timeline.SteppedEnvelopeConfig{MaxIndex: 7, EntryLength: 14, ExitLead: 16, Step: 2}
}

// UnionDiskCopierLCDMotion starts three fractional tile clocks at their
// independent operation cues, then wraps strictly after source frame 82.
func UnionDiskCopierLCDMotion() motion.GatedWrapBankConfig {
	return motion.GatedWrapBankConfig{
		Start: []float64{0, 0, 0}, Velocity: []float64{.35}, Gates: []float64{120, 500, 760},
		Upper: &motion.WrapLimit{Boundary: 82, Restart: 0},
	}
}
