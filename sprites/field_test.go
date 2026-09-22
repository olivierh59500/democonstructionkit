package sprites

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/geometry"
)

func TestFieldMatchesWrappedRotatingProjection(t *testing.T) {
	points := []Point{{X: -10, Y: 4, Z: 0}, {X: 60, Y: -45, Z: 75}, {X: 3, Y: 1, Z: 130}}
	f, err := NewField(FieldConfig{Points: points, Depth: DepthWrap, Near: 0, Far: 130})
	if err != nil {
		t.Fatal(err)
	}
	for _, travel := range []float64{0, 1.5, 130, 900, 10000.5} {
		view := FieldView{Camera: geometry.Camera{Center: geometry.Vec2{X: 160, Y: 100}, Focal: 128, Near: math.SmallestNonzeroFloat64}, Offset: geometry.Vec3{Z: -travel}, Angle: .37}
		got := f.Sample(view)
		n := 0
		for index, p := range points {
			z := p.Z - travel
			if z > 130 || z < 0 {
				z -= 130 * math.Floor(z/130)
			}
			if z == 0 {
				continue
			}
			sin, cos := math.Sincos(view.Angle)
			x := (p.X*cos-p.Y*sin)*(128/z) + 160
			y := (p.X*sin+p.Y*cos)*(128/z) + 100
			if n >= len(got) || got[n].Index != index || got[n].X != x || got[n].Y != y || got[n].Z != z {
				t.Fatalf("travel %g index %d: %v, want %g,%g,%g", travel, index, got, x, y, z)
			}
			n++
		}
		if len(got) != n {
			t.Fatal("unexpected projected points")
		}
	}
}

func TestFieldRendererReplacesAtlasCacheInsteadOfRetainingOldStyles(t *testing.T) {
	source := ebiten.NewImage(256, 4)
	defer source.Deallocate()
	dst := ebiten.NewImage(20, 20)
	defer dst.Deallocate()
	r := NewFieldRenderer(4)
	defer r.Close()
	for x := 0; x < 250; x++ {
		r.Draw(dst, []FieldSample{{Image: 0}}, FieldStyle{Image: source, DrawImages: true, Frames: []image.Rectangle{image.Rect(x, 0, x+4, 4)}})
		if len(r.frameImages) != 1 || len(r.frameRects) != 1 {
			t.Fatal("obsolete atlas frames retained")
		}
	}
	r.Draw(dst, nil, FieldStyle{})
	if r.frameSource != nil || len(r.frameImages) != 0 {
		t.Fatal("skin replacement retained borrowed atlas")
	}
}

func TestFieldRespawnOrderAndIndependentSkinSamples(t *testing.T) {
	var calls []int
	f, err := NewField(FieldConfig{Count: 3, Near: 0, Far: 32, Depth: DepthRespawn, Spawn: func(i int, reset bool) Point {
		calls = append(calls, i)
		z := float64(i+1) * .1
		if reset {
			z = 32
		}
		return Point{X: float64(i), Y: 2, Z: z, Image: i}
	}})
	if err != nil {
		t.Fatal(err)
	}
	f.Step(1, geometry.Vec3{Z: -.2})
	if len(calls) != 5 || calls[3] != 0 || calls[4] != 1 {
		t.Fatal(calls)
	}
	view := FieldView{Camera: geometry.Camera{Focal: 64, Near: .001}}
	first := append([]FieldSample(nil), f.Sample(view)...)
	for repeat := 0; repeat < 3; repeat++ {
		got := f.Samples()
		for i := range first {
			if got[i] != first[i] {
				t.Fatal("reading samples advanced the field")
			}
		}
	}
	f.Step(1, geometry.Vec3{Z: -.2})
	second := f.Sample(view)
	if !second[0].Connected || !second[1].Connected || second[2].Connected {
		t.Fatal(second)
	}
	if err := f.ResetCount(2); err != nil {
		t.Fatal(err)
	}
	for _, p := range f.Sample(view) {
		if p.Connected {
			t.Fatal("reset kept old trail")
		}
	}
}

func TestFieldWrapBreaksStreakAndSortingRetainsIdentity(t *testing.T) {
	f, err := NewField(FieldConfig{Points: []Point{{Z: 1}, {Z: 5}, {Z: 3}}, Near: 0, Far: 10, Depth: DepthWrap})
	if err != nil {
		t.Fatal(err)
	}
	view := FieldView{Camera: geometry.Camera{Focal: 64, Near: .001}, SortDepth: true}
	first := f.Sample(view)
	if first[0].Index != 1 || first[1].Index != 2 || first[2].Index != 0 {
		t.Fatal(first)
	}
	f.Step(1, geometry.Vec3{Z: -2})
	for _, p := range f.Sample(view) {
		if p.Connected != (p.Index != 0) {
			t.Fatal("trail crossed a wrap", p)
		}
	}
}

func TestFieldWarmSamplingDoesNotAllocate(t *testing.T) {
	f, err := NewField(FieldConfig{Count: 400, Depth: DepthWrap, Near: 0, Far: 130, Spawn: func(i int, _ bool) Point { return Point{X: float64(i % 25), Y: float64(i % 13), Z: float64(i) * .325} }})
	if err != nil {
		t.Fatal(err)
	}
	view := FieldView{Camera: geometry.Camera{Focal: 128, Near: .001}, SortDepth: true}
	f.Sample(view)
	allocs := testing.AllocsPerRun(100, func() { f.Step(1, geometry.Vec3{Z: -1.5}); f.Sample(view) })
	if allocs != 0 {
		t.Fatalf("steady-state allocations: %g", allocs)
	}
}

func BenchmarkField400(b *testing.B) {
	f, _ := NewField(FieldConfig{Count: 400, Depth: DepthWrap, Near: 0, Far: 130, Spawn: func(i int, _ bool) Point { return Point{X: float64(i % 25), Y: float64(i % 13), Z: float64(i) * .325} }})
	view := FieldView{Camera: geometry.Camera{Focal: 128, Near: .001}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Step(1, geometry.Vec3{Z: -1.5})
		f.Sample(view)
	}
}
