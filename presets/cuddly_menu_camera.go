package presets

import "github.com/olivierh59500/democonstructionkit/motion"

// CuddlyMenuCamera keeps the authored bottom trim and centered sprite anchor
// while exposing viewport, world and sprite dimensions as editable geometry.
func CuddlyMenuCamera(viewW, viewH, worldW, worldH, spriteW, spriteH, bottomTrim int) motion.CameraFollowConfig {
	return motion.CameraFollowConfig{
		ViewportW: viewW, ViewportH: viewH,
		WorldW: worldW, WorldH: worldH - bottomTrim,
		AnchorX: viewW/2 - spriteW/2, AnchorY: viewH/2 - spriteH/2,
	}
}
