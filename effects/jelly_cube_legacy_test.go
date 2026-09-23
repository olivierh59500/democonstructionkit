package effects

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"image/color"
	"math"
)

// Frozen pre-extraction arithmetic, motion and projection form an independent
// fidelity oracle. No production controller or renderer is reused here.
const (
	stCanvasWidth  = 640
	stCanvasHeight = 400
	rotationSpeed  = .05
	zoomSpeed      = .01
	posSpeed       = .014
)
const (
	// Cube rotation modes
	rotationModeNormal = iota
	rotationModeTumble
	rotationModePulsate
	rotationModeSwing
	rotationModeBounce
	rotationModeTotal // Total number of modes
)

// Color definitions for 3D cube faces
var (
	col0 = color.RGBA{0xE0, 0xA0, 0xC0, 0xFF}
	col1 = color.RGBA{0xE0, 0x60, 0xC0, 0xFF}
	col2 = color.RGBA{0xE0, 0xE0, 0xE0, 0xFF}

	blackFill      color.Color = color.Black
	mainScrollFill color.Color = color.RGBA{0x00, 0x00, 0x60, 0xFF}
)

// Vector3 represents a 3D point in space
type Vector3 = geometry.Vec3

// Face represents a quad face with 4 vertices and a color
type Face struct {
	P1, P2, P3, P4 int
	Color          color.RGBA
}

// faceWithDepth is used for depth sorting
type faceWithDepth struct {
	face  Face
	depth float64
}

type jellyLegacy struct {
	smoothCubeTransitions                                                                       bool
	cubeHandoff                                                                                 geometry.Handoff
	cubePrevious, cubeIncoming                                                                  []Vector3
	cubeTick                                                                                    int
	cubeChanged                                                                                 bool
	swingOriginX, swingOriginZ                                                                  float64
	pos, zoom3d                                                                                 float64
	rotation                                                                                    Vector3
	vertices                                                                                    []Vector3
	faces                                                                                       []Face
	transformedVertices                                                                         []Vector3
	facesWithDepth                                                                              []faceWithDepth
	cubeVertices                                                                                []ebiten.Vertex
	cubeIndices                                                                                 []uint16
	rotationMode                                                                                int
	rotationTimer, rotationDuration, pulsePhase, swingAmplitude, bounceVelocity, bouncePosition float64
	scrollIteration                                                                             int
	stCanvas, whiteImg                                                                          *ebiten.Image
}

