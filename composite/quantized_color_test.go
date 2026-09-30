package composite

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestQuantizedColorValidation(t *testing.T) {
	for _, config := range []QuantizedColorConfig{
		{Levels: [3]uint16{15, 0, 15}}, {Levels: [3]uint16{256, 15, 15}},
		{AlphaThreshold: -1}, {AlphaThreshold: 1.1}, {AlphaThreshold: float32(math.NaN())},
	} {
		if q, err := NewQuantizedColor(config); err == nil {
			q.Close()
			t.Fatalf("accepted invalid color-grid config %+v", config)
		}
	}
	q, err := NewQuantizedColor(QuantizedColorConfig{})
	if err != nil {
		t.Fatal(err)
	}
	defer q.Close()
	for _, state := range []QuantizedColorState{
		{Mode: QuantizedColorMode(5)}, {Mode: QuantizedScale},
		{Mode: QuantizedScale, Numerator: 2, Denominator: 1},
		{Mode: QuantizedScale, Denominator: 65537},
		{Mode: QuantizedFromTarget, Target: [3]uint16{16, 0, 0}, Denominator: 16},
	} {
		if q.validate(state) == nil {
			t.Fatalf("accepted invalid quantized-color state %+v", state)
		}
	}
	for _, state := range []QuantizedColorState{
		{}, {Mode: QuantizedKeep}, {Mode: QuantizedReplace, Target: [3]uint16{16, 0, 0}},
		{Mode: QuantizedScale, Numerator: 65536, Denominator: 65536},
		{Mode: QuantizedFromTarget, Target: [3]uint16{15, 0, 8}, Numerator: 4, Denominator: 16},
	} {
		if err := q.validate(state); err != nil {
			t.Fatal(err)
		}
	}
	if q.levels != [3]uint16{15, 15, 15} {
		t.Fatal("zero configuration did not select RGB12")
	}
	src := ebiten.NewImage(1, 1)
	defer src.Deallocate()
	if q.Draw(src, src, QuantizedColorState{}) == nil {
		t.Fatal("accepted source/destination aliasing")
	}
	q.Close()
	if q.Draw(src, src, QuantizedColorState{}) == nil {
		t.Fatal("drew a closed color pass")
	}
}
