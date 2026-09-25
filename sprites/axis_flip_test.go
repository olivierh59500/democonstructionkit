package sprites

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestAxisFlipPreservesFrontBackThresholdAndBounce(t *testing.T) {
	front, back := ebiten.NewImage(2, 3), ebiten.NewImage(4, 5)
	defer front.Deallocate()
	defer back.Deallocate()
	flip, err := NewAxisFlip(AxisFlipConfig{
		Front: front, Back: back, SwitchAt: .01, BackAngle: 180,
		Motion: motion.BounceBankConfig{Start: []float64{1}, Velocity: []float64{-.02}, Min: -1, Max: 1, Inclusive: true, Directional: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	scale, step := 1.0, -.02
	for tick := 0; tick < 1200; tick++ {
		pose := flip.Pose()
		wantImage, wantAngle := front, 0.0
		if scale <= .01 {
			wantImage, wantAngle = back, 180
		}
		if pose.Image != wantImage || pose.ScaleX != 1 || pose.ScaleY != scale || pose.Angle != wantAngle {
			t.Fatalf("tick %d: pose %+v, want image %p, scale %v, angle %v", tick, pose, wantImage, scale, wantAngle)
		}
		scale += step
		if scale <= -1 {
			step = .02
		}
		if scale >= 1 {
			step = -.02
		}
		flip.Step()
	}
	flip.Reset()
	if flip.Pose().Image != front || flip.Pose().ScaleY != 1 {
		t.Fatal("reset did not restore the front pose")
	}
}

func TestAxisFlipRejectsInvalidMotionAndCanFlipOneImage(t *testing.T) {
	front := ebiten.NewImage(2, 2)
	defer front.Deallocate()
	motionConfig := motion.BounceBankConfig{Start: []float64{1}, Velocity: []float64{-.1}, Min: -1, Max: 1}
	if _, err := NewAxisFlip(AxisFlipConfig{Motion: motionConfig}); err == nil {
		t.Fatal("missing front image accepted")
	}
	if _, err := NewAxisFlip(AxisFlipConfig{Front: front, Motion: motionConfig, SwitchAt: math.NaN()}); err == nil {
		t.Fatal("nonfinite threshold accepted")
	}
	flip, err := NewAxisFlip(AxisFlipConfig{Front: front, Motion: motionConfig})
	if err != nil {
		t.Fatal(err)
	}
	if flip.Pose().Image != front {
		t.Fatal("single-image flip lost its source")
	}
}

func TestAxisFlipSawSwitchesFacesOnlyAtStrictWrap(t *testing.T) {
	front, back := ebiten.NewImage(80, 16), ebiten.NewImage(80, 16)
	defer front.Deallocate()
	defer back.Deallocate()
	flip, err := NewAxisFlip(AxisFlipConfig{
		Front: front, Back: back,
		Saw: &motion.SawToggleConfig{Start: 0, Velocity: .08, Boundary: 1, Restart: -1},
	})
	if err != nil {
		t.Fatal(err)
	}
	scale, alternate := 0.0, false
	for tick := 0; tick < 300; tick++ {
		flip.Step()
		scale += .08
		if scale > 1 {
			scale = -1
			alternate = !alternate
		}
		pose := flip.Pose()
		want := front
		if alternate {
			want = back
		}
		if pose.Image != want || pose.ScaleY != scale {
			t.Fatalf("tick %d: pose %+v, want %p at scale %v", tick, pose, want, scale)
		}
	}
	flip.Reset()
	if flip.Pose().Image != front || flip.Pose().ScaleY != 0 {
		t.Fatal("saw reset retained face or phase")
	}
}
