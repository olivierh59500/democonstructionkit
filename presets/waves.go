package presets

import (
	"math"

	"github.com/olivierh59500/democonstructionkit/motion"
)

// BilizirWaveSections is the six-part horizontal deformation recipe. Modify
// amplitudes, lengths, frequencies or order before motion.CompileWaveTable;
// the resulting samples work with any image, scrolling text or sprite path.
func BilizirWaveSections() []motion.WaveSection {
	wide := func() []motion.WaveTerm {
		return []motion.WaveTerm{
			{Amplitude: 20, Step: 7.0 / 180.0 * math.Pi},
			{Amplitude: 30, Step: 3.0 / 180.0 * math.Pi, Cosine: true},
		}
	}
	return []motion.WaveSection{
		{Samples: 389, Terms: wide()},
		{Samples: 120, Terms: []motion.WaveTerm{{Amplitude: 4, Step: 72.0 / 180.0 * math.Pi}}},
		{Samples: 68, Terms: []motion.WaveTerm{{Amplitude: 40, Step: 8.0 / 180.0 * math.Pi}}},
		{Samples: 389, Terms: wide()},
		{Samples: 36, Terms: []motion.WaveTerm{{Amplitude: 4, Step: 72.0 / 180.0 * math.Pi}}},
		{Samples: 189, Terms: []motion.WaveTerm{{Amplitude: 30, Step: 8.0 / 180.0 * math.Pi}}},
	}
}

// TCBLogoWaveSections is a hold, two sinusoidal sections and a final hold.
// This is image-independent; no logo dimensions or atlas coordinates are stored.
func TCBLogoWaveSections() []motion.WaveSection {
	return []motion.WaveSection{
		{Samples: 40},
		{Samples: 804, Terms: []motion.WaveTerm{{Amplitude: 8, Step: .05, Phase: -2}}},
		{Samples: 810, Terms: []motion.WaveTerm{{Amplitude: 8, Step: .15}}},
		{Samples: 160},
	}
}
