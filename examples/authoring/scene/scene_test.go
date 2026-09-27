package scene

import (
	"os"
	"testing"

	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/authoring"
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

func TestSavedFormationExamplesCompileAndAdvance(t *testing.T) {
	for _, name := range []string{"formation.json", "individual-paths.json"} {
		t.Run(name, func(t *testing.T) {
			file, err := os.Open("../" + name)
			if err != nil {
				t.Fatal(err)
			}
			project, decodeErr := authoring.Decode(file)
			closeErr := file.Close()
			if decodeErr != nil {
				t.Fatal(decodeErr)
			}
			if closeErr != nil {
				t.Fatal(closeErr)
			}
			scene, err := New(project)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := scene.Close(); err != nil {
					t.Error(err)
				}
			})
			for _, second := range []float64{0, .5, 1, 2, 3.5, 4} {
				if err := scene.Update(kit.Frame{Time: second}); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
