package scene

import (
	"testing"

	kit "github.com/olivierh59500/democonstructionkit"
)

func TestProceduralProjectCompilesAtBothResolutions(t *testing.T) {
	for _, size := range [][2]int{{640, 360}, {320, 180}} {
		s, err := NewSize(size[0], size[1])
		if err != nil {
			t.Fatal(err)
		}
		for _, time := range []float64{0, 1, 3, 6, 9, 12, 15, 18, 120} {
			if err := s.Update(kit.Frame{Time: time}); err != nil {
				t.Fatal(err)
			}
		}
		if s.Width != size[0] || s.Height != size[1] || s.SurfaceBytes() <= int64(size[0]*size[1]*4) {
			t.Fatal("invalid scene dimensions or storage estimate")
		}
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
