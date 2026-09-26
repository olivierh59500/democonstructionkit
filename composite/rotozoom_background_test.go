package composite

import (
	"image"
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

func TestSourceQuadKeepsCocoVertexGeometryAndTint(t *testing.T) {
	pose := Repetition{CenterX: 400, CenterY: 300, Zoom: 2,
		PhaseX: 1600, PhaseY: 1200, Color: [4]float32{.5, .5, .5, 1}}
	var vertices [4]ebiten.Vertex
	fillSourceQuadVertices(&vertices, pose, image.Pt(3200, 2400))
	if vertices[0].DstX != -2800 || vertices[0].DstY != -2100 ||
		vertices[0].SrcX != 0 || vertices[0].SrcY != 0 ||
		vertices[3].DstX != 3600 || vertices[3].DstY != 2700 ||
		vertices[3].SrcX != 3200 || vertices[3].SrcY != 2400 ||
		vertices[2].ColorR != .5 || vertices[2].ColorA != 1 {
		t.Fatalf("source quad differs from the authored corners: %+v", vertices)
	}
}

func TestRotozoomBackgroundUsesReadyProgramAtFrameZero(t *testing.T) {
	image := ebiten.NewImage(2, 2)
	defer image.Deallocate()
	program := &testRotozoomProgram{pose: Repetition{CenterX: 3, Zoom: 1.5}}
	background, err := NewRotozoomBackground(RotozoomBackgroundConfig{Image: image, Program: program})
	if err != nil {
		t.Fatal(err)
	}
	if got := background.Repetition(); got.CenterX != 3 || got.Zoom != 1.5 {
		t.Fatalf("ready initial program pose %+v", got)
	}
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
