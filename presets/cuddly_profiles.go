package presets

import "github.com/olivierh59500/democonstructionkit/motion"

// CuddlyDigiWaveProgram is the editable lookup recipe for the Digi logo.
// Absolute write positions preserve its overlapping source-table sections.
func CuddlyDigiWaveProgram() []motion.WaveWrite {
	sine := func(at, samples, start int, amplitude, step, phase float64) motion.WaveWrite {
		return motion.WaveWrite{At: at, Section: motion.WaveSection{Samples: samples, SampleStart: start,
			Terms: []motion.WaveTerm{{Amplitude: amplitude, Step: step, Phase: phase}}}}
	}
	hold := func(at, samples int, value float64) motion.WaveWrite {
		return motion.WaveWrite{At: at, Section: motion.WaveSection{Samples: samples, Offset: value}}
	}
	double := func(at, samples, start int, amplitude float64) motion.WaveWrite {
		return motion.WaveWrite{At: at, Section: motion.WaveSection{Samples: samples, SampleStart: start,
			Terms: []motion.WaveTerm{{Amplitude: amplitude, Step: .01}, {Amplitude: 2, Step: 1}}}}
	}
	return []motion.WaveWrite{
		sine(0, 252, 0, 40, .05, 0),
		sine(252, 30, 0, 40, .05, 0),
		hold(282, 50, 40),
		sine(332, 60, 30, 40, .05, 0),
		hold(390, 62, -40), // The previous two samples are deliberately replaced.
		sine(452, 59, 90, 40, .02, 9.6),
		sine(motion.WaveAppend, 252, 0, 40, .05, 0),
		sine(motion.WaveAppend, 504, 0, 40, .025, 0),
		{At: motion.WaveAppend, Section: motion.WaveSection{Samples: 315, Terms: []motion.WaveTerm{{Amplitude: 40, Step: .02}, {Amplitude: 5, Step: .5}}}},
		sine(motion.WaveAppend, 630, 0, 40, .01, 0),
		double(motion.WaveAppend, 100, 0, 40),
		sine(motion.WaveAppend, 100, 100, 40, .01, 0),
		double(motion.WaveAppend, 100, 200, 40),
		sine(motion.WaveAppend, 328, 300, 40, .01, 0),
		sine(motion.WaveAppend, 200, 0, 40, .05, 0),
	}
}

// CuddlyIntroWaveProgram uses zero holds and three repeated, overlapping
// crest sections. Every call returns independent editable section data.
func CuddlyIntroWaveProgram() []motion.WaveWrite {
	sine := func(at, samples, start int, amplitude, step, phase float64) motion.WaveWrite {
		return motion.WaveWrite{At: at, Section: motion.WaveSection{Samples: samples, SampleStart: start,
			Terms: []motion.WaveTerm{{Amplitude: amplitude, Step: step, Phase: phase}}}}
	}
	program := []motion.WaveWrite{
		{At: 0, Section: motion.WaveSection{Samples: 252}},
		sine(motion.WaveAppend, 252, 0, 80, .05, 0),
	}
	for repeat := 0; repeat < 3; repeat++ {
		base := 504 + repeat*259
		program = append(program,
			motion.WaveWrite{At: motion.WaveAppend, Section: motion.WaveSection{Samples: 259}},
			sine(base, 30, 0, 200, .05, 0),
			motion.WaveWrite{At: base + 30, Section: motion.WaveSection{Samples: 50, Offset: 200}},
			sine(base+80, 60, 30, 200, .05, 0),
			motion.WaveWrite{At: base + 138, Section: motion.WaveSection{Samples: 62, Offset: -200}},
			sine(base+200, 59, 90, 220, .02, 9.6),
		)
	}
	return append(program, sine(motion.WaveAppend, 252, 0, 40, .05, 0))
}

// CuddlyEhhhProfile reuses the first 763 Digi samples at half amplitude,
// followed by its own editable late wave. It retains the original 250-tick
// blank lead-in used by the row-profile renderer.
func CuddlyEhhhProfile() ([]float64, error) {
	digi, err := motion.CompileWaveProgram(CuddlyDigiWaveProgram()...)
	if err != nil {
		return nil, err
	}
	tail, err := motion.CompileWaveProgram(CuddlyEhhhTailWaveProgram()...)
	if err != nil {
		return nil, err
	}
	values := make([]float64, 250, 250+763+len(tail))
	for _, value := range digi[:763] {
		values = append(values, value/2)
	}
	return append(values, tail...), nil
}

// CuddlyEhhhTailWaveProgram describes the two final 20-pixel sine sections.
func CuddlyEhhhTailWaveProgram() []motion.WaveWrite {
	sine := func(samples, start int, terms ...motion.WaveTerm) motion.WaveWrite {
		return motion.WaveWrite{At: motion.WaveAppend, Section: motion.WaveSection{Samples: samples, SampleStart: start, Terms: terms}}
	}
	return []motion.WaveWrite{
		sine(630, 0, motion.WaveTerm{Amplitude: 20, Step: .01}),
		sine(100, 0, motion.WaveTerm{Amplitude: 20, Step: .01}, motion.WaveTerm{Amplitude: 2, Step: 1}),
		sine(100, 100, motion.WaveTerm{Amplitude: 20, Step: .01}),
		sine(100, 200, motion.WaveTerm{Amplitude: 20, Step: .01}, motion.WaveTerm{Amplitude: 2, Step: 1}),
		sine(328, 300, motion.WaveTerm{Amplitude: 20, Step: .01}),
	}
}
