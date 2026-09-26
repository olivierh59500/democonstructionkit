package motion

import "testing"

func TestGravityBounceMatchesPhotonOvershootUntilFade(t *testing.T) {
	bounce, err := NewGravityBounce(GravityBounceConfig{
		StartPosition: 184, StartRebound: -9.50,
		Gravity: .30, Floor: 445, Damping: .70, StopRebound: -.70,
	})
	if err != nil {
		t.Fatal(err)
	}
	y, velocity, rebound := 184.0, 0.0, -9.50
	finishedAt := 0
	for tick := 1; tick <= 5000; tick++ {
		velocity += .30
		y += velocity
		if y > 445 {
			velocity = rebound
			rebound *= .70
		}
		finished := rebound >= -.70
		if gotFinished := bounce.Step(); gotFinished != finished {
			t.Fatalf("tick %d completion = %v, want %v", tick, gotFinished, finished)
		}
		if got := bounce.State(); got.Position != y || got.Velocity != velocity || got.Rebound != rebound || got.Finished != finished {
			t.Fatalf("tick %d pose = %+v, want y=%v velocity=%v rebound=%v finished=%v", tick, got, y, velocity, rebound, finished)
		}
		if finished {
			finishedAt = tick
			break
		}
	}
	if finishedAt == 0 {
		t.Fatal("photon never reached the fade cue")
	}
	final := bounce.State()
	for range 10 {
		if !bounce.Step() || bounce.State() != final {
			t.Fatal("finished bounce moved after handoff")
		}
	}
	if got := testing.AllocsPerRun(100, func() {
		bounce.Reset()
		for range 100 {
			bounce.Step()
		}
	}); got != 0 {
		t.Fatalf("gravity bounce step allocated %.2f objects", got)
	}
}