func newJellyLegacy() *jellyLegacy {
	g := &jellyLegacy{smoothCubeTransitions: true, rotationDuration: 300, swingAmplitude: 1}
	// Initialize 3D cube vertices - perfect cube with equal dimensions
	size := 80.0
	g.vertices = []Vector3{
		{X: -size, Y: -size, Z: -size}, // 0 - back bottom left
		{X: size, Y: -size, Z: -size},  // 1 - back bottom right
		{X: size, Y: size, Z: -size},   // 2 - back top right
		{X: -size, Y: size, Z: -size},  // 3 - back top left
		{X: -size, Y: -size, Z: size},  // 4 - front bottom left
		{X: size, Y: -size, Z: size},   // 5 - front bottom right
		{X: size, Y: size, Z: size},    // 6 - front top right
		{X: -size, Y: size, Z: size},   // 7 - front top left
	}

	// Initialize cube faces with proper winding order
	g.faces = []Face{
		{4, 5, 6, 7, col0}, // Front face
		{1, 0, 3, 2, col0}, // Back face
		{5, 1, 2, 6, col1}, // Right face
		{0, 4, 7, 3, col1}, // Left face
		{7, 6, 2, 3, col2}, // Top face
		{0, 1, 5, 4, col2}, // Bottom face
	}

	// Pre-allocate transformation buffers
	g.transformedVertices = make([]Vector3, len(g.vertices))
	g.facesWithDepth = make([]faceWithDepth, len(g.faces))
	g.cubeVertices = make([]ebiten.Vertex, 0, len(g.faces)*4)
	g.cubeIndices = make([]uint16, 0, len(g.faces)*6)
	g.cubePrevious = make([]Vector3, 8)
	g.cubeIncoming = make([]Vector3, 8)
	return g
}
func (g *jellyLegacy) update() error {
	g.sampleCube(g.cubePrevious)
	g.cubeChanged = false
	g.scrollIteration++
	g.pos += posSpeed
	// Update 3D cube after delay
	if g.scrollIteration > 25 {
		// Handle rotation mode changes
		g.rotationTimer++
		if g.rotationTimer >= g.rotationDuration {
			g.rotationTimer = 0
			// Smooth transition between modes
			g.rotationMode = (g.rotationMode + 1) % rotationModeTotal
			g.cubeChanged = true

			// Reset parameters based on new mode
			switch g.rotationMode {
			case rotationModeBounce:
				g.bounceVelocity = 0.08
				g.bouncePosition = 0
			case rotationModeSwing:
				// Adjust initial rotation to avoid jumps
				g.swingOriginX = g.rotation.X
				g.swingOriginZ = g.rotation.Z
				if !g.smoothCubeTransitions {
					g.swingOriginX, g.swingOriginZ = 0, 0
				}
				g.swingAmplitude = 1.0
			case rotationModePulsate:
				// Start pulse phase based on current rotation to avoid jumps
				g.pulsePhase = math.Atan2(g.rotation.Y, g.rotation.X)
			case rotationModeNormal:
				// Continue from current position
				// No reset needed
			}
		}

		// Apply different movements based on mode
		switch g.rotationMode {
		case rotationModeNormal:
			// Standard rotation (existing)
			g.rotation.X += rotationSpeed
			g.rotation.Y += rotationSpeed
			g.rotation.Z -= rotationSpeed

		case rotationModeTumble:
			// Chaotic rotation with changing speed
			speedVar := math.Sin(g.rotationTimer * 0.02)
			g.rotation.X += rotationSpeed * (1 + speedVar)
			g.rotation.Y += rotationSpeed * (1.5 - speedVar*0.5)
			g.rotation.Z -= rotationSpeed * (0.5 + speedVar*0.5)

		case rotationModePulsate:
			// Rotation with pulsation
			g.pulsePhase += 0.05
			pulse := 1.0 + 0.3*math.Sin(g.pulsePhase)
			g.rotation.X += rotationSpeed * pulse
			g.rotation.Y += rotationSpeed * 0.7 * pulse
			g.rotation.Z -= rotationSpeed * 0.3

		case rotationModeSwing:
			// Pendulum swing
			swing := math.Sin(g.rotationTimer*0.03) * g.swingAmplitude
			g.rotation.X = g.swingOriginX + swing*0.8
			g.rotation.Y += rotationSpeed * 0.5
			g.rotation.Z = g.swingOriginZ + swing*0.4
			g.swingAmplitude *= 0.998 // Slower damping

		case rotationModeBounce:
			// Bounce effect
			g.bounceVelocity -= 0.001 // Reduced gravity
			g.bouncePosition += g.bounceVelocity

			// Limit descent to stay visible
			if g.bouncePosition < -0.3 {
				g.bouncePosition = -0.3
				g.bounceVelocity = math.Abs(g.bounceVelocity) * 0.85 // Bounce with energy loss
			}

			g.rotation.X += rotationSpeed * 0.3
			g.rotation.Y += rotationSpeed * (1 + math.Max(0, g.bouncePosition))
			g.rotation.Z += rotationSpeed * 0.1
		}

		// Zoom in 3D cube (existing)
		if g.zoom3d < 1 {
			g.zoom3d += zoomSpeed
			if g.zoom3d > 1 {
				g.zoom3d = 1
			}
		}
	}
	g.cubeTick++
	if g.cubeChanged && g.smoothCubeTransitions {
		g.sampleRawCube(g.cubeIncoming)
		if err := g.cubeHandoff.Begin(g.cubePrevious, g.cubeIncoming, float64(g.cubeTick)/60, .75); err != nil {
			return err
		}
	}

	return nil
}

