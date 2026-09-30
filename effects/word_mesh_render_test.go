//go:build dck_gpu_rendercheck

package effects

import (
	"bytes"
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func wordTestMatrix(t *testing.T) *motion.WordEulerMatrix {
	t.Helper()
	m, err := motion.NewWordEulerMatrix(motion.WordEulerMatrixConfig{Sines: []int16{0, 32767, 0, -32767}, Period: 4, Quantum: 1, Quarter: 1, ProductShift: 6, ThirdAxisShift: 7})
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func wordTestConfig(t *testing.T, points []motion.WrappedPoint, faces []WordFace) WordMeshConfig {
	return WordMeshConfig{Models: []WordMeshModel{{Points: points, Faces: faces}}, Matrix: wordTestMatrix(t), Projection: geometry.WordProjectionConfig{DepthShift: 9}}
}
func wordTestPixels(t *testing.T, dst *ebiten.Image, want func(int, int) [4]byte) {
	t.Helper()
	bounds := dst.Bounds()
	pixels := make([]byte, bounds.Dx()*bounds.Dy()*4)
	dst.ReadPixels(pixels)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			v := want(x, y)
			at := ((y-bounds.Min.Y)*bounds.Dx() + x - bounds.Min.X) * 4
			if !bytes.Equal(pixels[at:at+4], v[:]) {
				t.Fatalf("pixel %d,%d got%v want%v", x, y, pixels[at:at+4], v)
			}
		}
	}
}

func TestWordMeshGPUContoursParityInstancesAndClock(t *testing.T) {
	points := []motion.WrappedPoint{{X: 2, Y: 2}, {X: 12, Y: 2}, {X: 12, Y: 12}, {X: 2, Y: 12}, {X: 4, Y: 4}, {X: 8, Y: 4}, {X: 8, Y: 8}, {X: 4, Y: 8}}
	faces := []WordFace{{Contours: [][]int{{0, 1, 2, 3, 0}, {4, 5, 6, 7, 4}}}}
	c := wordTestConfig(t, points, faces)
	c.FillRule = ebiten.FillRuleEvenOdd
	c.Color = func(kit.Frame) color.NRGBA { return color.NRGBA{B: 255, A: 255} }
	m, err := NewWordMesh(c)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	points[0].X = 99
	faces[0].Contours[0][0] = 7
	if err := m.SetPose(WordMeshPose{Depth: 255, Visible: true}); err != nil {
		t.Fatal(err)
	}
	dst := ebiten.NewImage(16, 16)
	defer dst.Deallocate()
	m.Draw(dst)
	wordTestPixels(t, dst, func(x, y int) [4]byte {
		if x >= 2 && x < 12 && y >= 2 && y < 12 && !(x >= 4 && x < 8 && y >= 4 && y < 8) {
			return [4]byte{0, 0, 255, 255}
		}
		return [4]byte{}
	})
	if err := m.SetInstanceCount(2); err != nil {
		t.Fatal(err)
	}
	dst.Clear()
	m.Draw(dst)
	wordTestPixels(t, dst, func(int, int) [4]byte { return [4]byte{} })
	if err := m.SetInstanceCount(1); err != nil {
		t.Fatal(err)
	}
	dst.Clear()
	m.Draw(dst)
	a := make([]byte, 16*16*4)
	dst.ReadPixels(a)
	dst.Clear()
	m.Draw(dst)
	b := make([]byte, len(a))
	dst.ReadPixels(b)
	if !bytes.Equal(a, b) {
		t.Fatal("draw changed geometry or clock")
	}
	if err := m.SetPose(WordMeshPose{Visible: false}); err != nil {
		t.Fatal(err)
	}
	dst.Clear()
	m.Draw(dst)
	wordTestPixels(t, dst, func(int, int) [4]byte { return [4]byte{} })
}

func TestWordMeshGPUClippingCullingUVAndBorrowedMaterial(t *testing.T) {
	texture := ebiten.NewImage(1, 16)
	defer texture.Deallocate()
	pixels := make([]byte, 16*4)
	for y := range 16 {
		pixels[y*4], pixels[y*4+3] = byte(y*16), 255
	}
	texture.WritePixels(pixels)
	points := []motion.WrappedPoint{{X: -2, Y: -2}, {X: 10, Y: -2}, {X: 10, Y: 10}, {X: -2, Y: 10}}
	c := wordTestConfig(t, points, []WordFace{{Contours: [][]int{{0, 1, 2, 3}}, Material: 2}})
	c.Textures = []*ebiten.Image{texture}
	c.CullBackFaces = true
	c.FaceClip = &[4]float64{0, 0, 8, 8}
	c.ClipVertexLimit = 8
	c.Offset = geometry.Vec2{X: 3, Y: 2}
	c.MaxInstances = 2
	c.InstanceCount = 2
	c.Instance = func(index int, _ kit.Frame) WordMeshInstance {
		return WordMeshInstance{Offset: geometry.Vec2{X: float64(index * 9)}}
	}
	c.UV = func(s WordMaterialSample) geometry.Vec2 { return geometry.Vec2{X: .5, Y: s.Screen.Y} }
	m, err := NewWordMesh(c)
	if err != nil {
		t.Fatal(err)
	}
	if err = m.SetPose(WordMeshPose{Depth: 255, Visible: true}); err != nil {
		t.Fatal(err)
	}
	dst := ebiten.NewImage(24, 16)
	defer dst.Deallocate()
	view := dst.SubImage(image.Rect(2, 1, 22, 14)).(*ebiten.Image)
	m.Draw(view)
	wordTestPixels(t, dst, func(x, y int) [4]byte {
		if y >= 2 && y < 10 && (x >= 3 && x < 11 || x >= 12 && x < 20) {
			return [4]byte{byte(y * 16), 0, 0, 255}
		}
		return [4]byte{}
	})
	if err = m.Close(); err != nil {
		t.Fatal(err)
	}
	m.Close()
	dst.Clear()
	m.Draw(dst)
	wordTestPixels(t, dst, func(int, int) [4]byte { return [4]byte{} })
	texture.Fill(color.White)
	dst.DrawImage(texture, nil)
	read := make([]byte, 24*16*4)
	dst.ReadPixels(read)
	if read[0] != 255 || read[3] != 255 {
		t.Fatal("close destroyed borrowed material")
	}
	// Reverse winding suppresses a face before clipping and material generation.
	c.Models[0].Faces[0].Contours[0] = []int{0, 3, 2, 1}
	c.InstanceCount = 1
	m, err = NewWordMesh(c)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if err = m.SetPose(WordMeshPose{Depth: 255, Visible: true}); err != nil {
		t.Fatal(err)
	}
	dst.Clear()
	m.Draw(dst)
	wordTestPixels(t, dst, func(int, int) [4]byte { return [4]byte{} })
}

