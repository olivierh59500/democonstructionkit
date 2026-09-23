package effects

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestProjectedBallTrainCustomSequenceAndStableDraw(t *testing.T) {
	image := ebiten.NewImage(1, 1)
	c := DefaultProjectedBallTrainConfig()
	c.Ball, c.Shadows, c.Count = image, []*ebiten.Image{image, image}, 2
	c.Programs = []BallMotionFunc{
		func(float64, int) BallMotion {
			return BallMotion{SpinSpeed: 3, Height: -20, AngleSpacing: 90, Radius: 50}
		},
		func(float64, int) BallMotion {
			return BallMotion{SpinSpeed: -1, Height: 20, AngleSpacing: 45, Radius: 100}
		},
	}
	c.Sequence = func(seconds float64, slot int) (int, int, float64) {
		return slot % 2, (slot + 1) % 2, math.Min(1, seconds/4)
	}
	p, err := NewProjectedBallTrain(c)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	if err := p.AdvanceAt(2); err != nil {
		t.Fatal(err)
	}
	phase, first, second := p.Phase(), p.Ball(0), p.Ball(1)
	if phase == 0 || first.X == second.X || first.Scale == 0 || second.Scale == 0 {
		t.Fatalf("custom sequence did not project two independent poses: phase=%g, balls=%+v %+v", phase, first, second)
	}
	canvas := ebiten.NewImage(768, 540)
	p.Draw(canvas)
	p.Draw(canvas)
	if p.Phase() != phase || p.Ball(0) != first || p.Ball(1) != second {
		t.Fatal("Draw advanced the ball sequence")
	}
	if err := p.PoseAt(4); err != nil || p.Phase() != phase {
		t.Fatalf("PoseAt advanced rotation: phase=%g, error=%v", p.Phase(), err)
	}
	c.Sequence = func(float64, int) (int, int, float64) { return 99, 0, 0 }
	if _, err := NewProjectedBallTrain(c); err == nil {
		t.Fatal("accepted an invalid movement index")
	}
}

func TestProjectedBallTrainLargeDepthOrder(t *testing.T) {
	p := &ProjectedBallTrain{balls: make([]ProjectedBall, 40), order: make([]int, 40)}
	for i := range p.balls {
		p.balls[i].Z = float64(i % 5)
	}
	p.sort()
	for i := 1; i < len(p.order); i++ {
		previous, current := p.order[i-1], p.order[i]
		if p.balls[previous].Z < p.balls[current].Z ||
			(p.balls[previous].Z == p.balls[current].Z && previous > current) {
			t.Fatalf("large depth order is not stable at %d: %d, %d", i, previous, current)
		}
	}
}
