//go:build dck_composition_rendercheck

package effects

import (
	"fmt"
	"image/color"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
	"github.com/olivierh59500/democonstructionkit/motion"
)

type compositionCheck struct {
	pair       *GatedBackgroundPair
	mask       *Mask
	pairTarget *ebiten.Image
	maskTarget *ebiten.Image
	checked    bool
	err        error
}

func (check *compositionCheck) Layout(int, int) (int, int) { return 16, 16 }
func (check *compositionCheck) Update() error              { return nil }
func (check *compositionCheck) Draw(*ebiten.Image) {
	if check.checked {
		return
	}
	check.checked = true
	check.pair.Draw(check.pairTarget)
	check.mask.Draw(check.maskTarget)
	pairPixels := make([]byte, 2*2*4)
	maskPixels := make([]byte, 6*8*4)
	check.pairTarget.ReadPixels(pairPixels)
	check.maskTarget.ReadPixels(maskPixels)
	if got := rgbaAt(pairPixels, 2, 0, 0); got != (color.RGBA{B: 255, A: 255}) {
		check.err = fmt.Errorf("paired background order produced %v, want blue", got)
		return
	}
	if got := rgbaAt(maskPixels, 6, 2, 4); got != (color.RGBA{R: 255, A: 255}) {
		check.err = fmt.Errorf("masked output pixel %v, want red", got)
		return
	}
	if got := rgbaAt(maskPixels, 6, 2, 3); got.A != 0 {
		check.err = fmt.Errorf("top clear did not remove row: %v", got)
	}
}

func rgbaAt(pixels []byte, width, x, y int) color.RGBA {
	offset := (y*width + x) * 4
	return color.RGBA{pixels[offset], pixels[offset+1], pixels[offset+2], pixels[offset+3]}
}

func newCompositionCheck() (*compositionCheck, error) {
	red, blue := ebiten.NewImage(2, 2), ebiten.NewImage(2, 2)
	red.Fill(color.RGBA{R: 255, A: 255})
	blue.Fill(color.RGBA{B: 255, A: 255})
	background := composite.BackgroundConfig{CopiesX: 1, CopiesY: 1}
	pair, err := NewGatedBackgroundPair(GatedBackgroundPairConfig{
		Images:      [2]*ebiten.Image{red, blue},
		Backgrounds: [2]composite.BackgroundConfig{background, background},
		Motion: motion.GatedBackgroundPairConfig{GateOpen: 1, GateReset: 2,
			FirstX:     motion.ThresholdAxis{Lower: -1, Upper: 1},
			FirstY:     motion.ThresholdAxis{Lower: -1, Upper: 1},
			SecondY:    motion.ThresholdAxis{Lower: -1, Upper: 1},
			SecondXMin: -1, SecondXMax: 1},
	})
	if err != nil {
		return nil, err
	}
	content := ebiten.NewImage(2, 4)
	alpha := ebiten.NewImage(1, 1)
	content.Fill(color.RGBA{R: 255, A: 255})
	alpha.Fill(color.White)
	mask, err := NewMaskWith(MaskConfig{
		Content: kit.Func{OnDraw: func(dst *ebiten.Image) { dst.DrawImage(content, nil) }},
		Alpha:   kit.Func{OnDraw: func(dst *ebiten.Image) { dst.DrawImage(alpha, nil) }},
		Width:   2, Height: 4, Blend: ebiten.BlendDestinationIn,
		MaskY: 1, OutputX: 2, OutputY: 3, ClearTop: 1,
	})
	if err != nil {
		return nil, err
	}
	return &compositionCheck{
		pair: pair, mask: mask,
		pairTarget: ebiten.NewImage(2, 2), maskTarget: ebiten.NewImage(6, 8),
	}, nil
}

// TestMain runs readback checks while Ebitengine's graphics context is active.
func TestMain(m *testing.M) {
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	directory, err := os.MkdirTemp("", "dck-composition-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var check *compositionCheck
	err = capture.Run(capture.Config{
		Directory: directory, Frames: []int{0}, Width: 16, Height: 16,
	}, func() (ebiten.Game, error) {
		check, err = newCompositionCheck()
		return check, err
	})
	if err == nil && check != nil {
		err = check.err
		if err == nil && !check.checked {
			err = fmt.Errorf("composition GPU check was not drawn")
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
