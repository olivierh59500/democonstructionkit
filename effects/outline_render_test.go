//go:build dck_gpu_rendercheck

package effects

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/fidelity/ebiten/testutil"
	"github.com/olivierh59500/democonstructionkit/geometry"
)

func TestMeshOutlineCullingAndSharedPoseGPU(t *testing.T) {
	testutil.RequireGPU(t)
	mesh := Mesh{Points: []geometry.Vec3{{X: -1, Y: -1, Z: 4}, {X: 1, Y: -1, Z: 4}, {X: 1, Y: 1, Z: 4}, {X: -1, Y: 1, Z: 4}}}
	m, err := NewMesh(mesh, nil, geometry.Camera{Center: geometry.Vec2{X: 16, Y: 16}, Focal: 16, Near: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	m.CullBackFaces = true
	if err := m.SetOutline(MeshOutlineConfig{Faces: [][]int{{3, 2, 1, 0}}, Width: 2, Color: color.NRGBA{G: 255, A: 255}}); err != nil {
		t.Fatal(err)
	}
	calls := 0
	m.Deform = func(_ int, p geometry.Vec3, _ float64) geometry.Vec3 { calls++; return p }
	if err := m.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	dst := ebiten.NewImage(32, 32)
	defer dst.Deallocate()
	pixels := make([]byte, 32*32*4)
	m.DrawOutline(dst)
	m.DrawOutline(dst)
	dst.ReadPixels(pixels)
	if pixels[(16*32+12)*4+1] != 255 || pixels[(16*32+16)*4+3] != 0 || calls != 4 {
		t.Fatal("outline lost its border, added a triangulation diagonal or resampled deformation")
	}
	if err := m.SetOutline(MeshOutlineConfig{Faces: [][]int{{0, 1, 2, 3}}, Width: 2}); err != nil {
		t.Fatal(err)
	}
	dst.Clear()
	m.DrawOutline(dst)
	dst.ReadPixels(pixels)
	for _, pixel := range pixels {
		if pixel != 0 {
			t.Fatal("back face was not culled")
		}
	}
}

func TestWarpOutlineUsesCurrentMapWithoutDrawingSourceGPU(t *testing.T) {
	testutil.RequireGPU(t)
	updates, sourceDraws, maps := 0, 0, 0
	image := ebiten.NewImage(8, 8)
	defer image.Deallocate()
	image.Fill(color.NRGBA{R: 70, G: 30, A: 128})
	w, err := NewWarp(kit.Func{
		OnUpdate: func(kit.Frame) error { updates++; return nil },
		OnDraw:   func(dst *ebiten.Image) { sourceDraws++; dst.DrawImage(image, nil) },
	}, 8, 8, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	w.Map = func(x, y, seconds float64) geometry.Vec2 { maps++; return geometry.Vec2{X: 4 + x + seconds, Y: 4 + y} }
	white := ebiten.NewImage(1, 1)
	defer white.Deallocate()
	white.Fill(color.White)
	if err := w.SetOutline(WarpOutlineConfig{Width: 2, White: white, Blend: ebiten.BlendLighter}); err != nil {
		t.Fatal(err)
	}
	if err := w.Update(kit.Frame{Time: 2}); err != nil {
		t.Fatal(err)
	}
	dst := ebiten.NewImage(24, 20)
	defer dst.Deallocate()
	w.DrawOutline(dst)
	pixels := make([]byte, 24*20*4)
	dst.ReadPixels(pixels)
	if pixels[(8*24+6)*4+3] != 255 || pixels[(8*24+10)*4+3] != 0 || updates != 1 || sourceDraws != 0 || maps != 4 {
		t.Fatal("outline failed to reuse the mapped grid or advanced/drew its source")
	}
	// An additive outline must not leak its blend mode into the filled material.
	dst.Fill(color.NRGBA{R: 20, G: 40, B: 60, A: 255})
	w.Draw(dst)
	dst.ReadPixels(pixels)
	at := (8*24 + 10) * 4
	for i, want := range []int{45, 35, 30, 255} {
		got := int(pixels[at+i])
		if got < want-1 || got > want+1 {
			t.Fatalf("filled blend channel %d = %d, want about %d", i, got, want)
		}
	}
	w.Close()
	dst.Clear()
	dst.DrawImage(white, nil)
	dst.ReadPixels(pixels)
	if pixels[0] != 255 || pixels[3] != 255 {
		t.Fatal("closing warp deallocated the borrowed outline material")
	}
}

func TestMeshOutlineClipsOnlyVisibleEdgesGPU(t *testing.T) {
	testutil.RequireGPU(t)
	m, err := NewMesh(Mesh{Points: []geometry.Vec3{
		{X: -.5, Y: -.5, Z: .5}, {X: .5, Y: -.5, Z: 2},
		{X: .5, Y: .5, Z: 2}, {X: -.5, Y: .5, Z: .5},
	}}, nil, geometry.Camera{Center: geometry.Vec2{X: 16, Y: 16}, Focal: 8, Near: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if err := m.SetOutline(MeshOutlineConfig{Faces: [][]int{{0, 1, 2, 3}}, Width: 2}); err != nil {
		t.Fatal(err)
	}
	if err := m.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	dst := ebiten.NewImage(32, 32)
	defer dst.Deallocate()
	m.DrawOutline(dst)
	pixels := make([]byte, 32*32*4)
	dst.ReadPixels(pixels)
	if pixels[(16*32+18)*4+3] != 255 || pixels[(16*32+14)*4+3] != 0 || pixels[(16*32+16)*4+3] != 0 {
		t.Fatal("near clipping lost the visible edge, connected hidden endpoints or added an interior line")
	}
}
