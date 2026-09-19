package effects

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/geometry"
)

func TestCRTCompiles(t *testing.T) {
	c, err := NewCRT(kit.Func{}, 16, 16, CRTConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if err = c.Close(); err != nil {
		t.Fatal(err)
	}
}
func TestMeshClippingAndStableAbsoluteSampling(t *testing.T) {
	m, err := NewMesh(Cube(2, geometry.Vec2{X: 1, Y: 1}, color.NRGBA{255, 255, 255, 255}), nil, geometry.Camera{Focal: 100, Near: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	m.Transform.Position.Z = 2
	m.Animate = func(s float64) Transform {
		return Transform{Position: geometry.Vec3{Z: 2}, Rotation: geometry.Vec3{Y: s}, Scale: 1}
	}
	if err = m.Update(kit.Frame{Time: .5}); err != nil {
		t.Fatal(err)
	}
	first := append([]geometry.Vec3(nil), m.points...)
	m.Update(kit.Frame{Time: 99})
	m.Update(kit.Frame{Time: .5})
	for i, p := range m.points {
		if p != first[i] {
			t.Fatal("sampling depends on previous updates")
		}
	}
}
func TestScrollerRejectsAmbiguousConfiguration(t *testing.T) {
	atlas := ebiten.NewImage(8, 8)
	defer atlas.Deallocate()
	f, err := font.NewGrid(font.Grid{Bounds: atlas.Bounds(), Cell: image.Pt(8, 8), Columns: 1, Order: "A"})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []ScrollConfig{{Width: 100, Height: 20, Message: "A\nA"}, {Width: 100, Height: 20, Message: "A", Gap: -1}, {Width: 100, Height: 20, Message: "A", Scale: -1}} {
		if _, err := NewScroller(atlas, f, c); err == nil {
			t.Fatal(c)
		}
	}
}
func TestPixelMappingAndPaletteTransparency(t *testing.T) {
	src := image.NewRGBA(image.Rect(4, 7, 6, 8))
	src.Set(4, 7, color.NRGBA{R: 255, A: 128})
	src.Set(5, 7, color.White)
	dst := image.NewRGBA(image.Rect(0, 0, 2, 1))
	SampleMap(src, true, func(x, y, s float64) (float64, float64) { return x - 1, y })(dst, 0)
	if dst.RGBAAt(0, 0) != (color.RGBA{255, 255, 255, 255}) || dst.RGBAAt(1, 0).A != 128 {
		t.Fatal(dst.Pix)
	}
	s := Indexed{Width: 2, Height: 1, Indices: []byte{0, 1}, Palette: []color.NRGBA{{}, {R: 255, A: 128}}}
	s.Render(dst, 0)
	if dst.RGBAAt(0, 0).A != 0 || dst.RGBAAt(1, 0).R != 128 {
		t.Fatal(dst.Pix)
	}
}
