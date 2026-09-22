package composite

import (
	"image"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestWaterReflectionMatchesVectorballsMapping(t *testing.T) {
	c := DefaultWaterReflectionConfig()
	c.Source = image.Rect(0, 288, 640, 368)
	c.Horizon = 400
	w, err := NewWaterReflection(c)
	if err != nil {
		t.Fatal(err)
	}
	rows, next := w.prepareRows(c.Source, 0, 123)
	if rows != 80 || next != 80 {
		t.Fatalf("prepared %d rows through %d, want 80", rows, next)
	}
	var original ebiten.GeoM
	original.Scale(1, -1)
	original.Translate(0, 480)
	for _, v := range w.vertices[:rows*4] {
		x, y := original.Apply(float64(v.SrcX), float64(v.SrcY)-288)
		if float64(v.DstX) != x || float64(v.DstY) != y ||
			v.ColorR != .5 || v.ColorG != .5 || v.ColorB != .5 || v.ColorA != .5 {
			t.Fatalf("reflection differs from original transform: %+v; want (%g, %g), alpha .5", v, x, y)
		}
	}
}

func TestWaterReflectionPreservesPartialRowsAndBatchEdges(t *testing.T) {
	c := DefaultWaterReflectionConfig()
	c.Source = image.Rect(17, 23, 44, 2052)
	c.X, c.Horizon, c.ScaleY = 5, 80, .75
	c.RowHeight = 3
	c.Fade = 1
	w, err := NewWaterReflection(c)
	if err != nil {
		t.Fatal(err)
	}
	previousY := float32(c.Horizon)
	totalRows := 0
	for top := 0; top < c.Source.Dy(); {
		rows, next := w.prepareRows(c.Source, top, 0)
		if rows < 1 || rows > waterBatchRows || next <= top {
			t.Fatalf("invalid batch: rows=%d, top=%d, next=%d", rows, top, next)
		}
		for row := range rows {
			vertices := w.vertices[row*4 : row*4+4]
			if vertices[0].DstY != previousY || vertices[2].DstY <= previousY {
				t.Fatalf("gap or overlap at row %d", totalRows+row)
			}
			previousY = vertices[2].DstY
		}
		if next == c.Source.Dy() {
			last := w.vertices[(rows-1)*4+2]
			if last.SrcY != float32(c.Source.Min.Y) || last.ColorA != 0 {
				t.Fatalf("last partial row lost the crop edge or fade: %+v", last)
			}
		}
		totalRows += rows
		top = next
	}
	if totalRows != (c.Source.Dy()+2)/3 || previousY != float32(c.Horizon+float64(c.Source.Dy())*c.ScaleY) {
		t.Fatalf("lost final partial row: %d rows through %g", totalRows, previousY)
	}
}

func TestWaterWaveUsesAbsoluteSecondsAndPhase(t *testing.T) {
	wave := WaterWave{Amplitude: 7, Wavelength: 24, Speed: -math.Pi, Phase: math.Pi / 2}
	for _, tc := range []struct{ distance, seconds, want float64 }{
		{0, 0, 7}, {6, 0, 0}, {12, 0, -7}, {0, 1, -7}, {12, 1, 7},
	} {
		if got := wave.offset(tc.distance, tc.seconds); math.Abs(got-tc.want) > 1e-12 {
			t.Fatalf("offset(%g, %g) = %g, want %g", tc.distance, tc.seconds, got, tc.want)
		}
	}
	if (WaterWave{}).offset(1, 1) != 0 {
		t.Fatal("disabled waves must not divide by a zero wavelength")
	}
}

func TestWaterReflectionValidatesConfigurationAtomically(t *testing.T) {
	valid := DefaultWaterReflectionConfig()
	w, err := NewWaterReflection(valid)
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*WaterReflectionConfig){
		func(c *WaterReflectionConfig) { c.ScaleY = 0 },
		func(c *WaterReflectionConfig) { c.Alpha = -1 },
		func(c *WaterReflectionConfig) { c.Fade = 1.1 },
		func(c *WaterReflectionConfig) { c.RowHeight = -1 },
		func(c *WaterReflectionConfig) { c.Horizon = math.Inf(1) },
		func(c *WaterReflectionConfig) { c.Wave.Phase = math.NaN() },
		func(c *WaterReflectionConfig) { c.Wave.Amplitude, c.Wave.Wavelength = 2, 0 },
	} {
		bad := valid
		change(&bad)
		if err := w.SetConfig(bad); err == nil {
			t.Fatalf("accepted invalid configuration: %+v", bad)
		}
		if w.Config() != valid {
			t.Fatal("invalid configuration partially modified the effect")
		}
	}
}

func TestWaterReflectionGeometryDoesNotAllocate(t *testing.T) {
	c := DefaultWaterReflectionConfig()
	c.Wave = WaterWave{Amplitude: 5, Wavelength: 40, Speed: 2, Phase: .3}
	c.Fade = .8
	w, err := NewWaterReflection(c)
	if err != nil {
		t.Fatal(err)
	}
	crop := image.Rect(11, 29, 651, 189)
	if allocations := testing.AllocsPerRun(100, func() { w.prepareRows(crop, 0, 3) }); allocations != 0 {
		t.Fatalf("geometry preparation allocated %g objects per frame", allocations)
	}
}

func BenchmarkWaterReflectionGeometry(b *testing.B) {
	c := DefaultWaterReflectionConfig()
	c.Wave = WaterWave{Amplitude: 5, Wavelength: 40, Speed: 2, Phase: .3}
	c.Fade = .8
	w, err := NewWaterReflection(c)
	if err != nil {
		b.Fatal(err)
	}
	crop := image.Rect(0, 0, 640, 160)
	b.ReportAllocs()
	for b.Loop() {
		w.prepareRows(crop, 0, 3)
	}
}
