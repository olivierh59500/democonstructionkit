package sprites

import (
	"image"
	"math"
	"testing"

	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
)

func depthQueueConfig() DepthQueueConfig {
	return DepthQueueConfig{Points: []geometry.Vec2{{X: 1}, {X: 2}, {X: 3}, {X: 4}},
		Depth: 30, Near: 25, Far: 30, Spacing: 5, UpperPolicy: QueueUpperInclusiveFirst,
		Project: func(slot int, point geometry.Vec2, depth int) (FieldSample, bool) {
			return FieldSample{X: point.X, Y: float64(slot), Z: float64(depth), Scale: 1, Image: slot}, true
		}}
}

func TestDepthQueueBoundaryPoliciesMatchSequentialIntegerLoops(t *testing.T) {
	for _, upper := range []DepthQueueUpperPolicy{QueueUpperStrict, QueueUpperInclusiveFirst, QueueUpperInclusive} {
		for _, lowerInclusive := range []bool{false, true} {
			config := depthQueueConfig()
			config.UpperPolicy, config.LowerInclusive = upper, lowerInclusive
			queue, err := NewDepthQueue(config)
			if err != nil {
				t.Fatal(err)
			}
			depth, head := config.Depth, 0
			for tick := 0; tick < 3000; tick++ {
				change := tick%133 - 66
				depth += change
				for depth < config.Near || lowerInclusive && depth == config.Near {
					depth += config.Spacing
					head--
				}
				if upper == QueueUpperInclusiveFirst {
					if depth >= config.Far {
						depth -= config.Spacing
						head++
						for depth > config.Far {
							depth -= config.Spacing
							head++
						}
					}
				} else {
					for depth > config.Far || upper == QueueUpperInclusive && depth == config.Far {
						depth -= config.Spacing
						head++
					}
				}
				head = (head%len(config.Points) + len(config.Points)) % len(config.Points)
				if err := queue.Step(change); err != nil {
					t.Fatal(err)
				}
				if queue.Head() != head || queue.Depth() != depth {
					t.Fatalf("upper=%d lower=%v tick=%d: head/depth %d,%d want %d,%d", upper, lowerInclusive, tick, queue.Head(), queue.Depth(), head, depth)
				}
				for slot, pose := range queue.Poses() {
					pointIndex := (head + slot) % len(config.Points)
					if pose.PointIndex != pointIndex || pose.Index != pointIndex || pose.X != config.Points[pointIndex].X || pose.Z != float64(depth-slot*5) || !pose.Visible {
						t.Fatal("ring order, identity or per-slot depth changed", pose)
					}
				}
			}
			queue.Close()
		}
	}
}

func TestDepthQueueInclusiveFirstDistinguishesExactMultiplesAndLargeSteps(t *testing.T) {
	for _, example := range []struct{ delta, head, depth int }{
		{0, 1, 25}, {5, 1, 30}, {10, 2, 30}, {-5, 0, 25}, {-6, 3, 29},
		{1 << 30, 1, 29}, {-1 << 30, 0, 26},
	} {
		queue, err := NewDepthQueue(depthQueueConfig())
		if err != nil {
			t.Fatal(err)
		}
		if err := queue.Step(example.delta); err != nil {
			t.Fatal(err)
		}
		if queue.Head() != example.head || queue.Depth() != example.depth {
			t.Fatalf("delta=%d: %d,%d want %d,%d", example.delta, queue.Head(), queue.Depth(), example.head, example.depth)
		}
		queue.Close()
	}
}

func TestDepthQueueOwnsPointsAndFramesButKeepsCallerClocksSeparate(t *testing.T) {
	config := depthQueueConfig()
	config.Head, config.DepthStep = -1, -1
	config.Style.Frames = []image.Rectangle{image.Rect(2, 3, 5, 7)}
	calls := 0
	config.Project = func(slot int, point geometry.Vec2, depth int) (FieldSample, bool) {
		calls++
		return FieldSample{X: point.X, Y: float64(slot), Z: float64(depth), Scale: 1, Image: slot}, slot%2 == 0
	}
	queue, err := NewDepthQueue(config)
	if err != nil {
		t.Fatal(err)
	}
	defer queue.Close()
	config.Points[3].X = 100
	config.Style.Frames[0] = image.Rectangle{}
	if calls != 0 || len(queue.Samples()) != 0 || queue.Head() != 3 || queue.Style.Frames[0] != image.Rect(2, 3, 5, 7) {
		t.Fatal("constructor sampled or borrowed configuration storage")
	}
	queue.Update(kit.Frame{Time: 99})
	if queue.Depth() != 29 || queue.Poses()[0].X != 4 || len(queue.Samples()) != 2 || calls != 4 {
		t.Fatal("owned point bank, explicit velocity or visibility changed")
	}
	queue.SetDepthStep(2)
	queue.Update(kit.Frame{})
	if queue.Depth() != 26 || queue.Head() != 0 {
		t.Fatal("live velocity reset ring transport")
	}
	queue.Reset()
	if queue.Head() != 3 || queue.Depth() != 30 || len(queue.Samples()) != 0 || queue.Poses()[0].Visible {
		t.Fatal("reset did not restore unsampled initial state")
	}
	if allocations := testing.AllocsPerRun(100, func() { queue.Step(-2) }); allocations != 0 {
		t.Fatal("queue sampling allocated", allocations)
	}
}

