package presets

import (
	"image/color"

	"github.com/olivierh59500/democonstructionkit/effects"
)

// BilizirCube supplies the six pink face colors, outlined borders and authored
// depth order. The returned configuration is independent: another demo can edit
// its camera, material, size and line width without duplicating a renderer.
func BilizirCube(size float64) effects.SolidCubeConfig {
	c := effects.DefaultSolidCubeConfig(size)
	c.Perspective = 200
	c.FarToNear = false
	c.FaceColors = [6]color.RGBA{
		{R: 255, G: 80, B: 160, A: 255},
		{R: 255, G: 120, B: 200, A: 255},
		{R: 200, G: 60, B: 140, A: 255},
		{R: 255, G: 100, B: 180, A: 255},
		{R: 220, G: 80, B: 160, A: 255},
		{R: 255, G: 140, B: 200, A: 255},
	}
	for i, face := range c.FaceColors {
		// Retain the original palette's byte arithmetic for exact raster colors.
		c.EdgeColors[i] = color.RGBA{R: face.R * 3 / 4, G: face.G * 3 / 4, B: face.B * 3 / 4, A: 255}
	}
	return c
}
