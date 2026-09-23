package presets

import "github.com/olivierh59500/democonstructionkit/effects"

// TeamG1TexturedCube preserves the authored near-first face order and camera.
// The returned configuration can use any image or live effect surface.
func TeamG1TexturedCube() effects.TexturedCubeConfig {
	c := effects.DefaultTexturedCubeConfig(200)
	c.X, c.Y, c.FarToNear = 320, 200, false
	c.FrontClockwise = true
	return c
}
