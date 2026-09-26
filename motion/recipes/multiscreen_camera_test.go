package recipes

import (
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

type legacyMultiscreenCamera struct {
	state                  int
	stateTimer, transition float64
	x, y                   float64
}

func legacyCubicInOut(value float64) float64 {
	if value < .5 {
		return 4 * value * value * value
	}
	u := -2*value + 2
	return 1 - u*u*u/2
}

func (c *legacyMultiscreenCamera) step() {
	dt := 1.0 / 60.0
	c.stateTimer += dt
	switch c.state {
	case MultiscreenView1:
		if c.stateTimer >= 7 {
			c.state, c.transition, c.stateTimer = MultiscreenMove12, 0, 0
		}
	case MultiscreenMove12:
		c.transition += dt
		progress := c.transition / 4
		if progress >= 1 {
			c.state, c.x, c.y, c.stateTimer = MultiscreenView2, 800, 0, 0
		} else {
			c.x, c.y = legacyCubicInOut(progress)*800, 0
		}
	case MultiscreenView2:
		if c.stateTimer >= 7 {
			c.state, c.transition, c.stateTimer = MultiscreenMove23, 0, 0
		}
	case MultiscreenMove23:
		c.transition += dt
		progress := c.transition / 4
		if progress >= 1 {
			c.state, c.x, c.y, c.stateTimer = MultiscreenView3, 800, 600, 0
		} else {
			c.x, c.y = 800, legacyCubicInOut(progress)*600
		}
	case MultiscreenView3:
		if c.stateTimer >= 7 {
			c.state, c.transition, c.stateTimer = MultiscreenMove34, 0, 0
		}
	case MultiscreenMove34:
		c.transition += dt
		progress := c.transition / 4
		if progress >= 1 {
			c.state, c.x, c.y, c.stateTimer = MultiscreenView4, 0, 600, 0
		} else {
			c.x, c.y = 800-legacyCubicInOut(progress)*800, 600
		}
	case MultiscreenView4:
		if c.stateTimer >= 7 {
			c.state, c.transition, c.stateTimer = MultiscreenZoomOut, 0, 0
		}
	case MultiscreenZoomOut:
		c.transition += dt
		if c.transition/4 >= 1 {
			c.state, c.x, c.y, c.stateTimer = MultiscreenOverview, 400, 300, 0
		}
	case MultiscreenOverview:
		if c.stateTimer >= 7 {
			c.state, c.transition, c.stateTimer = MultiscreenLoop, 0, 0
		}
	case MultiscreenLoop:
		c.transition += dt
		progress := c.transition / 4
		if progress >= 1 {
			c.state, c.x, c.y, c.stateTimer = MultiscreenView1, 0, 0, 0
		} else {
			eased := legacyCubicInOut(progress)
			c.x, c.y = 400-eased*400, 300-eased*300
		}
	}
}

func (c legacyMultiscreenCamera) pose() (x, y, zoom float64, mask uint64, direct int) {
	x, y, zoom, direct = c.x+400, c.y+300, 1, -1
	switch c.state {
	case MultiscreenView1:
		return x, y, zoom, 1, 0
	case MultiscreenMove12:
		mask = 1 | 2
	case MultiscreenView2:
		return x, y, zoom, 2, 1
	case MultiscreenMove23:
		mask = 2 | 4
	case MultiscreenView3:
		return x, y, zoom, 4, 2
	case MultiscreenMove34:
		mask = 4 | 8
	case MultiscreenView4:
		return x, y, zoom, 8, 3
	default:
		mask = 15
	}
	switch c.state {
	case MultiscreenZoomOut:
		progress := c.transition / 4
		if progress > 1 {
			progress = 1
		}
		eased := legacyCubicInOut(progress)
		x, y = 400+(800-400)*eased, 900+(600-900)*eased
		zoom = 1 - eased*.5
	case MultiscreenOverview:
		x, y, zoom = 800, 600, .5
	case MultiscreenLoop:
		progress := c.transition / 4
		if progress > 1 {
			progress = 1
		}
		eased := legacyCubicInOut(progress)
		x, y = 800+(400-800)*eased, 600+(300-600)*eased
		zoom = .5 + (1-.5)*eased
	}
	return
}

func TestMultiscreenCameraTourMatchesOriginalAcrossLoops(t *testing.T) {
	tour, err := motion.NewCameraTour(MultiscreenCameraTour())
	if err != nil {
		t.Fatal(err)
	}
	legacy := legacyMultiscreenCamera{}
	completedLoops := 0
	previous := legacy.state
	for tick := 0; tick < 12000; tick++ {
		legacy.step()
		tour.Step()
		state := tour.State()
		x, y, zoom, mask, direct := legacy.pose()
		elapsed := legacy.stateTimer
		if legacy.state == MultiscreenMove12 || legacy.state == MultiscreenMove23 ||
			legacy.state == MultiscreenMove34 || legacy.state == MultiscreenZoomOut || legacy.state == MultiscreenLoop {
			elapsed = legacy.transition
		}
		if state.Segment != legacy.state || state.Elapsed != elapsed ||
			state.CenterX != x || state.CenterY != y || state.Zoom != zoom ||
			state.VisibleMask != mask || state.Direct != direct {
			t.Fatalf("tick %d tour=%+v, source state=%d elapsed=%v pose=(%v,%v,%v,%b,%d)",
				tick, state, legacy.state, elapsed, x, y, zoom, mask, direct)
		}
		if previous == MultiscreenLoop && legacy.state == MultiscreenView1 {
			completedLoops++
		}
		previous = legacy.state
	}
	if completedLoops < 2 {
		t.Fatalf("only %d full camera loops were checked", completedLoops)
	}
	if got := testing.AllocsPerRun(100, func() { tour.Step() }); got != 0 {
		t.Fatalf("camera tour step allocated %.2f objects", got)
	}
}

func BenchmarkMultiscreenCameraTour(b *testing.B) {
	tour, err := motion.NewCameraTour(MultiscreenCameraTour())
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		tour.Step()
	}
}
