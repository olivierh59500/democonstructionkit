package effects

import (
	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"image"
	"math"
	"reflect"
	"testing"
)

// Frozen pre-extraction implementation; the capture target records draw commands.
type legacyTextureVec struct{ X, Y, Z float64 }
type legacyTextureFace struct {
	P1, P2, P3, P4     int
	UV1, UV2, UV3, UV4 [2]float32
}
type legacyTextureDepth struct {
	faceIndex int
	depth     float64
}
type legacyTextureCanvas struct {
	vertices []ebiten.Vertex
	indices  []uint16
}

func (c *legacyTextureCanvas) Clear()                  { c.vertices = c.vertices[:0]; c.indices = c.indices[:0] }
func (c *legacyTextureCanvas) Bounds() image.Rectangle { return image.Rect(0, 0, 640, 400) }
func (c *legacyTextureCanvas) DrawTriangles(v []ebiten.Vertex, idx []uint16, _ *ebiten.Image, _ *ebiten.DrawTrianglesOptions) {
	base := uint16(len(c.vertices))
	c.vertices = append(c.vertices, v...)
	for _, i := range idx {
		c.indices = append(c.indices, base+i)
	}
}

type legacyTexturedCube struct {
	cubeVertices        [8]legacyTextureVec
	cubeFaces           [6]legacyTextureFace
	transformedVertices [8]legacyTextureVec
	faceDepths          [6]legacyTextureDepth
	cubeDrawVertices    [4]ebiten.Vertex
	cubeDrawIndices     [6]uint16
	cubeRotation        legacyTextureVec
	cubeCanvas          legacyTextureCanvas
	texture             *ebiten.Image
	cubeTrianglesOpt    ebiten.DrawTrianglesOptions
}

