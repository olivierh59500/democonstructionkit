package sprites

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
)

func TestProjectedFieldSharesOneTransportAcrossMaterials(t *testing.T) {
	c := ProjectedFieldConfig{
		Field: FieldConfig{Count: 2, Near: 0, Far: 10, Depth: DepthRespawn,
			Spawn: func(i int, reset bool) Point {
				if reset {
					return Point{X: float64(i), Z: 10}
				}
				return Point{X: float64(i), Z: float64(i + 4)}
			}},
		View:  FieldView{Camera: geometry.Camera{Center: geometry.Vec2{X: 10, Y: 10}, Focal: 10, Near: .001}},
		Delta: 1, Velocity: geometry.Vec3{Z: -1},
		Style: FieldStyle{VectorRects: true, Appearance: FieldAppearance{Width: 2, Height: 2, FillColor: color.RGBA{R: 255, A: 255}}},
	}
	p, err := NewProjectedField(c)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	if len(p.Samples()) != 2 || p.Samples()[0].Connected {
		t.Fatal("initial projection should precede the first movement")
	}
	if err := p.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	before := append([]FieldSample(nil), p.Samples()...)
	if !before[0].Connected || before[0].Z != 3 {
		t.Fatal("movement did not retain previous projected point", before)
	}
	dst := ebiten.NewImage(20, 20)
	p.Draw(dst)
	other := c.Style
	other.Appearance.FillColor = color.RGBA{B: 255, A: 255}
	p.DrawStyle(dst, other)
	for i, sample := range before {
		if p.Samples()[i] != sample {
			t.Fatal("changing the material advanced the field")
		}
	}
	if err := p.ResetCount(3); err != nil {
		t.Fatal(err)
	}
	if err := p.Update(kit.Frame{}); err != nil || len(p.Samples()) != 3 {
		t.Fatalf("resizing the projected field failed: count=%d, error=%v", len(p.Samples()), err)
	}
}

func BenchmarkProjectedFieldUpdate500(b *testing.B) {
	c := ProjectedFieldConfig{
		Field: FieldConfig{Count: 500, Near: 0, Far: 100, Depth: DepthRespawn,
			Spawn: func(i int, reset bool) Point {
				if reset {
					return Point{X: float64(i % 20), Z: 100}
				}
				return Point{X: float64(i % 20), Z: float64(i%99 + 1)}
			}},
		View:  FieldView{Camera: geometry.Camera{Center: geometry.Vec2{X: 320, Y: 240}, Focal: 200, Near: .001}},
		Delta: 1, Velocity: geometry.Vec3{Z: -2},
	}
	p, err := NewProjectedField(c)
	if err != nil {
		b.Fatal(err)
	}
	defer p.Close()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := p.Update(kit.Frame{}); err != nil {
			b.Fatal(err)
		}
	}
}
