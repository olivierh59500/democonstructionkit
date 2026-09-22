package presets

import "github.com/olivierh59500/democonstructionkit/composite"

// RibbonCurves provides the shared progression from straight to sinusoidal,
// harmonic and alternating split-row distortion, followed by three backdrop
// waves. Rate changes the sampling step, not the completed forward drift.
func RibbonCurves(rate float64) ([][]int, error) {
	definitions := []composite.DeltaCurve{
		{Step: 2.25},
		{Step: .20, Drift: 140, Terms: []composite.CurveTerm{{Amplitude: 100, Frequency: 1}}},
		{Step: .25, Drift: 175, Terms: []composite.CurveTerm{{Amplitude: 110, Frequency: 1}}},
		{Step: .30, Drift: 210, Terms: []composite.CurveTerm{{Amplitude: 120, Frequency: 1}}},
		{Step: .12, Drift: 175, Terms: []composite.CurveTerm{{Amplitude: 100, Frequency: 1}, {Amplitude: 25, Frequency: 10}}},
		{Step: .16, Drift: 210, Terms: []composite.CurveTerm{{Amplitude: 110, Frequency: 1}, {Amplitude: 27.5, Frequency: 9}}},
		{Step: .20, Drift: 245, Terms: []composite.CurveTerm{{Amplitude: 120, Frequency: 1}, {Amplitude: 30, Frequency: 8}}},
		{Step: .18, Extent: 720, Terms: []composite.CurveTerm{{Amplitude: 90, Frequency: 1}, {Amplitude: 12, Frequency: 3, Alternate: true, Attack: 160, Release: 160}}},
		{Step: .5, Terms: []composite.CurveTerm{{Amplitude: -60, Frequency: 1}}},
		{Step: .8, Terms: []composite.CurveTerm{{Amplitude: -60, Frequency: 1}}},
		{Step: .5, Terms: []composite.CurveTerm{{Amplitude: -60, Frequency: 1}, {Amplitude: -15, Frequency: 4}}},
	}
	result := make([][]int, len(definitions))
	for i, c := range definitions {
		if c.Extent == 0 {
			c.Extent = 360
		}
		c.Step *= rate
		c.Degrees = true
		c.OmitLastStep = true
		var err error
		result[i], err = c.Compile()
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
