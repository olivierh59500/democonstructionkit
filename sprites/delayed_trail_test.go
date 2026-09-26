package sprites

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
)

func TestDelayedTrailKeepsIndependentImageDelaysAndReverseOrder(t *testing.T) {
	images := make([]*ebiten.Image, 4)
	for i := range images {
		images[i] = ebiten.NewImage(2, 2)
		defer images[i].Deallocate()
	}
	trail, err := NewDelayedTrail(DelayedTrailConfig{
		Images: images, Delays: []int{0, 20, 40, 60}, Capacity: 61,
		Initial: geometry.Vec2{X: 384, Y: 268}, Reverse: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !trail.group.config.Reverse || len(trail.Poses()) != 4 {
		t.Fatal("trail changed pointer image order")
	}
	for tick := 0; tick < 90; tick++ {
		if err := trail.SetPosition(geometry.Vec2{X: float64(tick), Y: 100}); err != nil {
			t.Fatal(err)
		}
		if err := trail.Update(kit.Frame{}); err != nil {
			t.Fatal(err)
		}
		for image, delay := range []int{0, 20, 40, 60} {
			wantX, wantY := 384.0, 268.0
			if tick >= delay {
				wantX, wantY = float64(tick-delay), 100
			}
			if pose := trail.Poses()[image]; pose.X != wantX || pose.Y != wantY || pose.Frame != image {
				t.Fatalf("tick %d image %d pose %+v, want (%v,%v)", tick, image, pose, wantX, wantY)
			}
		}
	}
	if allocations := testing.AllocsPerRun(100, func() {
		if err := trail.Update(kit.Frame{}); err != nil {
			t.Fatal(err)
		}
	}); allocations != 0 {
		t.Fatalf("trail update allocated %v", allocations)
	}
	if err := trail.Reset(); err != nil {
		t.Fatal(err)
	}
	for _, pose := range trail.Poses() {
		if pose.X != 384 || pose.Y != 268 {
			t.Fatal("trail reset retained old history")
		}
	}
}

func TestDelayedTrailRejectsHistoryShorterThanOldestSprite(t *testing.T) {
	image := ebiten.NewImage(1, 1)
	defer image.Deallocate()
	if _, err := NewDelayedTrail(DelayedTrailConfig{
		Images: []*ebiten.Image{image}, Delays: []int{60}, Capacity: 60,
	}); err == nil {
		t.Fatal("accepted a history that cannot sample its last delay")
	}
}