func TestWordMeshGPUPerFaceMaterialsAndDestinationCrop(t *testing.T) {
	red, blue := ebiten.NewImage(1, 1), ebiten.NewImage(1, 1)
	defer red.Deallocate()
	defer blue.Deallocate()
	red.Fill(color.NRGBA{R: 255, A: 255})
	blue.Fill(color.NRGBA{B: 255, A: 255})
	points := []motion.WrappedPoint{{X: 0, Y: 0}, {X: 6, Y: 0}, {X: 6, Y: 6}, {X: 0, Y: 6}, {X: 7, Y: 0}, {X: 13, Y: 0}, {X: 13, Y: 6}, {X: 7, Y: 6}}
	c := wordTestConfig(t, points, []WordFace{{Contours: [][]int{{0, 1, 2, 3}}}, {Contours: [][]int{{4, 5, 6, 7}}, Texture: 1}})
	c.Textures = []*ebiten.Image{red, blue}
	c.PerFaceBatch = true
	c.DestinationClip = image.Rect(2, 1, 12, 5)
	m, err := NewWordMesh(c)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if err = m.SetPose(WordMeshPose{Depth: 255, Visible: true}); err != nil {
		t.Fatal(err)
	}
	dst := ebiten.NewImage(16, 8)
	defer dst.Deallocate()
	m.Draw(dst)
	wordTestPixels(t, dst, func(x, y int) [4]byte {
		if y >= 1 && y < 5 {
			if x >= 2 && x < 6 {
				return [4]byte{255, 0, 0, 255}
			}
			if x >= 7 && x < 12 {
				return [4]byte{0, 0, 255, 255}
			}
		}
		return [4]byte{}
	})
	c.PerFaceBatch = false
	m2, err := NewWordMesh(c)
	if err != nil {
		t.Fatal(err)
	}
	defer m2.Close()
	m2.SetPose(WordMeshPose{Depth: 255, Visible: true})
	dst.Clear()
	m2.Draw(dst)
	if m2.Err() == nil {
		t.Fatal("accepted mixed whole-run textures")
	}
	wordTestPixels(t, dst, func(int, int) [4]byte { return [4]byte{} })
}

func TestWordMeshGPUBudgetsAndInvalidCoordinates(t *testing.T) {
	points := []motion.WrappedPoint{{X: 1, Y: 1}, {X: 5, Y: 1}, {X: 5, Y: 5}, {X: 1, Y: 5}}
	c := wordTestConfig(t, points, []WordFace{{Contours: [][]int{{0, 1, 2, 3}}}})
	for _, bad := range []WordMeshConfig{{}, {Models: c.Models, Matrix: c.Matrix, BatchTriangles: 1}, {Models: c.Models, Matrix: c.Matrix, ClipVertexLimit: 2}, {Models: c.Models, Matrix: c.Matrix, Offset: geometry.Vec2{X: math.NaN()}}} {
		if m, err := NewWordMesh(bad); err == nil {
			m.Close()
			t.Fatal("invalid config accepted")
		}
	}
	c.FillRule = ebiten.FillRuleEvenOdd
	c.BatchTriangles = 2
	c.InstanceCount = 2
	m, err := NewWordMesh(c)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	m.SetPose(WordMeshPose{Depth: 255, Visible: true})
	dst := ebiten.NewImage(8, 8)
	defer dst.Deallocate()
	m.Draw(dst)
	if m.Err() == nil {
		t.Fatal("parity budget accepted")
	}
	wordTestPixels(t, dst, func(int, int) [4]byte { return [4]byte{} })
	c.InstanceCount = 1
	c.UV = func(WordMaterialSample) geometry.Vec2 { return geometry.Vec2{X: math.Inf(1)} }
	m2, err := NewWordMesh(c)
	if err != nil {
		t.Fatal(err)
	}
	defer m2.Close()
	m2.SetPose(WordMeshPose{Depth: 255, Visible: true})
	m2.Draw(dst)
	if m2.Err() == nil {
		t.Fatal("invalid UV accepted")
	}
	wordTestPixels(t, dst, func(int, int) [4]byte { return [4]byte{} })
}
