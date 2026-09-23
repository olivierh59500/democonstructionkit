package motion

import (
	"fmt"
	"math"
)

// WaveTerm contributes Amplitude*sin(sample*Step+Phase), or cosine when Cosine
// is set. Step and Phase use radians. Keeping cosine explicit preserves exact
// authored samples without introducing a quarter-turn rounding difference.
type WaveTerm struct {
	Amplitude, Step, Phase float64
	Cosine                 bool
}

// WaveSection is one finite part of a lookup table. Its local sample index
// starts at zero. Empty Terms produces a hold at Offset. Sections may be
// reordered or repeated to program scanline, logo, raster or sprite motion.
type WaveSection struct {
	Samples int
	Offset  float64
	Terms   []WaveTerm
}

// CompileWaveTable precomputes a waveform once; the renderer can subsequently
// sample it without trigonometry or allocations. It returns caller-owned data.
func CompileWaveTable(sections ...WaveSection) ([]float64, error) {
	count := 0
	finite := func(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
	for _, section := range sections {
		if section.Samples < 0 || section.Samples > 1<<20-count || !finite(section.Offset) {
			return nil, fmt.Errorf("motion: invalid wave section")
		}
		for _, term := range section.Terms {
			if !finite(term.Amplitude) || !finite(term.Step) || !finite(term.Phase) {
				return nil, fmt.Errorf("motion: nonfinite wave term")
			}
		}
		count += section.Samples
	}
	if count == 0 {
		return nil, fmt.Errorf("motion: empty wave table")
	}
	values := make([]float64, 0, count)
	for _, section := range sections {
		for i := 0; i < section.Samples; i++ {
			value := section.Offset
			for _, term := range section.Terms {
				phase := float64(i)*term.Step + term.Phase
				var wave float64
				if term.Cosine {
					wave = math.Cos(phase)
				} else {
					wave = math.Sin(phase)
				}
				value += term.Amplitude * wave
			}
			if !finite(value) {
				return nil, fmt.Errorf("motion: wave sample overflow")
			}
			values = append(values, value)
		}
	}
	return values, nil
}
