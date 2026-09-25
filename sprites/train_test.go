package sprites

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestTrainCombinesRasterImagesWithExactBounceState(t *testing.T) {
	images := []*ebiten.Image{ebiten.NewImage(2, 2), ebiten.NewImage(2, 2), ebiten.NewImage(2, 2)}
	for _, image := range images {
		defer image.Deallocate()
	}
	train, err := NewTrain(TrainConfig{
		Images: images, ScaleX: 390,
		Y: TrainAxis{Offset: 60, Bounce: &motion.BounceBankConfig{
			Start: []float64{94, 124, 154}, Velocity: []float64{2}, Min: 94, Max: 160,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := train.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	for index, pose := range train.Poses() {
		if pose.X != 0 || pose.Y != float64(60+[]int{96, 126, 156}[index]) || pose.ScaleX != 390 || pose.Frame != index {
			t.Fatalf("raster %d pose = %+v", index, pose)
		}
	}
	canvas := ebiten.NewImage(16, 16)
	defer canvas.Deallocate()
	before := append([]GroupPose(nil), train.Poses()...)
	train.Draw(canvas)
	train.Draw(canvas)
	for index, pose := range train.Poses() {
		if pose != before[index] {
			t.Fatalf("drawing advanced raster %d: %+v", index, pose)
		}
	}
	if allocs := testing.AllocsPerRun(100, func() { _ = train.Update(kit.Frame{}) }); allocs != 0 {
		t.Fatalf("train update allocates %v times", allocs)
	}
}

func TestTrainSamplesPhaseSpacedCosineAndRejectsMixedAxis(t *testing.T) {
	image := ebiten.NewImage(2, 2)
	defer image.Deallocate()
	wave := motion.Wave{Amplitude: 44, Spatial: .3, Speed: .06, Phase: .3, Cos: true}
	train, err := NewTrain(TrainConfig{Images: []*ebiten.Image{image, image, image}, ScaleX: 77, Y: TrainAxis{Offset: 106, Wave: &wave}})
	if err != nil {
		t.Fatal(err)
	}
	if err := train.Update(kit.Frame{Time: 120}); err != nil {
		t.Fatal(err)
	}
	for index, pose := range train.Poses() {
		want := 106 + 44*math.Cos(float64(index)*.3+120*.06+.3)
		if math.Abs(pose.Y-want) > 1e-12 || pose.Frame != index {
			t.Fatalf("raster %d pose = %+v, want y %v", index, pose, want)
		}
	}
	_, err = NewTrain(TrainConfig{Images: []*ebiten.Image{image}, Y: TrainAxis{
		Bounce: &motion.BounceBankConfig{Start: []float64{0}, Velocity: []float64{1}, Min: 0, Max: 10}, Wave: &wave,
	}})
	if err == nil {
		t.Fatal("axis accepted both bounce and wave")
	}
}
