package geometry

import (
	"math"
	"testing"
)

func TestClippingPreservesIntersectionUVs(t *testing.T) {
	tri := [3]Vertex{{Position: Vec3{-2, 0, 0}, UV: Vec2{0, 0}}, {Position: Vec3{2, 0, 2}, UV: Vec2{1, 0}}, {Position: Vec3{0, 2, 2}, UV: Vec2{0, 1}}}
	var scratch [4]Vertex
	got := ClipNear(scratch[:0], tri, 1)
	if len(got) != 4 {
		t.Fatal(got)
	}
	if got[0].Position != (Vec3{-1, 1, 1}) || got[0].UV != (Vec2{0, .5}) || got[1].UV != (Vec2{.5, 0}) {
		t.Fatal(got)
	}
	for _, v := range got {
		if v.Position.Z < 1 {
			t.Fatal(v)
		}
	}
}
func TestRotationAndProjection(t *testing.T) {
	v := Vec3{2, 3, 4}
	r := RotateXYZ(Vec3{.3, .7, .2}).Apply(v)
	if math.Abs(r.Dot(r)-v.Dot(v)) > 1e-10 {
		t.Fatal("rotation changed length")
	}
	c := Camera{Focal: 100, Near: 1, Center: Vec2{320, 200}}
	p, s, ok := c.Project(Vec3{2, 3, 10})
	if !ok || p != (Vec2{340, 230}) || s != 10 {
		t.Fatal(p, s, ok)
	}
	if _, _, ok = c.Project(Vec3{Z: .5}); ok {
		t.Fatal("projected point behind near plane")
	}
}
func TestMorphDifferentPointCounts(t *testing.T) {
	out := make([]Vec3, 3)
	Morph(out, []Vec3{{X: 2}}, []Vec3{{X: 4}, {X: 6}, {X: 8}}, .5)
	if out[0].X != 3 || out[2].X != 5 {
		t.Fatal(out)
	}
}
