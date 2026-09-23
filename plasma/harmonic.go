// Package plasma provides reusable procedural plasma kernels with independent
// spatial waves, animation phases and color presentation.
package plasma

import (
	"fmt"
	"math"
)

// HarmonicShape determines the coordinate used to sample a sine wave.
type HarmonicShape uint8

const (
	Horizontal HarmonicShape = iota // x - CenterX
	Vertical                        // y - CenterY
	Diagonal                        // x + y - CenterX - CenterY
	Radial                          // distance from (CenterX, CenterY)
)

// HarmonicWave produces Amplitude*sin(coordinate*Frequency + time*Speed + Phase).
// Speed and Phase are radians per second and radians respectively. Coordinates
// are output pixels. Negative amplitudes/frequencies/speeds are supported.
type HarmonicWave struct {
	Shape                       HarmonicShape
	Frequency, Speed, Phase     float64
	Amplitude, CenterX, CenterY float64
}

// HarmonicChannel maps the combined field to an RGB channel:
// (SinWeight*sin(value) + CosWeight*cos(value) + Offset) * Scale.
// value is the wave sum divided by Divisor and multiplied by ColorFrequency.
type HarmonicChannel struct {
	SinWeight, CosWeight, Offset, Scale float64
}

type HarmonicConfig struct {
	Width, Height  int
	Waves          []HarmonicWave
	Divisor        float64
	ColorFrequency float64
	Channels       [3]HarmonicChannel
	MaximumColor   byte
	// ColorLookupSize selects a bounded approximate RGB table for the standard
	// four-wave kernel. Zero preserves exact per-pixel trigonometry. Use a power
	// of two from 256 to 65536 for mobile CPU savings.
	ColorLookupSize int
}

// DefaultHarmonicConfig is a four-wave RGB plasma using axial, radial and
// diagonal waves. Its channels are separated by one third of a color cycle.
func DefaultHarmonicConfig(width, height int) HarmonicConfig {
	const third = 0.8660254037844386
	return HarmonicConfig{
		Width: width, Height: height, Divisor: 4, ColorFrequency: math.Pi, MaximumColor: 254,
		Waves: []HarmonicWave{
			{Shape: Horizontal, Frequency: .02, Speed: 1, Amplitude: 1},
			{Shape: Vertical, Frequency: .03, Speed: 1.5, Amplitude: 1},
			{Shape: Radial, Frequency: .01, Speed: .5, Amplitude: 1},
			{Shape: Diagonal, Frequency: .01, Speed: 2, Amplitude: 1},
		},
		Channels: [3]HarmonicChannel{
			{SinWeight: 1, Offset: 1, Scale: 127},
			{SinWeight: -.5, CosWeight: third, Offset: 1, Scale: 127},
			{SinWeight: -.5, CosWeight: -third, Offset: 1, Scale: 127},
		},
	}
}

type harmonicWave struct {
	config           HarmonicWave
	sine, cosine     []float64
	animated         []float64
	sinTime, cosTime float64
}

// Harmonic caches spatial sine/cosine terms and reuses its axial wave buffers.
// Rendering allocates no memory and evaluates only temporal phases plus one
// color sine/cosine pair per pixel. Use separate instances for concurrent draws.
// The renderer owns no GPU resources; the caller uploads RGBA only when changed.
type Harmonic struct {
	config               HarmonicConfig
	waves                []harmonicWave
	fourWaveRGB          bool
	colors               []uint32
	colorMin, colorScale float64
}

