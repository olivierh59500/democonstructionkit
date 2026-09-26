package recipes

import (
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestUnionTNTCarouselKeepsPoseAcrossRecessionalHandoffs(t *testing.T) {
	config := UnionTNTCarouselMotion()
	carousel, err := motion.NewModelCarousel(config)
	if err != nil {
		t.Fatal(err)
	}
	active, pending := 1, 1
	angles := config.Models[active].Rotation
	camera := motion.Vector3{Z: 10}
	distance := 0.0
	changing := false
	switches := 0
	selections := map[int]int{120: 3, 150: 4, 400: 0, 900: 2, 1450: 1, 2100: 4, 3000: 0}
	for tick := 0; tick < 3600; tick++ {
		if selected, ok := selections[tick]; ok {
			pending, changing = selected, true
			if err := carousel.Select(selected); err != nil {
				t.Fatal(err)
			}
		}
		rotating := distance >= 700
		if !rotating {
			distance += 10
		}
		if changing {
			if camera.Z < 10000 {
				camera.Z += 100
			} else {
				active = pending
				distance = 0
				camera = config.Models[active].Camera
				angles = config.Models[active].Rotation
				changing = false
				switches++
			}
		}
		want := motion.ModelCarouselPose{Active: active, Pending: pending,
			Camera: camera, Distance: distance, Rotation: angles,
			Changing: changing, Rotating: rotating}
		if rotating {
			angles = angles.Add(config.Models[active].RotationStep)
		}
		carousel.Step()
		if got := carousel.Pose(); got != want {
			t.Fatalf("tick %d pose=%+v, want %+v", tick, got, want)
		}
	}
	if switches < 3 {
		t.Fatalf("only %d handoffs reached the new object", switches)
	}
	if err := carousel.Select(carousel.Count()); err == nil {
		t.Fatal("accepted an out-of-range model selection")
	}
	if got := testing.AllocsPerRun(100, carousel.Step); got != 0 {
		t.Fatalf("model carousel step allocated %.2f objects", got)
	}
}

func BenchmarkUnionTNTCarouselMotion(b *testing.B) {
	carousel, err := motion.NewModelCarousel(UnionTNTCarouselMotion())
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		carousel.Step()
	}
}
