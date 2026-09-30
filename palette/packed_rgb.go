package palette

import (
	"fmt"
	"image/color"
)

// PackedRGBConfig describes three non-overlapping channels in a 32-bit word.
// Bits/Shift use R,G,B order; channels have one through eight bits. Zero arrays
// select RGB12: four bits at shifts 8,4,0. RGB565 uses {5,6,5}/{11,5,0}.
type PackedRGBConfig struct {
	Bits, Shift [3]uint8
}

// PackedRGB decodes colors and performs signed channel-delta interpolation.
// Unused word bits are ignored. Its zero value is RGB12. Division truncates
// toward zero before adding the first channel, preserving classic packed-color
// fades in both directions rather than rounding a weighted unsigned sum.
type PackedRGB struct {
	shift [3]uint8
	mask  [3]uint32
}

func NewPackedRGB(c PackedRGBConfig) (PackedRGB, error) {
	if c == (PackedRGBConfig{}) {
		return PackedRGB{}, nil
	}
	var f PackedRGB
	used := uint32(0)
	for i, bits := range c.Bits {
		if bits < 1 || bits > 8 || int(bits)+int(c.Shift[i]) > 32 {
			return f, fmt.Errorf("palette: invalid packed RGB channel width/shift")
		}
		mask := uint32(1)<<bits - 1
		if used&(mask<<c.Shift[i]) != 0 {
			return f, fmt.Errorf("palette: overlapping packed RGB channels")
		}
		used |= mask << c.Shift[i]
		f.mask[i], f.shift[i] = mask, c.Shift[i]
	}
	return f, nil
}

func (f PackedRGB) normalized() PackedRGB {
	if f.mask == [3]uint32{} {
		return PackedRGB{shift: [3]uint8{8, 4, 0}, mask: [3]uint32{15, 15, 15}}
	}
	return f
}

// Color expands channels with integer division to an opaque eight-bit color.
func (f PackedRGB) Color(word uint32) color.NRGBA {
	f = f.normalized()
	return color.NRGBA{R: uint8((word >> f.shift[0] & f.mask[0]) * 255 / f.mask[0]),
		G: uint8((word >> f.shift[1] & f.mask[1]) * 255 / f.mask[1]),
		B: uint8((word >> f.shift[2] & f.mask[2]) * 255 / f.mask[2]), A: 255}
}

// Blend returns first+(second-first)*amount/steps per packed channel. Ratios
// must be within 0..1. Computation uses signed 64-bit products and allocates no
// storage. The resulting word contains only the configured channel bits.
func (f PackedRGB) Blend(first, second uint32, amount, steps uint32) (uint32, error) {
	if steps == 0 || amount > steps {
		return 0, fmt.Errorf("palette: invalid packed RGB blend ratio")
	}
	f = f.normalized()
	word := uint32(0)
	for i, shift := range f.shift {
		a, b := int64(first>>shift&f.mask[i]), int64(second>>shift&f.mask[i])
		word |= uint32(a+(b-a)*int64(amount)/int64(steps)) << shift
	}
	return word, nil
}
