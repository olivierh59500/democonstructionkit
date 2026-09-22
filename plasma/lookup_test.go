package plasma

import (
	"bytes"
	"math"
	"testing"
)

func TestLookupCustomWavesAndInterleavedRows(t *testing.T) {
	config := LookupConfig{
		Width: 5, ColorTable: []byte{10, 20, 30, 40, 250, 240, 230, 220},
		Waves: []Wave{
			{Displacement: []uint16{0, 2, 4, 6}, ColorStepX: 1, ColorStepY: -1, DisplacementStepX: -1, DisplacementStepY: 3, DisplacementFractionBits: 1},
			{Displacement: []uint16{1}, ColorOffset: 3, ColorStepX: -2, ColorStepY: 2},
		},
	}
	p, err := NewLookup(config)
	if err != nil {
		t.Fatal(err)
	}
	const stride, height = 7, 5
	out := bytes.Repeat([]byte{0xab}, stride*height)
	want := append([]byte(nil), out...)
	phases := []Phase{{Color: -5, Displacement: -3}, {Color: 11}}
	// Independently evaluate the scalar wave definition using floor division and
	// signed modulo, rather than the renderer's precompiled unsigned addresses.
	for y := 1; y < height; y += 2 {
		for x := 0; x < config.Width; x++ {
			value := 0
			for i, wave := range config.Waves {
				lookup := wave.DisplacementOffset + int64(x)*wave.DisplacementStepX + int64(y)*wave.DisplacementStepY + phases[i].Displacement
				lookup = int64(math.Floor(float64(lookup) / float64(uint64(1)<<wave.DisplacementFractionBits)))
				lookup = (lookup%int64(len(wave.Displacement)) + int64(len(wave.Displacement))) % int64(len(wave.Displacement))
				color := wave.ColorOffset + int64(x)*wave.ColorStepX + int64(y)*wave.ColorStepY + phases[i].Color + int64(wave.Displacement[lookup])
				color = (color%int64(len(config.ColorTable)) + int64(len(config.ColorTable))) % int64(len(config.ColorTable))
				value += int(config.ColorTable[color])
			}
			want[y*stride+x] = byte(value)
		}
	}
	if err := p.RenderRows(out, stride, Rows{First: 1, Count: 2, Step: 2}, phases); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, want) {
		t.Fatalf("custom/interleaved plasma = %v, want %v", out, want)
	}
	if p.Width() != config.Width {
		t.Fatal("compiled width changed")
	}
	// Subsequent caller mutation cannot change a compiled effect.
	config.ColorTable[0] = 0
	config.Waves[0].Displacement[0] = 255
	if err := p.RenderRows(out, stride, Rows{First: 1, Count: 2, Step: 2}, phases); err != nil || !bytes.Equal(out, want) {
		t.Fatal("compiled plasma retains caller-owned table slices")
	}
	if n := testing.AllocsPerRun(50, func() {
		if err := p.RenderRows(out, stride, Rows{First: 1, Count: 2, Step: 2}, phases); err != nil {
			panic(err)
		}
	}); n != 0 {
		t.Fatalf("steady render allocations = %g", n)
	}
}

func TestLookupValidationDoesNotPartiallyWrite(t *testing.T) {
	base := LookupConfig{Width: 3, ColorTable: []byte{1, 2}, Waves: []Wave{{Displacement: []uint16{0}}}}
	for _, invalid := range []LookupConfig{
		{}, {Width: -1, ColorTable: base.ColorTable, Waves: base.Waves},
		{Width: 3, ColorTable: []byte{1, 2, 3}, Waves: base.Waves},
		{Width: 3, ColorTable: base.ColorTable, Waves: []Wave{{Displacement: []uint16{1, 2, 3}}}},
		{Width: 3, ColorTable: base.ColorTable, Waves: []Wave{{Displacement: []uint16{0}, DisplacementFractionBits: 17}}},
	} {
		if _, err := NewLookup(invalid); err == nil {
			t.Fatal("accepted invalid configuration")
		}
	}
	p, err := NewLookup(base)
	if err != nil {
		t.Fatal(err)
	}
	for _, rows := range []Rows{
		{First: -1, Count: 1}, {Count: -1}, {Count: 1, Step: -1},
		{First: 4, Count: 1}, {Count: 3, Step: 2}, {Count: int(^uint(0) >> 1)},
	} {
		out := bytes.Repeat([]byte{0xc7}, 12)
		if err := p.RenderRows(out, 3, rows, nil); err == nil {
			t.Fatal("accepted invalid rows", rows)
		}
		if !bytes.Equal(out, bytes.Repeat([]byte{0xc7}, 12)) {
			t.Fatal("invalid call partially modified the destination")
		}
	}
	if err := p.Render(make([]byte, 12), 2, 4, nil); err == nil {
		t.Fatal("accepted overlapping row stride")
	}
	if err := p.Render(make([]byte, 12), 3, 4, []Phase{{}, {}}); err == nil {
		t.Fatal("accepted mismatched phase count")
	}
	if err := p.Render(nil, 3, 0, nil); err != nil {
		t.Fatal("empty output should be valid", err)
	}
	if err := p.Render(make([]byte, 11), 4, 3, nil); err != nil {
		t.Fatal("last row does not require padding", err)
	}
}

func BenchmarkLookup(b *testing.B) {
	colors := make([]byte, 16384)
	displacement := make([]uint16, 8192)
	for i := range colors {
		colors[i] = byte(i * 23)
	}
	for i := range displacement {
		displacement[i] = uint16(i * 79)
	}
	p, err := NewLookup(LookupConfig{Width: 84, ColorTable: colors, Waves: []Wave{
		{Displacement: displacement, ColorStepX: 8, DisplacementOffset: 640, DisplacementStepX: -8, DisplacementStepY: 2, DisplacementFractionBits: 1},
		{Displacement: displacement, ColorOffset: 320, ColorStepX: -4, ColorStepY: 2, DisplacementStepX: 32, DisplacementStepY: 2, DisplacementFractionBits: 1},
	}})
	if err != nil {
		b.Fatal(err)
	}
	dst := make([]byte, 84*280)
	phases := []Phase{{Color: 3500, Displacement: 4600}, {Color: 3900, Displacement: 7340}}
	b.SetBytes(int64(len(dst)))
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := p.Render(dst, 84, 280, phases); err != nil {
			b.Fatal(err)
		}
	}
}
