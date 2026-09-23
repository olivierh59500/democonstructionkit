package effects

import "github.com/hajimehoshi/ebiten/v2"

// Geometry returns borrowed vertex/index buffers, valid until its next call.
// It does not advance motion, and its storage is allocated at construction.
func (c *JellyCube) Geometry() ([]ebiten.Vertex, []uint16) {
	// Calculate face depths
	for i, face := range c.faces {
		// Calculate average Z depth
		avgZ := (c.transformed[face.P1].Z + c.transformed[face.P2].Z +
			c.transformed[face.P3].Z + c.transformed[face.P4].Z) / 4.0
		c.depths[i].face = face
		c.depths[i].depth = avgZ
	}

	// Sort the six faces back to front. An insertion sort avoids the reflection
	// and heap escape caused by sort.Slice in this per-frame hot path.
	for i := 1; i < len(c.depths); i++ {
		face := c.depths[i]
		j := i
		for j > 0 && c.depths[j-1].depth > face.depth {
			c.depths[j] = c.depths[j-1]
			j--
		}
		c.depths[j] = face
	}

	// Draw faces
	centerX := float32(c.config.X)
	centerY := float32(c.config.Y)
	vertices := c.drawVertices[:0]
	indices := c.drawIndices[:0]

	// High FOV for minimal perspective
	fov := c.config.CameraFOV

	for _, f := range c.depths {
		face := f.face

		// Get transformed vertices
		v0 := c.transformed[face.P1]
		v1 := c.transformed[face.P2]
		v2 := c.transformed[face.P3]
		v3 := c.transformed[face.P4]

		// Project to 2D
		offset := c.config.CameraOffset
		if fov+v0.Z+offset <= 0 || fov+v1.Z+offset <= 0 || fov+v2.Z+offset <= 0 || fov+v3.Z+offset <= 0 {
			continue
		}

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
		zoom := float32(c.zoom)
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
		base := uint16(len(vertices))
		vertices = append(vertices,
			ebiten.Vertex{DstX: x0, DstY: y0, SrcX: 0, SrcY: 0, ColorR: red, ColorG: green, ColorB: blue, ColorA: alpha},
			ebiten.Vertex{DstX: x1, DstY: y1, SrcX: 1, SrcY: 0, ColorR: red, ColorG: green, ColorB: blue, ColorA: alpha},
			ebiten.Vertex{DstX: x2, DstY: y2, SrcX: 1, SrcY: 1, ColorR: red, ColorG: green, ColorB: blue, ColorA: alpha},
			ebiten.Vertex{DstX: x3, DstY: y3, SrcX: 0, SrcY: 1, ColorR: red, ColorG: green, ColorB: blue, ColorA: alpha},
		)
		indices = append(indices,
			base, base+1, base+2,
			base, base+2, base+3,
		)
	}

	return vertices, indices
}

// Draw composites the current cube without advancing its animation.
func (c *JellyCube) Draw(dst *ebiten.Image) {
	if dst == nil || c.closed || (c.delay > 0 && c.tick <= c.delay) {
		return
	}
	vertices, indices := c.Geometry()
	if len(indices) > 0 {
		dst.DrawTriangles(vertices, indices, c.texture, nil)
	}
}
