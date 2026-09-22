package plasma

import "fmt"

// Wave describes a nested lookup wave. For an output pixel (x, y), its
// contribution is ColorTable[(colorAddress + Displacement[lookupAddress]) & mask].
// The addresses are the respective offset + x*stepX + y*stepY + phase. The lookup
// address is shifted right by DisplacementFractionBits before table wrapping.
// A shift of one permits half-entry phases, including exact word-aligned tables.
// Address arithmetic wraps modulo 2^64; negative offsets and steps wrap correctly.
type Wave struct {
	Displacement []uint16

	ColorOffset, ColorStepX, ColorStepY                      int64
	DisplacementOffset, DisplacementStepX, DisplacementStepY int64
	DisplacementFractionBits                                 uint
}

// LookupConfig supplies immutable tables and spatial coefficients. Tables
// must have power-of-two lengths from 1 to 65536. Width is 1..1048576 and Waves
// contains 1..64 entries. The constructor copies all input slices.
type LookupConfig struct {
	Width      int
	ColorTable []byte
	Waves      []Wave
}

// Phase changes a wave's addresses without rebuilding its spatial tables.
// Phases are independent of time: callers can use seconds, ticks, music events,
// recorded trajectories or a paused editor playhead to determine them.
type Phase struct {
	Color, Displacement int64
}

// Rows selects destination rows; the same absolute row is used as the
// wave's y coordinate. Step zero means one. Step two selects interleaved fields.
// Other destination rows and any row padding remain unchanged.
type Rows struct {
	First, Count, Step int
}

type plasmaColumn struct {
	color, displacement uint64
}

type plasmaWave struct {
	displacement []uint16
	columns      []plasmaColumn
	mask         uint64
	fractionBits uint
	colorY       uint64
	lookupY      uint64
}

// Lookup compiles nested lookup waves into an allocation-free indexed
// renderer. Contributions add modulo 256; palette expansion is deliberately
// separate, so the output can be a background, mask, sprite or animated texture.
// No trigonometry, GPU readback or mutable frame state is used by Render. A
// compiled renderer may be shared by goroutines drawing to separate buffers.
type Lookup struct {
	width  int
	colors []byte
	mask   uint64
	waves  []plasmaWave
}

func NewLookup(cfg LookupConfig) (*Lookup, error) {
	if cfg.Width < 1 || cfg.Width > 1<<20 || len(cfg.Waves) < 1 || len(cfg.Waves) > 64 {
		return nil, fmt.Errorf("plasma: invalid plasma width or wave count")
	}
	if !plasmaTableSize(len(cfg.ColorTable)) {
		return nil, fmt.Errorf("plasma: plasma color table must have a power-of-two length from 1 to 65536")
	}
	for i, wave := range cfg.Waves {
		if !plasmaTableSize(len(wave.Displacement)) || wave.DisplacementFractionBits > 16 {
			return nil, fmt.Errorf("plasma: invalid plasma displacement table or fractional bits for wave %d", i)
		}
	}
	p := &Lookup{
		width: cfg.Width, colors: append([]byte(nil), cfg.ColorTable...),
		mask: uint64(len(cfg.ColorTable) - 1), waves: make([]plasmaWave, len(cfg.Waves)),
	}
	for i, wave := range cfg.Waves {
		compiled := plasmaWave{
			displacement: append([]uint16(nil), wave.Displacement...),
			columns:      make([]plasmaColumn, cfg.Width), mask: uint64(len(wave.Displacement) - 1),
			fractionBits: wave.DisplacementFractionBits,
			colorY:       uint64(wave.ColorStepY), lookupY: uint64(wave.DisplacementStepY),
		}
		for x := range compiled.columns {
			compiled.columns[x] = plasmaColumn{
				color:        uint64(wave.ColorOffset) + uint64(x)*uint64(wave.ColorStepX),
				displacement: uint64(wave.DisplacementOffset) + uint64(x)*uint64(wave.DisplacementStepX),
			}
		}
		p.waves[i] = compiled
	}
	return p, nil
}

func plasmaTableSize(n int) bool { return n >= 1 && n <= 65536 && n&(n-1) == 0 }

// Width returns the number of indices written to each selected row.
func (p *Lookup) Width() int { return p.width }

// Render writes height consecutive rows into a caller-owned indexed buffer.
// A nil phases slice means zero phases; otherwise supply one phase per wave.
// Validation finishes before any output is changed.
func (p *Lookup) Render(dst []byte, stride, height int, phases []Phase) error {
	return p.RenderRows(dst, stride, Rows{Count: height}, phases)
}

// RenderRows writes selected rows without clearing the rest of the destination.
// The output must not alias a concurrently written buffer or the phases slice.
func (p *Lookup) RenderRows(dst []byte, stride int, rows Rows, phases []Phase) error {
	if p == nil || p.width < 1 || stride < p.width || rows.First < 0 || rows.Count < 0 || rows.Step < 0 {
		return fmt.Errorf("plasma: invalid plasma output layout")
	}
	if len(phases) != 0 && len(phases) != len(p.waves) {
		return fmt.Errorf("plasma: plasma phases must match the wave count")
	}
	if rows.Count == 0 {
		return nil
	}
	step := rows.Step
	if step == 0 {
		step = 1
	}
	if len(dst) < p.width {
		return fmt.Errorf("plasma: short plasma output buffer")
	}
	last := (len(dst) - p.width) / stride
	if rows.First > last || rows.Count-1 > (last-rows.First)/step {
		return fmt.Errorf("plasma: short plasma output buffer")
	}
	y := rows.First
	for row := 0; row < rows.Count; row++ {
		pixels := dst[y*stride : y*stride+p.width]
		clear(pixels)
		for i := range p.waves {
			wave := &p.waves[i]
			colorY := uint64(y) * wave.colorY
			lookupY := uint64(y) * wave.lookupY
			if len(phases) != 0 {
				colorY += uint64(phases[i].Color)
				lookupY += uint64(phases[i].Displacement)
			}
			for x, column := range wave.columns {
				address := ((column.displacement + lookupY) >> wave.fractionBits) & wave.mask
				displacement := uint64(wave.displacement[address])
				pixels[x] += p.colors[(column.color+colorY+displacement)&p.mask]
			}
		}
		y += step
	}
	return nil
}
