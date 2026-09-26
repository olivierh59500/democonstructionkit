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

func TestTrainOwnedClockMatchesReplicantsThroughSpeedChanges(t *testing.T) {
	image := ebiten.NewImage(2, 2)
	defer image.Deallocate()
	wave := motion.Wave{Amplitude: -24, Spatial: -math.Pi / 2, Speed: 1, Phase: math.Pi / 2, Rectify: true}
	train, err := NewTrain(TrainConfig{
		Images: []*ebiten.Image{image, image}, OwnTime: true, TimeStep: .1,
		Y: TrainAxis{Offset: 326, Wave: &wave}, Spacing: motion.Point{X: 480},
	})
	if err != nil {
		t.Fatal(err)
	}
	phase := 0.0
	for tick := 0; tick < 5000; tick++ {
		speed := 1.4
		if tick >= 1500 && tick < 3000 {
			speed = .5
		} else if tick >= 3000 {
			speed = 2
		}
		if err := train.SetSpeedMultiplier(speed); err != nil {
			t.Fatal(err)
		}
		phase += .1 * speed
		if err := train.Update(kit.Frame{Time: 999}); err != nil {
			t.Fatal(err)
		}
		poses := train.Poses()
		if train.Time() != phase || len(poses) != 2 ||
			poses[0].Y != 326+wave.At(0, phase) || poses[1].Y != 326+wave.At(1, phase) ||
			math.Abs(poses[0].Y-(326-math.Abs(math.Cos(phase)*24))) > 1e-10 ||
			math.Abs(poses[1].Y-(326-math.Abs(math.Sin(phase)*24))) > 1e-10 ||
			poses[1].X-poses[0].X != 480 {
			t.Fatalf("tick %d: train time %g poses %+v, phase %g", tick, train.Time(), poses, phase)
		}
	}
	if err := train.SetTimeStep(.2); err != nil {
		t.Fatal(err)
	}
	if err := train.Reset(); err != nil || train.Time() != 0 {
		t.Fatalf("train reset time %g, error %v", train.Time(), err)
	}
	if err := train.Update(kit.Frame{}); err != nil || train.Time() != .4 {
		t.Fatalf("train changed step after reset: time %g, error %v", train.Time(), err)
	}
	if err := train.SetTimeStep(math.NaN()); err == nil {
		t.Fatal("accepted a nonfinite time step")
	}
	if allocations := testing.AllocsPerRun(100, func() {
		_ = train.Update(kit.Frame{})
		_ = train.Poses()
	}); allocations != 0 {
		t.Fatalf("owned train update allocated %v times", allocations)
	}
}

func TestTrainExternalClockRejectsOwnedSettings(t *testing.T) {
	image := ebiten.NewImage(1, 1)
	defer image.Deallocate()
	if _, err := NewTrain(TrainConfig{Images: []*ebiten.Image{image}, TimeStep: .1}); err == nil {
		t.Fatal("accepted an ignored owned-clock step")
	}
	train, err := NewTrain(TrainConfig{Images: []*ebiten.Image{image}})
	if err != nil {
		t.Fatal(err)
	}
	if err := train.SetSpeedMultiplier(2); err == nil {
		t.Fatal("accepted an owned-clock multiplier on an external train")
	}
	if err := train.Update(kit.Frame{Time: 12}); err != nil || train.Time() != 12 {
		t.Fatalf("external time = %g, error %v", train.Time(), err)
	}
}