// This independent reference retains the original arithmetic as a fidelity oracle.
func (g *jellyLegacy) sampleRawCube(dst []Vector3) {
	// Time factor for animation
	time := g.pos * 3.0

	// Adjust parameters based on mode
	var extraScale float64 = 1.0
	var extraOffsetY float64 = 0
	var twistFactor float64 = 0

	switch g.rotationMode {
	case rotationModePulsate:
		// Pulsing zoom effect
		extraScale = 1.0 + 0.2*math.Sin(g.pulsePhase)
	case rotationModeBounce:
		// Vertical offset for bounce (limited)
		extraOffsetY = g.bouncePosition * 50 // Reduced from 100 to 50
	case rotationModeTumble:
		// Twist effect
		twistFactor = math.Sin(g.rotationTimer*0.01) * 0.5
	}

	// Pre-calculate sin/cos for rotation
	sinX, cosX := math.Sincos(g.rotation.X)
	sinY, cosY := math.Sincos(g.rotation.Y)
	sinZ, cosZ := math.Sincos(g.rotation.Z)

	// Pre-calculate common animation values
	squashFactor := 1.0 + 0.15*math.Sin(time*2.0)
	stretchFactor := 1.0 + 0.15*math.Cos(time*2.0)
	sinTime25, cosTime25 := math.Sincos(time * 2.5)
	secondaryBounce := sinTime25 + 0.5*math.Sin(time*5.0)
	secondaryOffsetY := cosTime25*8.0 + extraOffsetY
	secondaryOffsetZ := math.Sin(time*3.7) * 4.0
	deformScale := 1.0
	if g.rotationMode == rotationModePulsate {
		deformScale += 0.3 * math.Sin(g.pulsePhase*2)
	}

	// Apply transformations to each vertex
	for i, vertex := range g.vertices {
		x, y, z := vertex.X, vertex.Y, vertex.Z

		// Apply extra scale
		x *= extraScale
		y *= extraScale
		z *= extraScale

		// Calculate jelly deformation (existing code)
		positionKey := vertex.X*0.01 + vertex.Y*0.02 + vertex.Z*0.03
		deformAmount := 25.0 * deformScale

		// Multiple wobble frequencies for complex motion
		wobbleX := math.Sin(time+positionKey*5.0) * deformAmount * 0.4
		wobbleX += math.Sin(time*2.1+positionKey*3.0) * deformAmount * 0.2

		wobbleY := math.Cos(time*1.3+positionKey*7.0) * deformAmount * 0.4
		wobbleY += math.Cos(time*1.7+positionKey*4.0) * deformAmount * 0.2

		wobbleZ := math.Sin(time*0.7+positionKey*3.0) * deformAmount * 0.3
		wobbleZ += math.Cos(time*1.9+positionKey*6.0) * deformAmount * 0.15

		// Apply deformation based on distance from center
		distFromCenter := math.Sqrt(x*x+y*y+z*z) / 80.0
		wobbleInfluence := 0.5 + distFromCenter*0.5

		x += wobbleX * wobbleInfluence
		y += wobbleY * wobbleInfluence
		z += wobbleZ * wobbleInfluence

		// Squash and stretch effect
		x *= squashFactor
		y *= stretchFactor
		z *= 1.0 / (squashFactor*stretchFactor*0.5 + 0.5)

		// Add ripple effect
		ripple := math.Sin(time*4.0+distFromCenter*10.0) * 5.0
		x += ripple * (vertex.Y / 80.0)
		y += ripple * (vertex.X / 80.0)

		// Add twist effect if applicable
		if twistFactor != 0 {
			angle := twistFactor * (vertex.Y / 80.0)
			sinAngle, cosAngle := math.Sincos(angle)
			newX := x*cosAngle - z*sinAngle
			newZ := x*sinAngle + z*cosAngle
			x, z = newX, newZ
		}

		// Rotate around X axis
		newY := y*cosX - z*sinX
		newZ := y*sinX + z*cosX
		y, z = newY, newZ

		// Rotate around Y axis
		newX := x*cosY + z*sinY
		newZ = -x*sinY + z*cosY
		x, z = newX, newZ

		// Rotate around Z axis
		newX = x*cosZ - y*sinZ
		newY = x*sinZ + y*cosZ

		// Add secondary wobble
		newX += secondaryBounce * 8.0
		newY += secondaryOffsetY
		newZ += secondaryOffsetZ

		dst[i] = Vector3{X: newX, Y: newY, Z: newZ}
	}

}

// sampleCube includes any active handoff without depending on Draw calls.
func (g *jellyLegacy) sampleCube(dst []Vector3) {
	g.sampleRawCube(dst)
	if g.smoothCubeTransitions {
		g.cubeHandoff.Apply(dst, dst, float64(g.cubeTick)/60)
	}
}

func (g *jellyLegacy) draw3DCube() {
	g.sampleCube(g.transformedVertices)
	g.drawCubeVertices()
}

