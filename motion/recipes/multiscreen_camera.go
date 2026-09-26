package recipes

import "github.com/olivierh59500/democonstructionkit/motion"

const (
	MultiscreenView1 = iota
	MultiscreenMove12
	MultiscreenView2
	MultiscreenMove23
	MultiscreenView3
	MultiscreenMove34
	MultiscreenView4
	MultiscreenZoomOut
	MultiscreenOverview
	MultiscreenLoop
)

// MultiscreenCameraTour keeps four screens running while the view moves around
// a 2x2 world. Durations, positions, easing, source masks and direct fixed
// views are editable before constructing the controller.
func MultiscreenCameraTour() motion.CameraTourConfig {
	offset := motion.TourCameraOffset
	center := motion.TourWorldCenter
	cubic := motion.TourCubicInOut
	pose := func(x, y, zoom float64) motion.TourPose { return motion.TourPose{X: x, Y: y, Zoom: zoom} }
	return motion.CameraTourConfig{
		ViewportWidth: 800, ViewportHeight: 600, StepSeconds: 1.0 / 60.0,
		Segments: []motion.TourSegment{
			{Name: "phenomena", Duration: 7, From: pose(0, 0, 1), To: pose(0, 0, 1), Coordinates: offset, Direct: 0, VisibleMask: 1},
			{Name: "to-tcb", Duration: 4, From: pose(0, 0, 1), To: pose(800, 0, 1), Coordinates: offset, Ease: cubic, Direct: -1, VisibleMask: 1 | 2},
			{Name: "tcb", Duration: 7, From: pose(800, 0, 1), To: pose(800, 0, 1), Coordinates: offset, Direct: 1, VisibleMask: 2},
			{Name: "to-coco", Duration: 4, From: pose(800, 0, 1), To: pose(800, 600, 1), Coordinates: offset, Ease: cubic, Direct: -1, VisibleMask: 2 | 4},
			{Name: "coco", Duration: 7, From: pose(800, 600, 1), To: pose(800, 600, 1), Coordinates: offset, Direct: 2, VisibleMask: 4},
			{Name: "to-viva", Duration: 4, From: pose(800, 600, 1), To: pose(0, 600, 1), Coordinates: offset, Ease: cubic, Direct: -1, VisibleMask: 4 | 8},
			{Name: "viva", Duration: 7, From: pose(0, 600, 1), To: pose(0, 600, 1), Coordinates: offset, Direct: 3, VisibleMask: 8},
			{Name: "zoom-out", Duration: 4, From: pose(400, 900, 1), To: pose(800, 600, .5), Coordinates: center, Ease: cubic, Direct: -1, VisibleMask: 15},
			{Name: "overview", Duration: 7, From: pose(800, 600, .5), To: pose(800, 600, .5), Coordinates: center, Direct: -1, VisibleMask: 15},
			{Name: "loop", Duration: 4, From: pose(800, 600, .5), To: pose(400, 300, 1), Coordinates: center, Ease: cubic, Direct: -1, VisibleMask: 15},
		},
	}
}