func (g *legacyTexturedCube) initCube() {
	// Cube vertices
	size := 100.0
	g.cubeVertices = [8]legacyTextureVec{
		{-size, -size, -size}, // 0
		{size, -size, -size},  // 1
		{size, size, -size},   // 2
		{-size, size, -size},  // 3
		{-size, -size, size},  // 4
		{size, -size, size},   // 5
		{size, size, size},    // 6
		{-size, size, size},   // 7
	}

	// Cube faces with texture coordinates
	g.cubeFaces = [6]legacyTextureFace{
		{4, 5, 6, 7, [2]float32{0, 0}, [2]float32{1, 0}, [2]float32{1, 1}, [2]float32{0, 1}}, // Front
		{1, 0, 3, 2, [2]float32{0, 0}, [2]float32{1, 0}, [2]float32{1, 1}, [2]float32{0, 1}}, // Back
		{5, 1, 2, 6, [2]float32{0, 0}, [2]float32{1, 0}, [2]float32{1, 1}, [2]float32{0, 1}}, // Right
		{0, 4, 7, 3, [2]float32{0, 0}, [2]float32{1, 0}, [2]float32{1, 1}, [2]float32{0, 1}}, // Left
		{7, 6, 2, 3, [2]float32{0, 0}, [2]float32{1, 0}, [2]float32{1, 1}, [2]float32{0, 1}}, // Top
		{0, 1, 5, 4, [2]float32{0, 0}, [2]float32{1, 0}, [2]float32{1, 1}, [2]float32{0, 1}}, // Bottom
	}
	g.cubeDrawIndices = [6]uint16{0, 1, 2, 0, 2, 3}
	for i := range g.cubeDrawVertices {
		g.cubeDrawVertices[i].ColorR = 1
		g.cubeDrawVertices[i].ColorG = 1
		g.cubeDrawVertices[i].ColorB = 1
		g.cubeDrawVertices[i].ColorA = 1
	}
}
func (g *legacyTexturedCube) drawTexturedCube() {
	g.cubeCanvas.Clear()

	sinX, cosX := math.Sincos(g.cubeRotation.X)
	sinY, cosY := math.Sincos(g.cubeRotation.Y)
	sinZ, cosZ := math.Sincos(g.cubeRotation.Z)

	// Transform vertices
	for i, v := range g.cubeVertices {
		x := v.X
		y := v.Y
		z := v.Z

		y2 := y*cosX - z*sinX
		z2 := y*sinX + z*cosX
		y = y2
		z = z2

		x2 := x*cosY + z*sinY
		z2 = -x*sinY + z*cosY
		x = x2

		x2 = x*cosZ - y*sinZ
		y2 = x*sinZ + y*cosZ

		g.transformedVertices[i] = legacyTextureVec{X: x2, Y: y2, Z: z2}
	}

	for i, face := range g.cubeFaces {
		g.faceDepths[i] = legacyTextureDepth{
			faceIndex: i,
			depth: (g.transformedVertices[face.P1].Z + g.transformedVertices[face.P2].Z +
				g.transformedVertices[face.P3].Z + g.transformedVertices[face.P4].Z) / 4,
		}
	}
	for i := 1; i < len(g.faceDepths); i++ {
		item := g.faceDepths[i]
		j := i
		for j > 0 && item.depth < g.faceDepths[j-1].depth {
			g.faceDepths[j] = g.faceDepths[j-1]
			j--
		}
		g.faceDepths[j] = item
	}

	// Draw faces
	centerX := float32(g.cubeCanvas.Bounds().Dx() / 2)
	centerY := float32(g.cubeCanvas.Bounds().Dy() / 2)
	textureWidth := float32(g.texture.Bounds().Dx())
	textureHeight := float32(g.texture.Bounds().Dy())
	const fov = 300.0

	for _, fd := range g.faceDepths {
		face := g.cubeFaces[fd.faceIndex]

		// Project vertices
		var screenPoints [4][2]float32
		pointIndices := [4]int{face.P1, face.P2, face.P3, face.P4}
		for i, pointIndex := range pointIndices {
			v := g.transformedVertices[pointIndex]
			scale := fov / (fov + v.Z + 300)
			screenPoints[i][0] = centerX + float32(v.X*scale)
			screenPoints[i][1] = centerY + float32(v.Y*scale)
		}

		// Check if face is visible (backface culling)
		v1x := screenPoints[1][0] - screenPoints[0][0]
		v1y := screenPoints[1][1] - screenPoints[0][1]
		v2x := screenPoints[2][0] - screenPoints[0][0]
		v2y := screenPoints[2][1] - screenPoints[0][1]

		if v1x*v2y-v1y*v2x < 0 {
			continue
		}

		uvs := [4][2]float32{face.UV1, face.UV2, face.UV3, face.UV4}
		for i := range g.cubeDrawVertices {
			g.cubeDrawVertices[i].DstX = screenPoints[i][0]
			g.cubeDrawVertices[i].DstY = screenPoints[i][1]
			g.cubeDrawVertices[i].SrcX = uvs[i][0] * textureWidth
			g.cubeDrawVertices[i].SrcY = uvs[i][1] * textureHeight
		}

		g.cubeCanvas.DrawTriangles(
			g.cubeDrawVertices[:],
			g.cubeDrawIndices[:],
			g.texture,
			&g.cubeTrianglesOpt,
		)
	}
}
func TestTexturedCubeMatchesEveryAuthoredVertex(t *testing.T) {
	texture := ebiten.NewImage(96, 80)
	defer texture.Deallocate()
	old := legacyTexturedCube{texture: texture}
	old.initCube()
	cfg := DefaultTexturedCubeConfig(200)
	cfg.FarToNear = false
	cfg.FrontClockwise = true
	cube, err := NewTexturedCube(texture, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer cube.Close()
	for i := 0; i < 3000; i++ {
		old.drawTexturedCube()
		vertices, indices := cube.Geometry(320, 200)
		if !reflect.DeepEqual(vertices, old.cubeCanvas.vertices) || !reflect.DeepEqual(indices, old.cubeCanvas.indices) {
			t.Fatalf("geometry changed at pose%d", i)
		}
		old.cubeRotation.X += .02
		old.cubeRotation.Y += .03
		old.cubeRotation.Z += .01
		cube.Rotate(.02, .03, .01)
	}
	if n := testing.AllocsPerRun(100, func() { cube.Geometry(320, 200) }); n != 0 {
		t.Fatal(n)
	}
}

func TestTexturedCubeRejectsDerivedOverflowWithoutChangingPose(t *testing.T) {
	texture := ebiten.NewImage(8, 8)
	defer texture.Deallocate()
	for _, change := range []func(*TexturedCubeConfig){
		func(c *TexturedCubeConfig) { c.Focal, c.Depth = 1e308, 1e308 },
		func(c *TexturedCubeConfig) { c.X = 1e100 },
		func(c *TexturedCubeConfig) { c.UV = &[6][4]geometry.Vec2{}; c.UV[0][0].X = 1e100 },
	} {
		c := DefaultTexturedCubeConfig(40)
		change(&c)
		if cube, err := NewTexturedCube(texture, c); err == nil {
			cube.Close()
			t.Fatal("invalid projected cube accepted")
		}
	}
	c := DefaultTexturedCubeConfig(40)
	c.AngularVelocity.X = 1e308
	cube, err := NewTexturedCube(texture, c)
	if err != nil {
		t.Fatal(err)
	}
	defer cube.Close()
	before := cube.Rotation
	if err := cube.Update(kit.Frame{Time: 100}); err == nil || cube.Rotation != before {
		t.Fatal("invalid animation changed the pose", err)
	}
}
