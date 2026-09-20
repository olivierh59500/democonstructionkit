package sprites

import (
	"github.com/olivierh59500/democonstructionkit/motion"
	"math"
	"testing"
)

func TestCubeTopology(t *testing.T) {
	for fill, want := range map[Fill]int{Edges: 32, Surface: 56, Solid: 64} {
		points, err := Cube(CubeConfig{Size: 12, Segments: 3, Fill: fill, Image: 7})
		if err != nil || len(points) != want {
			t.Fatalf("fill %d: count %d, err %v", fill, len(points), err)
		}
		seen := map[Point]bool{}
		for _, p := range points {
			if seen[p] || p.Image != 7 || math.Abs(p.X) > 6 || math.Abs(p.Y) > 6 || math.Abs(p.Z) > 6 {
				t.Fatalf("invalid or duplicate point: %+v", p)
			}
			seen[p] = true
		}
	}
}

func TestPyramidAndFlag(t *testing.T) {
	for _, fill := range []Fill{Edges, Surface, Solid} {
		ps, err := Pyramid(PyramidConfig{Width: 10, Height: 12, Segments: 4, Fill: fill})
		if err != nil {
			t.Fatal(err)
		}
		seen := map[Point]bool{}
		apex := 0
		for _, p := range ps {
			if seen[p] {
				t.Fatal("duplicate point", p)
			}
			seen[p] = true
			if p.Y == 6 {
				apex++
				if p.X != 0 || p.Z != 0 {
					t.Fatal("invalid apex")
				}
			}
		}
		if apex != 1 {
			t.Fatal("apex missing")
		}
	}
	base, err := Plane(PlaneConfig{Width: 10, Height: 6, Columns: 5, Rows: 3, Image: 2})
	if err != nil {
		t.Fatal(err)
	}
	dst := make([]Point, len(base))
	flag := Flag{Width: 10, PinLeft: true, Wave: motion.Wave{Amplitude: 3, Spatial: .3, Speed: 2}}
	flag.Apply(dst, base, 1)
	changed := false
	for i, p := range dst {
		if base[i].Z != 0 || p.X != base[i].X || p.Y != base[i].Y || p.Image != 2 {
			t.Fatal("rest shape or attributes changed")
		}
		if p.X == -5 && p.Z != 0 {
			t.Fatal("pinned edge moved")
		}
		changed = changed || p.Z != 0
	}
	if !changed {
		t.Fatal("flag did not wave")
	}
}

func TestRejectInvalidShapes(t *testing.T) {
	if _, err := Cube(CubeConfig{Size: math.NaN(), Segments: 3}); err == nil {
		t.Fatal("NaN accepted")
	}
	if _, err := Pyramid(PyramidConfig{Width: 2, Height: 3, Segments: 0}); err == nil {
		t.Fatal("zero segments accepted")
	}
	if _, err := Plane(PlaneConfig{Width: 2, Height: 3, Rows: 2, Columns: 1000000}); err == nil {
		t.Fatal("unbounded allocation accepted")
	}
}
