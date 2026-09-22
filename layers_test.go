package democonstructionkit

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/timeline"
)

func TestTimedLayersSkipInactiveEffectsAndCanRetrigger(t *testing.T) {
	var updates []float64
	draws := 0
	effect := Func{OnUpdate: func(f Frame) error { updates = append(updates, f.Time); return nil }, OnDraw: func(*ebiten.Image) { draws++ }}
	layers, err := NewLayers(8, 8, TimedLayer{Effect: effect, Window: timeline.Window{Start: 2, Duration: 1}, LocalTime: true})
	if err != nil {
		t.Fatal(err)
	}
	defer layers.Close()
	dst := ebiten.NewImage(8, 8)
	defer dst.Deallocate()
	for _, time := range []float64{0, 2.25, 3} {
		layers.Update(Frame{Time: time})
		layers.Draw(dst)
	}
	if len(updates) != 1 || updates[0] != .25 || draws != 1 {
		t.Fatal(updates, draws)
	}
	if err := layers.SetWindow(0, timeline.Window{Start: 5, Duration: 1}); err != nil {
		t.Fatal(err)
	}
	layers.Update(Frame{Time: 5.5})
	layers.Draw(dst)
	if len(updates) != 2 || updates[1] != .5 || draws != 2 {
		t.Fatal(updates, draws)
	}
}

func TestPipelineRendersSourceOnceAndSkipsDisabledPasses(t *testing.T) {
	draws, updates, calls := 0, 0, 0
	base := Func{OnUpdate: func(Frame) error { updates++; return nil }, OnDraw: func(*ebiten.Image) { draws++ }}
	p, err := NewPipeline(base, 8, 8, ImagePass{Enabled: func(f Frame) bool { return f.Time >= 2 }, Apply: func(dst, src *ebiten.Image, _ Frame) {
		if dst == src {
			t.Fatal("feedback alias")
		}
		calls++
		dst.DrawImage(src, nil)
	}})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	dst := ebiten.NewImage(8, 8)
	defer dst.Deallocate()
	p.Update(Frame{Time: 0})
	p.Draw(dst)
	if draws != 1 || calls != 0 {
		t.Fatal(draws, calls)
	}
	p.Update(Frame{Time: 2})
	p.Draw(dst)
	if draws != 2 || updates != 2 || calls != 1 {
		t.Fatal(draws, updates, calls)
	}
	p.Draw(dst)
	if updates != 2 || draws != 3 || calls != 2 {
		t.Fatal("drawing advanced time")
	}
}
