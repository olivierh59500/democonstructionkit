package presets

// MultiscreenCompositeShaderSource samples four retained 800x600 tiles in
// source order: upper left, upper right, lower right, lower left. The generic
// SceneTour renderer also accepts another shader or its DrawImage fallback.
const MultiscreenCompositeShaderSource = `//kage:unit pixels

package main

var CameraCenter vec2
var CameraZoom float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	screenPos := srcPos - imageSrc0Origin()
	worldPos := (screenPos - vec2(400, 300)) / CameraZoom + CameraCenter
	if worldPos.x < 0 || worldPos.y < 0 || worldPos.x >= 1600 || worldPos.y >= 1200 {
		return vec4(0, 0, 0, 1)
	}

	sourceOrigin := imageSrc0Origin()
	if worldPos.y < 600 {
		if worldPos.x < 800 {
			return imageSrc0UnsafeAt(sourceOrigin + worldPos)
		}
		return imageSrc1UnsafeAt(sourceOrigin + worldPos - vec2(800, 0))
	}

	if worldPos.x < 800 {
		return imageSrc3UnsafeAt(sourceOrigin + worldPos - vec2(0, 600))
	}
	return imageSrc2UnsafeAt(sourceOrigin + worldPos - vec2(800, 600))
}
`
