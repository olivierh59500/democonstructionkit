package composite

import (
	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

// Sprites renders a configurable number of independently sampled sprites/logos.
// Sample controls the frame, crop, pivot, transform, blend and color for each one.
// Prepare can initialize exact sine recurrences before ordered sampling begins.
// Reverse chooses reverse drawing order without changing the sample indices.
type Sprites struct {
	Count   int
	CountAt func(kit.Frame) int
	Prepare func(kit.Frame)
	Sample  func(index int, frame kit.Frame) Instance
	Reverse bool
	frame   kit.Frame
}

func (s *Sprites) Update(f kit.Frame) error { s.frame = f; return nil }
func (s *Sprites) Draw(dst *ebiten.Image) {
	if s.Sample == nil {
		return
	}
	if s.Prepare != nil {
		s.Prepare(s.frame)
	}
	count := s.Count
	if s.CountAt != nil {
		count = s.CountAt(s.frame)
	}
	if s.Reverse {
		for i := count - 1; i >= 0; i-- {
			s.Sample(i, s.frame).Draw(dst)
		}
	} else {
		for i := 0; i < count; i++ {
			s.Sample(i, s.frame).Draw(dst)
		}
	}
}