func TestDepthQueueInvalidPoseAndDeltaKeepThePreviousPreparedFrame(t *testing.T) {
	bad := false
	config := depthQueueConfig()
	config.Project = func(slot int, point geometry.Vec2, depth int) (FieldSample, bool) {
		x := point.X
		if bad && slot == 2 {
			x = math.NaN()
		}
		return FieldSample{X: x, Z: float64(depth), Scale: 1}, true
	}
	queue, err := NewDepthQueue(config)
	if err != nil {
		t.Fatal(err)
	}
	defer queue.Close()
	queue.Step(-2)
	previous := append([]FieldSample(nil), queue.Samples()...)
	head, depth := queue.Head(), queue.Depth()
	bad = true
	if queue.Step(100) == nil || queue.Step(1<<30+1) == nil || queue.SetDepthStep(-1<<30-1) == nil {
		t.Fatal("invalid queue update was accepted")
	}
	if queue.Head() != head || queue.Depth() != depth || len(queue.Samples()) != len(previous) {
		t.Fatal("invalid update changed transport or prepared count")
	}
	for i, sample := range queue.Samples() {
		if sample != previous[i] {
			t.Fatal("invalid pose changed the preceding prepared frame")
		}
	}
	bad = false
	if err := queue.Step(0); err != nil {
		t.Fatal("queue could not recover after rejection", err)
	}
}

func TestDepthQueueDefaultCameraAndConfigurationBounds(t *testing.T) {
	config := depthQueueConfig()
	config.Project = nil
	config.Camera = geometry.Camera{Focal: 50, Near: 19, Center: geometry.Vec2{X: 10, Y: 20}}
	queue, err := NewDepthQueue(config)
	if err != nil {
		t.Fatal(err)
	}
	queue.Step(-1)
	if len(queue.Samples()) != 3 {
		t.Fatal("default camera did not clip the near slot")
	}
	position, scale, visible := config.Camera.Project(geometry.Vec3{X: 1, Z: 29})
	if !visible || queue.Poses()[0].X != position.X || queue.Poses()[0].Y != position.Y || queue.Poses()[0].Scale != scale {
		t.Fatal("default camera changed ordinary perspective")
	}
	queue.Close()
	queue.Close()
	if queue.Step(0) == nil || queue.Reset() == nil || queue.SetDepthStep(0) == nil || len(queue.Poses()) != 0 {
		t.Fatal("closed queue accepted updates or retained poses")
	}
	for i := 0; i < 12; i++ {
		invalid := depthQueueConfig()
		switch i {
		case 0:
			invalid.Points = nil
		case 1:
			invalid.Points = make([]geometry.Vec2, 65537)
		case 2:
			invalid.Near = invalid.Far
		case 3:
			invalid.Spacing = 0
		case 4:
			invalid.Spacing = 6
		case 5:
			invalid.Points[0].X = math.Inf(1)
		case 6:
			invalid.Depth = 1<<30 + 1
		case 7:
			invalid.DepthStep = -1<<30 - 1
		case 8:
			invalid.UpperPolicy = 3
		case 9:
			invalid.RendererCapacity = 65537
		case 10:
			invalid.Project = nil
		case 11:
			invalid.Near, invalid.Far, invalid.Spacing = -(1 << 30), -(1<<30)+5, 5
		}
		if value, err := NewDepthQueue(invalid); err == nil {
			value.Close()
			t.Fatal("invalid queue configuration accepted", i)
		}
	}
	var absent *DepthQueue
	if absent.Step(0) == nil || absent.Update(kit.Frame{}) == nil || absent.Reset() == nil {
		t.Fatal("nil queue accepted updates")
	}
	absent.Draw(nil)
	absent.Close()
}
