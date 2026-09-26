package recipes

import "github.com/olivierh59500/democonstructionkit/motion"

// CuddlyStarwarsProfile preserves twelve independently restarted sine segments
// and one deliberately unused tail sample that extends the strict wrap tick.
func CuddlyStarwarsProfile() motion.SegmentedProfileConfig {
	lengths := [...]int{100, 100, 50, 60, 30, 50, 30, 72, 60, 50, 30, 100}
	segments := make([]motion.ProfileSegment, len(lengths))
	for i, length := range lengths {
		segments[i] = motion.ProfileSegment{Samples: length, Amplitude: 50, Offset: .5}
	}
	segments[3].Rectify = true
	return motion.SegmentedProfileConfig{
		Segments: segments, AppendTail: true, TailValue: 0,
		Start: 20, Restart: 20, WrapMargin: 80, VisibleSamples: 20,
		Round: motion.ProfileHalfUp,
	}
}
