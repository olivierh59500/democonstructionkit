package composite

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
)

type staticTourSource struct{}

func (staticTourSource) Update() error      { return nil }
func (staticTourSource) Draw(*ebiten.Image) {}

func TestSceneTourCanvasStorageMask(t *testing.T) {
	camera, err := motion.NewCameraTour(motion.CameraTourConfig{
		ViewportWidth: 8, ViewportHeight: 8, StepSeconds: 1.0 / 60,
		Segments: []motion.TourSegment{{Name: "still", Duration: 1,
			From: motion.TourPose{Zoom: 1}, To: motion.TourPose{Zoom: 1},
			Direct: 0, VisibleMask: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	config := SceneTourConfig{Camera: camera, Sources: []TourSource{staticTourSource{}},
		TileOrigins: []image.Point{{}}, TileWidth: 8, TileHeight: 8,
		ViewportWidth: 8, ViewportHeight: 8, UnmanagedMask: 2}
	if _, err := NewSceneTour(config); err == nil {
		t.Fatal("accepted an unmanaged bit beyond the source count")
	}
	config.UnmanagedMask = 1
	tour, err := NewSceneTour(config)
	if err != nil {
		t.Fatal(err)
	}
	if tour.Canvases()[0] == nil || !tour.owned[0] {
		t.Fatal("the requested canvas was not allocated")
	}
	tour.Close()
}
