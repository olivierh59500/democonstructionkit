package motion

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"testing"
)

func TestGlyphDelayPatternsMatchAuthoredGridOrders(t *testing.T) {
	tests := []struct {
		name string
		make func(int, int) ([]int, error)
		hash string
	}{
		{"serpentine", SerpentineGlyphDelays, "ad77e68c3c7b65518cd6ff3f239945ebb455a0d0d5973f0ad57ac34afce7a35b"},
		{"mirrored columns", MirroredColumnGlyphDelays, "dd58efbd13e2eae3836aab1f17ffef5673e97ce74935446e6a844d9c1b2129ef"},
		{"spiral", SpiralGlyphDelays, "e20b61dde54591acd5f3827028e44fb17e1141bd09f3c17430286009efee0198"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			values, err := test.make(20, 8)
			if err != nil {
				t.Fatal(err)
			}
			if len(values) != 160 {
				t.Fatalf("got %d delay cells", len(values))
			}
			encoded := make([]byte, len(values)*4)
			for i, value := range values {
				binary.LittleEndian.PutUint32(encoded[i*4:], uint32(value))
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(encoded)); got != test.hash {
				t.Fatalf("grid delay order changed: %s", got)
			}
		})
	}
	if _, err := MirroredColumnGlyphDelays(3, 4); err == nil {
		t.Fatal("accepted mirrored delay grid with odd width")
	}
	if _, err := SpiralGlyphDelays(0, 4); err == nil {
		t.Fatal("accepted empty delay grid")
	}
}
