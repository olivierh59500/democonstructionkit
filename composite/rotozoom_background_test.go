package composite

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

type testRotozoomProgram struct{ pose Repetition }

func (p *testRotozoomProgram) Update(f kit.Frame) error {
	p.pose = Repetition{CenterX: f.Time * 2, Zoom: 2}
	return nil
}
func (p *testRotozoomProgram) Repetition() Repetition { return p.pose }

func TestRotozoomBackgroundMotionAndProgram(t *testing.T) {
	image := ebiten.NewImage(4, 4)
	t.Cleanup(image.Deallocate)
	background, err := NewRotozoomBackground(RotozoomBackgroundConfig{Image: image,
		Pose:     Repetition{CenterX: 10, CenterY: 20, Zoom: 1, Rotation: .1, PhaseX: 5, PhaseY: 4},
		Velocity: RotozoomVelocity{CenterX: 3, CenterY: -2, Zoom: .25, Rotation: .2, PhaseX: 7, PhaseY: -1}})
	if err != nil {
		t.Fatal(err)
	}
	if err := background.Update(kit.Frame{Time: 2}); err != nil {
		t.Fatal(err)
	}
	p := background.Repetition()
	if p.CenterX != 16 || p.CenterY != 16 || p.Zoom != 1.5 || p.Rotation != .5 || p.PhaseX != 19 || p.PhaseY != 2 {
		t.Fatalf("velocity pose = %+v", p)
	}
	program := &testRotozoomProgram{}
	background, err = NewRotozoomBackground(RotozoomBackgroundConfig{Image: image, Program: program, Velocity: RotozoomVelocity{CenterX: 3}})
	if err != nil {
		t.Fatal(err)
	}
	if err := background.Update(kit.Frame{Time: 2}); err != nil {
		t.Fatal(err)
	}
	if p := background.Repetition(); p.CenterX != 10 || p.Zoom != 2 {
		t.Fatalf("program pose = %+v", p)
	}
}

func TestRotozoomBackgroundRejectsInvalidPose(t *testing.T) {
	if _, err := NewRotozoomBackground(RotozoomBackgroundConfig{}); err == nil {
		t.Fatal("accepted missing texture")
	}
	image := ebiten.NewImage(4, 4)
	t.Cleanup(image.Deallocate)
	if _, err := NewRotozoomBackground(RotozoomBackgroundConfig{Image: image, Pose: Repetition{Zoom: -1}}); err == nil {
		t.Fatal("accepted negative zoom")
	}
	if _, err := NewRotozoomBackground(RotozoomBackgroundConfig{Image: image, Velocity: RotozoomVelocity{Rotation: math.NaN()}}); err == nil {
		t.Fatal("accepted nonfinite velocity")
	}
	background, err := NewRotozoomBackground(RotozoomBackgroundConfig{Image: image, Pose: Repetition{Zoom: 1}, Velocity: RotozoomVelocity{Zoom: -1}})
	if err != nil {
		t.Fatal(err)
	}
	if err := background.Update(kit.Frame{Time: 2}); err == nil {
		t.Fatal("accepted animation crossing zero zoom")
	}
}
