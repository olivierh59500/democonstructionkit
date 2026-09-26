package sprites

import (
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestCoupledLogoPairUsesQuantizedFramesAndDepthOrder(t *testing.T) {
	red, blue := ebiten.NewImage(2, 2), ebiten.NewImage(2, 2)
	defer red.Deallocate()
	defer blue.Deallocate()
	red.Fill(color.RGBA{R: 255, A: 255})
	blue.Fill(color.RGBA{B: 255, A: 255})
	for _, tc := range []struct {
		phase float64
		want  color.RGBA
		order [2]int
	}{{0, color.RGBA{B: 255, A: 255}, [2]int{0, 1}},
		{math.Pi, color.RGBA{R: 255, A: 255}, [2]int{1, 0}}} {
		pair, err := NewCoupledLogoPair(CoupledLogoPairConfig{
			Motion: motion.CoupledLogoConfig{DepthBase: 1, DepthDivisor: 2, YBase: 2,
				SecondaryDepthCos: .125, PhaseSecondary: tc.phase},
			Primary:   QuantizedLogoLayer{Image: red, FrameCount: 2, Gain: 2},
			Secondary: QuantizedLogoLayer{Image: blue, FrameCount: 2, Gain: 2},
			CenterX:   2,
		})
		if err != nil {
			t.Fatal(err)
		}
		if got := pair.DrawOrder(); got != tc.order {
			t.Fatalf("depth order %v, want %v", got, tc.order)
		}
		if got := pair.FrameIndex(0); got != 1 {
			t.Fatalf("primary frame %d, want 1", got)
		}
		dst := ebiten.NewImage(4, 4)
		pair.Draw(dst)
		got := color.RGBAModel.Convert(dst.At(1, 1)).(color.RGBA)
		if got != tc.want {
			t.Fatalf("front logo pixel %+v, want %+v", got, tc.want)
		}
		dst.Deallocate()
		pair.Close()
	}
}