func NewHarmonic(config HarmonicConfig) (*Harmonic, error) {
	if config.Width < 1 || config.Height < 1 || config.Width > (1<<24)/config.Height || len(config.Waves) < 1 || len(config.Waves) > 64 {
		return nil, fmt.Errorf("plasma: invalid harmonic dimensions or wave count")
	}
	if !finite(config.Divisor) || config.Divisor == 0 || !finite(config.ColorFrequency) {
		return nil, fmt.Errorf("plasma: invalid harmonic normalization")
	}
	if config.ColorLookupSize != 0 && (config.ColorLookupSize < 256 || config.ColorLookupSize > 65536 || config.ColorLookupSize&(config.ColorLookupSize-1) != 0) {
		return nil, fmt.Errorf("plasma: color lookup size must be a power of two from 256 to 65536")
	}
	amplitudeBound := 0.0
	for _, wave := range config.Waves {
		if wave.Shape > Radial || !finite(wave.Frequency) || !finite(wave.Speed) || !finite(wave.Phase) ||
			!finite(wave.Amplitude) || !finite(wave.CenterX) || !finite(wave.CenterY) {
			return nil, fmt.Errorf("plasma: invalid harmonic wave")
		}
		amplitudeBound += math.Abs(wave.Amplitude)
	}
	if !finite(amplitudeBound) || !finite(amplitudeBound/math.Abs(config.Divisor)*math.Abs(config.ColorFrequency)) {
		return nil, fmt.Errorf("plasma: harmonic amplitudes overflow the color phase")
	}
	for _, channel := range config.Channels {
		if !finite(channel.SinWeight) || !finite(channel.CosWeight) || !finite(channel.Offset) || !finite(channel.Scale) {
			return nil, fmt.Errorf("plasma: invalid harmonic channel")
		}
		bound := math.Abs(channel.SinWeight) + math.Abs(channel.CosWeight) + math.Abs(channel.Offset)
		if !finite(bound) || !finite(bound*math.Abs(channel.Scale)) {
			return nil, fmt.Errorf("plasma: harmonic channel overflows")
		}
	}
	config.Waves = append([]HarmonicWave(nil), config.Waves...)
	p := &Harmonic{config: config, waves: make([]harmonicWave, len(config.Waves))}
	// Compile the common four-wave/color combination to a branch-free pixel
	// loop. Frequencies, phases, centers, speeds and normalization remain free.
	standard := DefaultHarmonicConfig(1, 1)
	if len(config.Waves) == 4 && config.Channels == standard.Channels && config.MaximumColor == 254 {
		p.fourWaveRGB = true
		for i, shape := range [4]HarmonicShape{Horizontal, Vertical, Radial, Diagonal} {
			if config.Waves[i].Shape != shape || config.Waves[i].Amplitude != 1 {
				p.fourWaveRGB = false
			}
		}
	}
	if config.ColorLookupSize != 0 {
		if !p.fourWaveRGB {
			return nil, fmt.Errorf("plasma: color lookup requires the standard four-wave RGB kernel")
		}
		bound := amplitudeBound / math.Abs(config.Divisor) * math.Abs(config.ColorFrequency)
		if bound == 0 {
			bound = 1
		}
		if !finite(2*bound) || !finite(float64(config.ColorLookupSize-1)/(2*bound)) || float64(config.ColorLookupSize-1)/(2*bound) <= 0 {
			return nil, fmt.Errorf("plasma: color lookup range overflows")
		}
		p.colors = make([]uint32, config.ColorLookupSize)
		p.colorMin = -bound
		p.colorScale = float64(len(p.colors)-1) / (2 * bound)
		const third = 0.8660254037844386
		for i := range p.colors {
			phase := p.colorMin + float64(i)/p.colorScale
			sine, cosine := math.Sincos(phase)
			r := harmonicColor(sine)
			g := harmonicColor(-.5*sine + third*cosine)
			b := harmonicColor(-.5*sine - third*cosine)
			p.colors[i] = uint32(r) | uint32(g)<<8 | uint32(b)<<16 | 0xff000000
		}
	}

	for i, wave := range config.Waves {
		count := config.Width * config.Height
		switch wave.Shape {
		case Horizontal:
			count = config.Width
		case Vertical:
			count = config.Height
		case Diagonal:
			count = config.Width + config.Height - 1
		}
		compiled := harmonicWave{config: wave, sine: make([]float64, count), cosine: make([]float64, count)}
		if wave.Shape != Radial {
			compiled.animated = make([]float64, count)
		}
		for n := 0; n < count; n++ {
			coordinate := float64(n)
			switch wave.Shape {
			case Horizontal:
				coordinate -= wave.CenterX
			case Vertical:
				coordinate -= wave.CenterY
			case Diagonal:
				coordinate -= wave.CenterX + wave.CenterY
			case Radial:
				x, y := float64(n%config.Width)-wave.CenterX, float64(n/config.Width)-wave.CenterY
				coordinate = math.Sqrt(x*x + y*y)
			}
			phase := coordinate * wave.Frequency
			if !finite(phase) {
				return nil, fmt.Errorf("plasma: harmonic spatial phase overflows")
			}
			compiled.sine[n], compiled.cosine[n] = math.Sincos(phase)
		}
		p.waves[i] = compiled
	}
	return p, nil
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

// RenderRGBA writes opaque RGBA pixels into caller-owned storage. Row padding is
// untouched. Time is expressed in the unit used by each wave's Speed. Invalid
// arguments are rejected before any destination bytes are changed.
func (p *Harmonic) RenderRGBA(dst []byte, stride int, seconds float64) error {
	if p == nil || p.config.Width < 1 || p.config.Height < 1 || !finite(seconds) || stride < p.config.Width*4 || len(dst) < p.config.Width*4 ||
		p.config.Height-1 > (len(dst)-p.config.Width*4)/stride {
		return fmt.Errorf("plasma: invalid harmonic destination or time")
	}
	for i := range p.waves {
		wave := &p.waves[i]
		phase := seconds*wave.config.Speed + wave.config.Phase
		if !finite(phase) {
			return fmt.Errorf("plasma: harmonic time overflows a wave phase")
		}
	}
	for i := range p.waves {
		wave := &p.waves[i]
		wave.sinTime, wave.cosTime = math.Sincos(seconds*wave.config.Speed + wave.config.Phase)
		for n := range wave.animated {
			wave.animated[n] = wave.sine[n]*wave.cosTime + wave.cosine[n]*wave.sinTime
		}
	}
	if p.fourWaveRGB {
		if len(p.colors) != 0 {
			p.renderFourRGBAFast(dst, stride)
		} else {
			p.renderFourRGBA(dst, stride)
		}
		return nil
	}
	for y := 0; y < p.config.Height; y++ {
		for x := 0; x < p.config.Width; x++ {
			index := y*p.config.Width + x
			value := 0.0
			for i := range p.waves {
				wave := &p.waves[i]
				var contribution float64
				switch wave.config.Shape {
				case Horizontal:
					contribution = wave.animated[x]
				case Vertical:
					contribution = wave.animated[y]
				case Diagonal:
					contribution = wave.animated[x+y]
				case Radial:
					contribution = wave.sine[index]*wave.cosTime + wave.cosine[index]*wave.sinTime
				}
				value += contribution * wave.config.Amplitude
			}
			sinColor, cosColor := math.Sincos(value / p.config.Divisor * p.config.ColorFrequency)
			pixel := dst[y*stride+x*4 : y*stride+x*4+4]
			for i, channel := range p.config.Channels {
				value := (sinColor*channel.SinWeight + cosColor*channel.CosWeight + channel.Offset) * channel.Scale
				if value <= 0 {
					pixel[i] = 0
				} else if value >= float64(p.config.MaximumColor) {
					pixel[i] = p.config.MaximumColor
				} else {
					pixel[i] = byte(value)
				}
			}
			pixel[3] = 255
		}
	}
	return nil
}

func (p *Harmonic) renderFourRGBA(dst []byte, stride int) {
	xWave, yWave := p.waves[0].animated, p.waves[1].animated
	radial, diagonal := &p.waves[2], p.waves[3].animated
	width, height := p.config.Width, p.config.Height
	divisor, colorFrequency := p.config.Divisor, p.config.ColorFrequency
	const third = 0.8660254037844386
	for y := 0; y < height; y++ {
		row := y * width
		for x := 0; x < width; x++ {
			index := row + x
			r := radial.sine[index]*radial.cosTime + radial.cosine[index]*radial.sinTime
			value := (xWave[x] + yWave[y] + r + diagonal[x+y]) / divisor
			s, c := math.Sincos(value * colorFrequency)
			pixel := y*stride + x*4
			dst[pixel] = harmonicColor(s)
			dst[pixel+1] = harmonicColor(-.5*s + third*c)
			dst[pixel+2] = harmonicColor(-.5*s - third*c)
			dst[pixel+3] = 255
		}
	}
}

// renderFourRGBAFast keeps the exact spatial/temporal wave recurrence and
// substitutes only the three final color sinusoids with a bounded RGB lookup.
func (p *Harmonic) renderFourRGBAFast(dst []byte, stride int) {
	xWave, yWave := p.waves[0].animated, p.waves[1].animated
	radial, diagonal := &p.waves[2], p.waves[3].animated
	width, height := p.config.Width, p.config.Height
	divisor, colorFrequency := p.config.Divisor, p.config.ColorFrequency
	for y := 0; y < height; y++ {
		row := y * width
		for x := 0; x < width; x++ {
			index := row + x
			r := radial.sine[index]*radial.cosTime + radial.cosine[index]*radial.sinTime
			value := (xWave[x] + yWave[y] + r + diagonal[x+y]) / divisor
			lookup := int((value*colorFrequency - p.colorMin) * p.colorScale)
			lookup = max(0, min(lookup, len(p.colors)-1))
			color := p.colors[lookup]
			pixel := y*stride + x*4
			dst[pixel] = byte(color)
			dst[pixel+1] = byte(color >> 8)
			dst[pixel+2] = byte(color >> 16)
			dst[pixel+3] = 255
		}
	}
}

func harmonicColor(value float64) byte {
	value = (value + 1) * 127
	if value <= 0 {
		return 0
	}
	if value >= 254 {
		return 254
	}
	return byte(value)
}
