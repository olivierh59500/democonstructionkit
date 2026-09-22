package indexed

import (
	"bytes"
	"testing"
)

func TestPaletteFadePreservesIntegerRoundingAndRanges(t *testing.T) {
	source := []byte{0, 31, 63, 100, 200, 255}
	out := []byte{91, 92, 93, 0, 0, 0, 0, 0, 0, 94, 95, 96}
	if err := FadePalette(out[3:9], source, [3]byte{63, 63, 63}, 17, 128); err != nil {
		t.Fatal(err)
	}
	for i, value := range source {
		want := byte((int(value)*(128-17) + 63*17) / 128)
		if out[3+i] != want {
			t.Fatalf("component %d: %d, want %d", i, out[3+i], want)
		}
	}
	if !bytes.Equal(out[:3], []byte{91, 92, 93}) || !bytes.Equal(out[9:], []byte{94, 95, 96}) {
		t.Fatal("palette range overwritten")
	}
	if err := FadePalette(source, source, [3]byte{}, 64, 64); err != nil || !bytes.Equal(source, make([]byte, 6)) {
		t.Fatal("in-place endpoint failed")
	}
	if err := MixPalette(source, source, []byte{1, 2, 3, 4, 5, 6}, 1, 2); err != nil || !bytes.Equal(source, []byte{0, 1, 1, 2, 2, 3}) {
		t.Fatal("palette mix rounding failed")
	}
	if allocations := testing.AllocsPerRun(100, func() { _ = FadePalette(source, source, [3]byte{63, 63, 63}, 1, 4) }); allocations != 0 {
		t.Fatalf("palette fade allocations = %g", allocations)
	}
	before := append([]byte(nil), source...)
	if err := FadePalette(source, source, [3]byte{}, 2, 1); err == nil || !bytes.Equal(source, before) {
		t.Fatal("invalid fade partially modified output")
	}
}
