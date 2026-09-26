package sprites

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestAnimatedFieldDrawsSelectedAtlasFrameAndRespawns(t *testing.T) {
	red, blue := ebiten.NewImage(1, 1), ebiten.NewImage(1, 1)
	defer red.Deallocate()
	defer blue.Deallocate()
	red.Fill(color.RGBA{R: 255, A: 255})
	blue.Fill(color.RGBA{B: 255, A: 255})
	field, err := NewAnimatedField(AnimatedFieldConfig{
		Frames: []*ebiten.Image{red, blue}, OffsetX: 3, OffsetY: 4,
		Motion: motion.FrameFieldConfig{Count: 1, EndPhase: 2,
			Spawn: func(int, bool) motion.FrameParticle {
				return motion.FrameParticle{X: 1, Y: 2, Rate: 1}
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	dst := ebiten.NewImage(8, 8)
	defer dst.Deallocate()
	check := func(want color.RGBA) {
		t.Helper()
		dst.Clear()
		field.Draw(dst)
		got := color.RGBAModel.Convert(dst.At(4, 6)).(color.RGBA)
		if got != want {
			t.Fatalf("animated frame pixel %+v, want %+v", got, want)
		}
	}
	check(color.RGBA{R: 255, A: 255})
	if err := field.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	check(color.RGBA{B: 255, A: 255})
	if err := field.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	check(color.RGBA{R: 255, A: 255})
}
