package plasma

import (
	"bytes"
	"math"
	"testing"
)

func TestHarmonicCustomWavesAndChannels(t *testing.T) {
	config := DefaultHarmonicConfig(19, 11)
	config.Waves = []HarmonicWave{
		{Shape: Radial, Frequency: .19, Speed: -2, Phase: .4, Amplitude: .7, CenterX: 8.5, CenterY: 3.25},
		{Shape: Horizontal, Frequency: -.07, Speed: .3, Phase: -1, Amplitude: -.2, CenterX: 1.25},
		{Shape: Diagonal, Frequency: .23, Speed: 1.1, Phase: 2.5, Amplitude: 1.3, CenterX: -5, CenterY: 2},
	}
	config.Divisor = 2.2
	config.MaximumColor = 233
	config.Channels[0] = HarmonicChannel{SinWeight: .5, CosWeight: .1, Offset: 1.25, Scale: 90}
	p, err := NewHarmonic(config)
	if err != nil {
		t.Fatal(err)
	}
	stride := config.Width*4 + 5
	dst := bytes.Repeat([]byte{0xa7}, stride*config.Height)
	for _, seconds := range []float64{0, -.7, 1.25, 35} {
		if err := p.RenderRGBA(dst, stride, seconds); err != nil {
			t.Fatal(err)
		}
		for y := 0; y < config.Height; y++ {
			for x := 0; x < config.Width; x++ {
				value := 0.0
				for _, wave := range config.Waves {
					coordinate := float64(x) - wave.CenterX
					switch wave.Shape {
					case Diagonal:
						coordinate = float64(x+y) - wave.CenterX - wave.CenterY
					case Radial:
						dx, dy := float64(x)-wave.CenterX, float64(y)-wave.CenterY
						coordinate = math.Sqrt(dx*dx + dy*dy)
					}
					value += wave.Amplitude * math.Sin(coordinate*wave.Frequency+seconds*wave.Speed+wave.Phase)
				}
				sine, cosine := math.Sincos(value / config.Divisor * config.ColorFrequency)
				for channel, curve := range config.Channels {
					color := (curve.SinWeight*sine + curve.CosWeight*cosine + curve.Offset) * curve.Scale
					want := byte(math.Max(0, math.Min(float64(config.MaximumColor), color)))
					got := dst[y*stride+x*4+channel]
					// Direct sin(a+b) and cached angle addition can differ by one
					// quantization unit exactly at an integer color boundary.
					if math.Abs(float64(got)-float64(want)) > 1 {
						t.Fatalf("(%d,%d) time %g channel %d: %d, want %d", x, y, seconds, channel, got, want)
					}
				}
				if dst[y*stride+x*4+3] != 255 {
					t.Fatal("nonopaque plasma output")
				}
			}
			if !bytes.Equal(dst[y*stride+config.Width*4:(y+1)*stride], bytes.Repeat([]byte{0xa7}, 5)) {
				t.Fatal("modified row padding")
			}
		}
	}
	before := append([]byte(nil), dst...)
	config.Waves[0].Frequency = 100
	if err := p.RenderRGBA(dst, stride, 35); err != nil || !bytes.Equal(dst, before) {
		t.Fatal("compiled harmonic retains configuration slice")
	}
	if allocations := testing.AllocsPerRun(20, func() {
		if err := p.RenderRGBA(dst, stride, 35); err != nil {
			panic(err)
		}
	}); allocations != 0 {
		t.Fatalf("harmonic render allocations = %g", allocations)
	}
}

func TestHarmonicRejectsInvalidInputsBeforeDrawing(t *testing.T) {
	config := DefaultHarmonicConfig(4, 3)
	p, err := NewHarmonic(config)
	if err != nil {
		t.Fatal(err)
	}
	dst := bytes.Repeat([]byte{0xa7}, 48)
	for _, seconds := range []float64{math.NaN(), math.Inf(1), math.MaxFloat64} {
		if err := p.RenderRGBA(dst, 16, seconds); err == nil {
			t.Fatal("accepted invalid time")
		}
		if !bytes.Equal(dst, bytes.Repeat([]byte{0xa7}, 48)) {
			t.Fatal("invalid input partially changed output")
		}
	}
	if err := p.RenderRGBA(dst, 15, 0); err == nil {
		t.Fatal("accepted short stride")
	}
	for _, change := range []func(*HarmonicConfig){
		func(c *HarmonicConfig) { c.Width = 0 },
		func(c *HarmonicConfig) { c.Divisor = 0 },
		func(c *HarmonicConfig) { c.Waves = nil },
		func(c *HarmonicConfig) { c.Waves[0].Phase = math.NaN() },
		func(c *HarmonicConfig) { c.Channels[1].Scale = math.Inf(1) },
	} {
		candidate := DefaultHarmonicConfig(4, 3)
		change(&candidate)
		if _, err := NewHarmonic(candidate); err == nil {
			t.Fatal("accepted invalid configuration")
		}
	}
}
