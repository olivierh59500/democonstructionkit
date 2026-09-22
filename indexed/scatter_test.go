package indexed

import (
	"bytes"
	"testing"
)

func TestScatterModesPreserveOrderingAndBackground(t *testing.T) {
	config := ScatterConfig{SourceSize: 3, DestinationSize: 6, Offsets: []uint32{0, 2, 4, 5}, Destinations: []uint32{1, 3, 2, 3, 4}}
	m, err := NewScatterMap(config)
	if err != nil {
		t.Fatal(err)
	}
	config.Offsets[1], config.Destinations[0] = 0, 0
	for _, test := range []struct {
		mode ScatterMode
		want []byte
	}{
		{ScatterReplace, []byte{99, 250, 0, 0, 7, 99}},
		{ScatterOverBackground, []byte{99, 250, 30, 40, 7, 99}},
		{ScatterAddBackground, []byte{99, 14, 30, 40, 57, 99}},
	} {
		dst := bytes.Repeat([]byte{99}, 6)
		if err := m.Render(dst, []byte{250, 0, 7}, []byte{10, 20, 30, 40, 50, 60}, test.mode); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(dst, test.want) {
			t.Fatalf("mode %d: %v, want %v", test.mode, dst, test.want)
		}
	}
	out, source, bg := make([]byte, 6), []byte{1, 0, 2}, make([]byte, 6)
	if allocations := testing.AllocsPerRun(100, func() {
		if err := m.Render(out, source, bg, ScatterAddBackground); err != nil {
			panic(err)
		}
	}); allocations != 0 {
		t.Fatalf("scatter allocations = %g", allocations)
	}
}

func TestScatterStreamClippingAndValidation(t *testing.T) {
	// Source 0 has two addresses; one is clipped. Source 1 has none. Source 2
	// overwrites destination 2. The final byte is permitted resource padding.
	data := []byte{2, 0, 2, 0, 255, 255, 0, 0, 1, 0, 2, 0, 99}
	m, err := DecodeScatterMap16(data, 3, 4)
	if err != nil {
		t.Fatal(err)
	}
	dst := []byte{7, 7, 7, 7}
	if err := m.Render(dst, []byte{8, 9, 10}, nil, ScatterReplace); err != nil || !bytes.Equal(dst, []byte{7, 7, 10, 7}) {
		t.Fatal("scatter stream order/clipping", dst, err)
	}
	for _, truncated := range [][]byte{nil, data[:4], data[:7], data[:11]} {
		if _, err := DecodeScatterMap16(truncated, 3, 4); err == nil {
			t.Fatal("accepted truncated stream", truncated)
		}
	}
	for _, config := range []ScatterConfig{
		{}, {SourceSize: 1, DestinationSize: 4, Offsets: []uint32{1, 1}, Destinations: []uint32{1}},
		{SourceSize: 2, DestinationSize: 4, Offsets: []uint32{0, 2, 1}, Destinations: []uint32{1}},
		{SourceSize: 1, DestinationSize: 4, Offsets: []uint32{0, 1}, Destinations: []uint32{4}},
	} {
		if _, err := NewScatterMap(config); err == nil {
			t.Fatal("accepted invalid map", config)
		}
	}
	for _, mode := range []ScatterMode{ScatterOverBackground, ScatterMode(255)} {
		out := []byte{7, 7, 7, 7}
		if err := m.Render(out, []byte{8, 9, 10}, nil, mode); err == nil || !bytes.Equal(out, []byte{7, 7, 7, 7}) {
			t.Fatal("invalid render must reject before writing", out)
		}
	}
}
