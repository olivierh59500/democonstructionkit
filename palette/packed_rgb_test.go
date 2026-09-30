package palette_test

import (
	"image/color"
	"math"
	"math/big"
	"testing"

	"github.com/olivierh59500/democonstructionkit/palette"
)

func TestPackedRGBZeroValueDecodesEveryRGB12Color(t *testing.T) {
	var zero palette.PackedRGB
	configured, err := palette.NewPackedRGB(palette.PackedRGBConfig{})
	if err != nil {
		t.Fatal(err)
	}
	for word := uint32(0); word < 4096; word++ {
		want := color.NRGBA{
			R: uint8(word/256) * 17,
			G: uint8(word/16%16) * 17,
			B: uint8(word%16) * 17,
			A: 255,
		}
		for _, input := range []uint32{word, word | 0xfffff000} {
			if got := zero.Color(input); got != want {
				t.Fatalf("zero RGB12 color %#x = %+v, want %+v", input, got, want)
			}
			if got := configured.Color(input); got != want {
				t.Fatalf("configured default color %#x = %+v, want %+v", input, got, want)
			}
		}
	}
}

func TestPackedRGBDecodesEveryRGB565Color(t *testing.T) {
	f, err := palette.NewPackedRGB(palette.PackedRGBConfig{
		Bits: [3]uint8{5, 6, 5}, Shift: [3]uint8{11, 5, 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	for word := uint32(0); word < 65536; word++ {
		want := color.NRGBA{
			R: uint8(word / 2048 * 255 / 31),
			G: uint8(word / 32 % 64 * 255 / 63),
			B: uint8(word % 32 * 255 / 31),
			A: 255,
		}
		if got := f.Color(word | 0xffff0000); got != want {
			t.Fatalf("RGB565 color %#x = %+v, want %+v", word, got, want)
		}
	}
}

func TestPackedRGB888AndHighBitLayouts(t *testing.T) {
	for _, config := range []palette.PackedRGBConfig{
		{Bits: [3]uint8{8, 8, 8}, Shift: [3]uint8{16, 8, 0}},
		{Bits: [3]uint8{8, 8, 8}, Shift: [3]uint8{24, 0, 12}},
	} {
		f, err := palette.NewPackedRGB(config)
		if err != nil {
			t.Fatal(err)
		}
		for channel := 0; channel < 3; channel++ {
			for value := uint32(0); value < 256; value++ {
				word := value << config.Shift[channel]
				want := color.NRGBA{A: 255}
				switch channel {
				case 0:
					want.R = uint8(value)
				case 1:
					want.G = uint8(value)
				case 2:
					want.B = uint8(value)
				}
				if got := f.Color(word); got != want {
					t.Fatalf("layout %+v channel %d value %d = %+v, want %+v", config, channel, value, got, want)
				}
			}
		}
	}
	f, err := palette.NewPackedRGB(palette.PackedRGBConfig{
		Bits: [3]uint8{1, 1, 1}, Shift: [3]uint8{31, 1, 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := f.Color(1 << 31); got != (color.NRGBA{R: 255, A: 255}) {
		t.Fatalf("highest word bit decoded as %+v", got)
	}
}

// The oracle uses unsigned distances and separate ascending/descending cases.
// In particular, the descending case rounds the distance before subtracting it.
func packedChannelFadeOracle(first, second, amount, steps uint32) uint32 {
	if second >= first {
		return first + (second-first)*amount/steps
	}
	return first - (first-second)*amount/steps
}

func TestPackedRGB12SignedFadesExhaustEveryChannelPair(t *testing.T) {
	var f palette.PackedRGB
	for _, shift := range []uint{0, 4, 8} {
		for first := uint32(0); first < 16; first++ {
			for second := uint32(0); second < 16; second++ {
				for steps := uint32(1); steps <= 32; steps++ {
					for amount := uint32(0); amount <= steps; amount++ {
						got, err := f.Blend(first<<shift, second<<shift, amount, steps)
						want := packedChannelFadeOracle(first, second, amount, steps) << shift
						if err != nil || got != want {
							t.Fatalf("channel shift %d fade %d -> %d at %d/%d = %#x, %v; want %#x", shift, first, second, amount, steps, got, err, want)
						}
					}
				}
			}
		}
	}
}

func TestPackedRGB12CombinedChannelsAndUnusedBits(t *testing.T) {
	var f palette.PackedRGB
	for first := uint32(0); first < 4096; first++ {
		second := 4095 - first
		for _, steps := range []uint32{2, 3, 7, 16, 31} {
			for amount := uint32(0); amount <= steps; amount++ {
				var want uint32
				for shift := uint(0); shift < 12; shift += 4 {
					want |= packedChannelFadeOracle(first>>shift&15, second>>shift&15, amount, steps) << shift
				}
				got, err := f.Blend(first|0xf0000000, second|0x00fff000, amount, steps)
				if err != nil || got != want {
					t.Fatalf("combined RGB12 fade %#x -> %#x at %d/%d = %#x, %v; want %#x", first, second, amount, steps, got, err, want)
				}
			}
		}
	}
	got, err := f.Blend(0xfff, 0, 1, 2)
	if err != nil || got != 0x888 {
		t.Fatalf("descending half fade = %#x, %v; want 0x888", got, err)
	}
}

func TestPackedRGB565FadesExhaustEveryChannelPair(t *testing.T) {
	config := palette.PackedRGBConfig{Bits: [3]uint8{5, 6, 5}, Shift: [3]uint8{11, 5, 0}}
	f, err := palette.NewPackedRGB(config)
	if err != nil {
		t.Fatal(err)
	}
	for channel, bits := range config.Bits {
		limit := uint32(1) << bits
		shift := config.Shift[channel]
		for first := uint32(0); first < limit; first++ {
			for second := uint32(0); second < limit; second++ {
				for _, steps := range []uint32{1, 2, 3, 7, 16, 31} {
					for amount := uint32(0); amount <= steps; amount++ {
						got, err := f.Blend(first<<shift, second<<shift, amount, steps)
						want := packedChannelFadeOracle(first, second, amount, steps) << shift
						if err != nil || got != want {
							t.Fatalf("RGB565 channel %d fade %d -> %d at %d/%d = %#x, %v; want %#x", channel, first, second, amount, steps, got, err, want)
						}
					}
				}
			}
		}
	}
}

// An arbitrary-precision oracle also checks products far beyond 32-bit range.
func packedBigFadeOracle(first, second, amount, steps uint32) uint32 {
	var delta, product, quotient, divisor big.Int
	delta.Sub(new(big.Int).SetUint64(uint64(second)), new(big.Int).SetUint64(uint64(first)))
	product.Mul(&delta, new(big.Int).SetUint64(uint64(amount)))
	divisor.SetUint64(uint64(steps))
	quotient.Quo(&product, &divisor)
	quotient.Add(&quotient, new(big.Int).SetUint64(uint64(first)))
	return uint32(quotient.Uint64())
}

func TestPackedRGBBlendWithMaximumUint32Ratios(t *testing.T) {
	for _, config := range []palette.PackedRGBConfig{
		{Bits: [3]uint8{4, 4, 4}, Shift: [3]uint8{8, 4, 0}},
		{Bits: [3]uint8{5, 6, 5}, Shift: [3]uint8{11, 5, 0}},
		{Bits: [3]uint8{8, 8, 8}, Shift: [3]uint8{16, 8, 0}},
		{Bits: [3]uint8{8, 8, 8}, Shift: [3]uint8{24, 0, 12}},
	} {
		f, err := palette.NewPackedRGB(config)
		if err != nil {
			t.Fatal(err)
		}
		var red, greenBlue, used uint32
		for channel, bits := range config.Bits {
			mask := (uint32(1)<<bits - 1) << config.Shift[channel]
			used |= mask
			if channel == 0 {
				red = mask
			} else {
				greenBlue |= mask
			}
		}
		for _, pair := range [][2]uint32{{0, used}, {used, 0}, {red, greenBlue}, {greenBlue, red}} {
			for _, amount := range []uint32{0, 1, math.MaxUint32 / 2, 1<<31 + 17, math.MaxUint32 - 1, math.MaxUint32} {
				var want uint32
				for channel, bits := range config.Bits {
					mask, shift := uint32(1)<<bits-1, config.Shift[channel]
					first, second := pair[0]>>shift&mask, pair[1]>>shift&mask
					want |= packedBigFadeOracle(first, second, amount, math.MaxUint32) << shift
				}
				got, err := f.Blend(pair[0]|^used, pair[1]|^used, amount, math.MaxUint32)
				if err != nil || got != want {
					t.Fatalf("layout %+v large-ratio fade %#x -> %#x at %d/%d = %#x, %v; want %#x", config, pair[0], pair[1], amount, uint32(math.MaxUint32), got, err, want)
				}
			}
		}
	}
}

func TestPackedRGBRejectsInvalidLayouts(t *testing.T) {
	for _, config := range []palette.PackedRGBConfig{
		{Bits: [3]uint8{0, 4, 4}, Shift: [3]uint8{8, 4, 0}},
		{Bits: [3]uint8{9, 4, 4}, Shift: [3]uint8{16, 4, 0}},
		{Bits: [3]uint8{255, 4, 4}, Shift: [3]uint8{16, 4, 0}},
		{Bits: [3]uint8{4, 4, 4}, Shift: [3]uint8{29, 4, 0}},
		{Bits: [3]uint8{4, 4, 4}, Shift: [3]uint8{255, 4, 0}},
		{Bits: [3]uint8{4, 4, 4}, Shift: [3]uint8{8, 8, 0}},
		{Bits: [3]uint8{8, 8, 8}, Shift: [3]uint8{16, 12, 0}},
		{Shift: [3]uint8{1, 0, 0}},
	} {
		if _, err := palette.NewPackedRGB(config); err == nil {
			t.Fatalf("accepted invalid packed layout %+v", config)
		}
	}
}

func TestPackedRGBRejectsInvalidRatios(t *testing.T) {
	var f palette.PackedRGB
	for _, ratio := range [][2]uint32{{0, 0}, {1, 0}, {2, 1}, {32, 31}, {math.MaxUint32, math.MaxUint32 - 1}} {
		if _, err := f.Blend(0xfff, 0, ratio[0], ratio[1]); err == nil {
			t.Fatalf("accepted invalid ratio %d/%d", ratio[0], ratio[1])
		}
	}
}
