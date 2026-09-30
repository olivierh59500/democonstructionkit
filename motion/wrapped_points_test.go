package motion

import (
	"image"
	"testing"
)

func TestWrappedPointProjectionMatchesNativeWordEquations(t *testing.T) {
	for _, kind := range []string{"star-pages", "perspective"} {
		t.Run(kind, func(t *testing.T) {
			config := WrappedPointProjectionConfig{Mask: [3]uint16{511, 511, 2047}, Bias: [2]int16{-256, -256},
				Numerator: 262144, DepthBias: 10, Shift: 9, Center: image.Pt(176, 143), Bounds: image.Rect(0, 0, 352, 286)}
			if kind == "perspective" {
				config.Mask[0], config.Bias[0], config.Numerator, config.DepthBias = 1023, -512, 263680, 390
				config.Center.Y, config.Bounds.Max.Y = 100, 199
			}
			projection, err := NewWrappedPointProjection(config)
			if err != nil {
				t.Fatal(err)
			}
			// Exhaust a full word cycle with independently evolving coordinates
			// and offsets; this includes negative words and overflowed additions.
			for i := 0; i < 65536; i++ {
				point := WrappedPoint{int16(i), int16(i*13 + 7), int16(i*29 + 11)}
				offset := [3]int16{int16(i*17 - 9), int16(i*31 + 32760), int16(i*37 - 32767)}
				var x, y int32
				var factor int32
				z := (uint16(point.Z) + uint16(offset[2])) & 2047
				centerY, height := 143, 286
				if kind == "star-pages" {
					x = int32(int16((uint16(point.X)+uint16(offset[0]))&511) - 256)
					y = int32(int16((uint16(point.Y)+uint16(offset[1]))&511) - 256)
					factor = int32(262144 / (int(z) + 10))
				} else {
					x = int32(int16((uint16(point.X)+uint16(offset[0]))&1023) - 512)
					y = int32(int16((uint16(point.Y)+uint16(offset[1]))&511) - 256)
					factor = int32(263680 / (int(z) + 390))
					centerY, height = 100, 199
				}
				px, py := int(int16(x*factor>>9))+176, int(int16(y*factor>>9))+centerY
				visible := px >= 0 && px < 352 && py >= 0 && py < height
				pose, ok := projection.Project(point, offset)
				if pose.X != px || pose.Y != py || pose.Depth != z || ok != visible {
					t.Fatalf("word %d: %+v/%v, want %d,%d,%d/%v", i, pose, ok, px, py, z, visible)
				}
			}
		})
	}
}

func TestWrappedPointProjectionRetainsEveryIntegerTruncation(t *testing.T) {
	config := WrappedPointProjectionConfig{Mask: [3]uint16{65535, 65535, 65535}, Bias: [2]int16{32767, -32768},
		Numerator: 2147483647, Shift: 4}
	projection, err := NewWrappedPointProjection(config)
	if err != nil {
		t.Fatal(err)
	}
	pose, ok := projection.Project(WrappedPoint{32760, -32760, 0}, [3]int16{10, -10, 1})
	// Word addition gives (-32766,32766); word bias gives (1,-2).
	// Long products are (2147483647,2); >>4 gives (134217727,0), whose
	// final words are (-1,0). Wide or floating-point products change this result.
	if !ok || pose != (WrappedPointPose{X: -1, Y: 0, Depth: 1}) {
		t.Fatal("projection discarded a word/long truncation", pose, ok)
	}
	if allocations := testing.AllocsPerRun(100, func() {
		projection.Project(WrappedPoint{32760, -32760, 0}, [3]int16{10, -10, 1})
	}); allocations != 0 {
		t.Fatal("projection allocated", allocations)
	}
}

func TestWrappedPointProjectionClippingDenominatorsAndConfigCopy(t *testing.T) {
	config := WrappedPointProjectionConfig{Mask: [3]uint16{65535, 65535, 65535}, Numerator: 1,
		DepthBias: -5, Center: image.Pt(11, 21), Bounds: image.Rect(10, 20, 14, 24)}
	projection, err := NewWrappedPointProjection(config)
	if err != nil {
		t.Fatal(err)
	}
	config.Center = image.Pt(99, 99)
	for _, z := range []int16{4, 5} {
		if _, ok := projection.Project(WrappedPoint{Z: z}, [3]int16{}); ok {
			t.Fatal("nonpositive denominator was projected")
		}
	}
	pose, ok := projection.Project(WrappedPoint{Z: 6}, [3]int16{})
	if !ok || pose.X != 11 || pose.Y != 21 {
		t.Fatal("source config was retained or nonzero clip origin changed", pose, ok)
	}
	if _, ok := projection.Project(WrappedPoint{X: 3, Z: 6}, [3]int16{}); ok {
		t.Fatal("half-open right edge was not clipped")
	}
	for _, invalid := range []WrappedPointProjectionConfig{
		{}, {Numerator: -1}, {Numerator: 1, Shift: 32},
		{Numerator: 1, Center: image.Pt(1<<30+1, 0)},
		{Numerator: 1, Bounds: image.Rect(0, 0, 8193, 1)},
	} {
		if _, err := NewWrappedPointProjection(invalid); err == nil {
			t.Fatal("invalid fixed projection was accepted", invalid)
		}
	}
	var absent *WrappedPointProjection
	if _, ok := absent.Project(WrappedPoint{}, [3]int16{}); ok {
		t.Fatal("nil projection accepted a point")
	}
}
