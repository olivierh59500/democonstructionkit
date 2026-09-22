package indexed

import "fmt"

// FadePalette interpolates packed RGB palette components towards one color.
// amount/steps is the target weight. Integer truncation and the caller's channel
// precision are preserved: a six-bit palette remains six-bit, an eight-bit
// palette remains eight-bit. Slice the palette to fade only selected entries.
// Source and destination may be the same slice. No allocation is performed.
func FadePalette(dst, source []byte, target [3]byte, amount, steps uint32) error {
	if steps == 0 || amount > steps || len(dst) != len(source) || len(source)%3 != 0 {
		return fmt.Errorf("indexed: invalid palette fade")
	}
	a, b, denominator := uint64(steps-amount), uint64(amount), uint64(steps)
	for i, value := range source {
		dst[i] = byte((uint64(value)*a + uint64(target[i%3])*b) / denominator)
	}
	return nil
}

// MixPalette interpolates two packed RGB palettes with integer truncation.
// amount/steps is the second palette's weight. Either input may alias dst.
func MixPalette(dst, first, second []byte, amount, steps uint32) error {
	if steps == 0 || amount > steps || len(dst) != len(first) || len(dst) != len(second) || len(dst)%3 != 0 {
		return fmt.Errorf("indexed: invalid palette mix")
	}
	a, b, denominator := uint64(steps-amount), uint64(amount), uint64(steps)
	for i, value := range first {
		dst[i] = byte((uint64(value)*a + uint64(second[i])*b) / denominator)
	}
	return nil
}
