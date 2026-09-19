package democonstructionkit

import (
	"errors"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type probe struct {
	frames        []Frame
	draws, closed int
	err           error
}

func (p *probe) Update(f Frame) error { p.frames = append(p.frames, f); return p.err }
func (p *probe) Draw(*ebiten.Image)   { p.draws++ }
func (p *probe) Close() error         { p.closed++; return p.err }
func TestSequencePassesLocalTimeAndStopsDrawingAtEnd(t *testing.T) {
	a, b := &probe{}, &probe{}
	s, err := NewSequence([]Effect{a, b}, []time.Duration{time.Second, 2 * time.Second}, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, seconds := range []float64{0, .5, 1, 2, 3} {
		if err = s.Update(Frame{Time: seconds, Delta: .1}); err != nil {
			t.Fatal(err)
		}
		s.Draw(nil)
	}
	if a.draws != 2 || b.draws != 2 || b.frames[0].Time != 0 || b.frames[1].Time != 1 {
		t.Fatal(a, b)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	if a.closed != 1 || b.closed != 1 {
		t.Fatal("child not closed")
	}
}
func TestCompositionPropagatesFailuresAndClosesEveryChild(t *testing.T) {
	want := errors.New("child error")
	a, b := &probe{err: want}, &probe{}
	g := Group{a, b}
	if !errors.Is(g.Update(Frame{}), want) {
		t.Fatal("lost update error")
	}
	if len(b.frames) != 0 {
		t.Fatal("updated after failure")
	}
	if !errors.Is(g.Close(), want) || b.closed != 1 {
		t.Fatal("cleanup stopped at first failure")
	}
}
