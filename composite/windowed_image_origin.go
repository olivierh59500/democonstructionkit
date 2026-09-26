package composite

import (
	"image"

	"github.com/olivierh59500/democonstructionkit/motion"
)

// WindowedImageWindow clips one source draw to an absolute destination area.
// Offset positions the source relative to the window's top-left corner.
type WindowedImageWindow struct {
	Clip   image.Rectangle
	Offset motion.Point
}

func windowedImageLocalOrigin(window WindowedImageWindow, phaseX, phaseY float64) motion.Point {
	return motion.Point{
		X: window.Offset.X + phaseX,
		Y: window.Offset.Y + phaseY,
	}
}
