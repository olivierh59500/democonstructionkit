package presets

import (
	"github.com/olivierh59500/democonstructionkit/motion"
	motionrecipes "github.com/olivierh59500/democonstructionkit/motion/recipes"
)

const (
	MultiscreenView1    = motionrecipes.MultiscreenView1
	MultiscreenMove12   = motionrecipes.MultiscreenMove12
	MultiscreenView2    = motionrecipes.MultiscreenView2
	MultiscreenMove23   = motionrecipes.MultiscreenMove23
	MultiscreenView3    = motionrecipes.MultiscreenView3
	MultiscreenMove34   = motionrecipes.MultiscreenMove34
	MultiscreenView4    = motionrecipes.MultiscreenView4
	MultiscreenZoomOut  = motionrecipes.MultiscreenZoomOut
	MultiscreenOverview = motionrecipes.MultiscreenOverview
	MultiscreenLoop     = motionrecipes.MultiscreenLoop
)

// MultiscreenCameraTour exposes the pure camera recipe with other scene presets.
func MultiscreenCameraTour() motion.CameraTourConfig {
	return motionrecipes.MultiscreenCameraTour()
}