func (g *jellyLegacy) drawCubeVertices() {
	// Calculate face depths
	for i, face := range g.faces {
		// Calculate average Z depth
		avgZ := (g.transformedVertices[face.P1].Z + g.transformedVertices[face.P2].Z +
			g.transformedVertices[face.P3].Z + g.transformedVertices[face.P4].Z) / 4.0
		g.facesWithDepth[i].face = face
		g.facesWithDepth[i].depth = avgZ
	}

	// Sort the six faces back to front. An insertion sort avoids the reflection
	// and heap escape caused by sort.Slice in this per-frame hot path.
	for i := 1; i < len(g.facesWithDepth); i++ {
		face := g.facesWithDepth[i]
		j := i
		for j > 0 && g.facesWithDepth[j-1].depth > face.depth {
			g.facesWithDepth[j] = g.facesWithDepth[j-1]
			j--
		}
		g.facesWithDepth[j] = face
	}

	// Draw faces
	centerX := float32(stCanvasWidth / 2)
	centerY := float32(stCanvasHeight / 2)
	g.cubeVertices = g.cubeVertices[:0]
	g.cubeIndices = g.cubeIndices[:0]

	// High FOV for minimal perspective
	fov := 2000.0

	for _, f := range g.facesWithDepth {
		face := f.face

		// Get transformed vertices
		v0 := g.transformedVertices[face.P1]
		v1 := g.transformedVertices[face.P2]
		v2 := g.transformedVertices[face.P3]
		v3 := g.transformedVertices[face.P4]

		// Project to 2D
		offset := 300.0

		scale0 := fov / (fov + v0.Z + offset)
		x0 := centerX + float32(v0.X*scale0)
		y0 := centerY + float32(v0.Y*scale0)

		scale1 := fov / (fov + v1.Z + offset)
		x1 := centerX + float32(v1.X*scale1)
		y1 := centerY + float32(v1.Y*scale1)

		scale2 := fov / (fov + v2.Z + offset)
		x2 := centerX + float32(v2.X*scale2)
		y2 := centerY + float32(v2.Y*scale2)

		scale3 := fov / (fov + v3.Z + offset)
		x3 := centerX + float32(v3.X*scale3)
		y3 := centerY + float32(v3.Y*scale3)

		// Slight expansion to avoid gaps
		expansion := float32(0.5)

		// Calculate face center
		centerFaceX := (x0 + x1 + x2 + x3) / 4.0
		centerFaceY := (y0 + y1 + y2 + y3) / 4.0

		// Expand vertices slightly
		x0 += (x0 - centerFaceX) * expansion / 100.0
		y0 += (y0 - centerFaceY) * expansion / 100.0
		x1 += (x1 - centerFaceX) * expansion / 100.0
		y1 += (y1 - centerFaceY) * expansion / 100.0
		x2 += (x2 - centerFaceX) * expansion / 100.0
		y2 += (y2 - centerFaceY) * expansion / 100.0
		x3 += (x3 - centerFaceX) * expansion / 100.0
		y3 += (y3 - centerFaceY) * expansion / 100.0

		// Apply the former intermediate-canvas zoom directly to the vertices.
		zoom := float32(g.zoom3d)
		x0 = centerX + (x0-centerX)*zoom
		y0 = centerY + (y0-centerY)*zoom
		x1 = centerX + (x1-centerX)*zoom
		y1 = centerY + (y1-centerY)*zoom
		x2 = centerX + (x2-centerX)*zoom
		y2 = centerY + (y2-centerY)*zoom
		x3 = centerX + (x3-centerX)*zoom
		y3 = centerY + (y3-centerY)*zoom

		red := float32(face.Color.R) / 255
		green := float32(face.Color.G) / 255
		blue := float32(face.Color.B) / 255
		alpha := float32(face.Color.A) / 255
		base := uint16(len(g.cubeVertices))
		g.cubeVertices = append(g.cubeVertices,
			ebiten.Vertex{DstX: x0, DstY: y0, SrcX: 0, SrcY: 0, ColorR: red, ColorG: green, ColorB: blue, ColorA: alpha},
			ebiten.Vertex{DstX: x1, DstY: y1, SrcX: 1, SrcY: 0, ColorR: red, ColorG: green, ColorB: blue, ColorA: alpha},
			ebiten.Vertex{DstX: x2, DstY: y2, SrcX: 1, SrcY: 1, ColorR: red, ColorG: green, ColorB: blue, ColorA: alpha},
			ebiten.Vertex{DstX: x3, DstY: y3, SrcX: 0, SrcY: 1, ColorR: red, ColorG: green, ColorB: blue, ColorA: alpha},
		)
		g.cubeIndices = append(g.cubeIndices,
			base, base+1, base+2,
			base, base+2, base+3,
		)
	}

	if len(g.cubeIndices) > 0 {
		g.stCanvas.DrawTriangles(g.cubeVertices, g.cubeIndices, g.whiteImg, nil)
	}
}
