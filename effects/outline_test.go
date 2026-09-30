package effects

import (
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
)

func TestMeshOutlineOwnsTopologyAndReusesSolidMaterial(t *testing.T) {
	m, err := NewMesh(Cube(2, geometry.Vec2{X: 1, Y: 1}, color.NRGBA{A: 255}), nil, geometry.Camera{Focal: 20, Near: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	faces := CubeFaces()
	if err := m.SetOutline(MeshOutlineConfig{Faces: faces}); err != nil {
		t.Fatal(err)
	}
	faces[0][0] = 7
	if m.outline.points[m.outline.faces[0][0]] != 0 || len(m.outline.points) != 8 || m.outline.white != m.texture || m.outline.owned {
		t.Fatal("outline topology or solid-pixel ownership is incorrect")
	}
	for _, config := range []MeshOutlineConfig{
		{Faces: [][]int{{0, 1}}}, {Faces: [][]int{{0, 1, 8}}}, {Width: -1}, {Width: math.NaN()},
	} {
		if err := m.SetOutline(config); err == nil {
			t.Fatalf("accepted invalid outline %+v", config)
		}
	}
	if len(m.outline.faces) != 6 {
		t.Fatal("invalid update replaced the prepared topology")
	}
	if err := m.SetOutline(MeshOutlineConfig{}); err != nil || len(m.outline.faces) != 12 {
		t.Fatal("default triangle boundaries were not prepared")
	}
	m.Close()
	m.Close()
	if err := m.SetOutline(MeshOutlineConfig{Faces: CubeFaces()}); err == nil {
		t.Fatal("configured a closed mesh")
	}
}

func TestWarpOutlineRejectsBadStylesWithoutResettingSource(t *testing.T) {
	updates := 0
	w, err := NewWarp(kit.Func{OnUpdate: func(kit.Frame) error { updates++; return nil }}, 16, 16, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	if err := w.SetOutline(WarpOutlineConfig{}); err != nil {
		t.Fatal(err)
	}
	white := ebiten.NewImage(2, 1)
	defer white.Deallocate()
	for _, config := range []WarpOutlineConfig{{Width: -1}, {Width: math.Inf(1)}, {White: white}} {
		if err := w.SetOutline(config); err == nil {
			t.Fatalf("accepted %+v", config)
		}
	}
	if updates != 0 || w.Frame.Time != 0 {
		t.Fatal("style configuration advanced the source")
	}
	w.Close()
	w.Close()
	if err := w.SetOutline(WarpOutlineConfig{}); err == nil {
		t.Fatal("configured a closed warp")
	}
}
