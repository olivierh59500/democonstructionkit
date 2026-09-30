package sprites

import (
	"image"
	"image/color"
	"testing"

	kit "github.com/olivierh59500/democonstructionkit"
)

func pointPlaneFixture(collision PointPlaneCollision, points *[]PointPlaneSample) IndexedPointPlaneConfig {
	return IndexedPointPlaneConfig{Width: 8, Height: 5, Count: len(*points), Collision: collision,
		Palette: []color.NRGBA{{}, {R: 255, A: 255}, {G: 255, A: 255}, {B: 255, A: 255}},
		Sample:  func(index int) (PointPlaneSample, bool) { return (*points)[index], true }}
}

func TestIndexedPointPlaneCollisionsDeduplicateZeroAndResetSparseStorage(t *testing.T) {
	for _, collision := range []PointPlaneCollision{PointPlaneOR, PointPlaneXOR} {
		points := []PointPlaneSample{{1, 2, 1}, {1, 2, 1}, {1, 2, 2}, {4, 3, 3}, {-1, 2, 1}, {8, 2, 1}, {1, 5, 1}}
		plane, err := NewIndexedPointPlane(pointPlaneFixture(collision, &points))
		if err != nil {
			t.Fatal(err)
		}
		if len(plane.Touched()) != 0 {
			t.Fatal("constructor called the sampler")
		}
		if err := plane.Sample(); err != nil {
			t.Fatal(err)
		}
		want := byte(3)
		if collision == PointPlaneXOR {
			want = 2
		}
		if len(plane.Touched()) != 2 || plane.Touched()[0] != 17 || plane.Masks()[17] != want || plane.Masks()[28] != 3 {
			t.Fatal("collision state or deduplicated order changed", plane.Touched(), plane.Masks())
		}
		points[0], points[1], points[2] = PointPlaneSample{2, 1, 1}, PointPlaneSample{2, 1, 1}, PointPlaneSample{2, 1, 2}
		plane.Update(kit.Frame{})
		if plane.Masks()[17] != 0 || len(plane.Touched()) != 2 || plane.Masks()[10] != want {
			t.Fatal("old touched pixels were retained or membership was not reset")
		}
		if allocations := testing.AllocsPerRun(100, func() {
			plane.Sample()
		}); allocations != 0 {
			t.Fatal("sparse sampling allocated", allocations)
		}
		plane.Reset()
		if len(plane.Touched()) != 0 {
			t.Fatal("reset retained active pixels")
		}
		for _, mask := range plane.Masks() {
			if mask != 0 {
				t.Fatal("reset retained a mask")
			}
		}
		plane.Close()
		plane.Close()
		if plane.Sample() == nil || len(plane.Masks()) != 0 || len(plane.Touched()) != 0 {
			t.Fatal("closed component retained sampling storage")
		}
	}
}

func TestIndexedPointPlaneVisibilityPaletteAndMembershipWordEdges(t *testing.T) {
	colors := []color.NRGBA{{R: 9, A: 255}, {R: 200, G: 100, B: 50, A: 128}}
	config := IndexedPointPlaneConfig{Width: 65, Height: 2, Count: 6, Palette: colors,
		Sample: func(index int) (PointPlaneSample, bool) {
			return PointPlaneSample{X: []int{0, 63, 64, 0, 63, 64}[index], Y: index / 3, Mask: 1}, index != 1
		}}
	plane, err := NewIndexedPointPlane(config)
	if err != nil {
		t.Fatal(err)
	}
	defer plane.Close()
	colors[1] = color.NRGBA{}
	if plane.paints[1][0] < .39 || plane.paints[1][3] < .5 {
		t.Fatal("palette was not copied or premultiplied")
	}
	plane.Sample()
	if len(plane.Touched()) != 5 || plane.Masks()[63] != 0 || plane.Masks()[64] != 1 || plane.Masks()[129] != 1 {
		t.Fatal("visibility or membership word boundaries changed")
	}
	newPalette := []color.NRGBA{{A: 255}}
	if allocations := testing.AllocsPerRun(100, func() { plane.SetPalette(newPalette) }); allocations != 0 {
		t.Fatal("palette update allocated", allocations)
	}
	if plane.paints[1] != ([4]float32{}) || plane.paints[0][3] != 1 || len(plane.Touched()) != 5 {
		t.Fatal("palette replacement retained colors or changed sampling")
	}
	if plane.SetPalette(nil) == nil || plane.SetPalette(make([]color.NRGBA, 257)) == nil {
		t.Fatal("invalid palette size was accepted")
	}
}

func TestIndexedPointPlaneValidatesAllStorageBudgets(t *testing.T) {
	for _, invalid := range []IndexedPointPlaneConfig{
		{}, {Width: 1, Height: 1}, {Width: 8193, Height: 1, Palette: []color.NRGBA{{}}},
		{Width: 8192, Height: 8192, Palette: []color.NRGBA{{}}},
		{Width: 1, Height: 1, Count: -1, Palette: []color.NRGBA{{}}},
		{Width: 1, Height: 1, Count: 1_000_001, Palette: []color.NRGBA{{}}},
		{Width: 1, Height: 1, Count: 1, Palette: []color.NRGBA{{}}},
		{Width: 1, Height: 1, Collision: 2, Palette: []color.NRGBA{{}}},
		{Width: 1, Height: 1, Palette: make([]color.NRGBA, 257)},
		{Width: 1, Height: 1, Palette: []color.NRGBA{{}}, Offset: image.Pt(1<<24+1, 0)},
		{Width: 1, Height: 1, Palette: []color.NRGBA{{}}, BatchTriangles: 1},
		{Width: 1, Height: 1, Palette: []color.NRGBA{{}}, BatchTriangles: 20001},
	} {
		if plane, err := NewIndexedPointPlane(invalid); err == nil {
			plane.Close()
			t.Fatal("invalid indexed plane configuration accepted", invalid)
		}
	}
	// A zero population is useful for optional layers and needs no sampler.
	empty, err := NewIndexedPointPlane(IndexedPointPlaneConfig{Width: 1, Height: 1, Palette: []color.NRGBA{{}}})
	if err != nil {
		t.Fatal(err)
	}
	defer empty.Close()
	if err := empty.Sample(); err != nil || len(empty.Touched()) != 0 {
		t.Fatal("empty plane was not usable", err)
	}
	var absent *IndexedPointPlane
	if absent.Sample() == nil {
		t.Fatal("nil sampler was accepted")
	}
	absent.Draw(nil)
	absent.Reset()
	absent.Close()
}
