//go:build dck_crt_rendercheck

package effects

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
)

type crtBlendCheck struct {
	source, overTarget, copyTarget *ebiten.Image
	over, copy                     *CRTOverlay
	checked                        bool
	err                            error
}

func (check *crtBlendCheck) Layout(int, int) (int, int) { return 32, 32 }
func (check *crtBlendCheck) Update() error              { return nil }
func (check *crtBlendCheck) Draw(*ebiten.Image) {
	if check.checked {
		return
	}
	check.checked = true
	check.overTarget.Fill(color.RGBA{255, 0, 0, 255})
	check.copyTarget.Fill(color.RGBA{255, 0, 0, 255})
	check.over.DrawAt(check.overTarget, check.source, 0, 0)
	check.copy.DrawAt(check.copyTarget, check.source, 0, 0)

	overPixels := make([]byte, 32*32*4)
	copyPixels := make([]byte, 32*32*4)
	check.overTarget.ReadPixels(overPixels)
	check.copyTarget.ReadPixels(copyPixels)
	pixel := func(data []byte, x, y int) color.RGBA {
		offset := (y*32 + x) * 4
		return color.RGBA{data[offset], data[offset+1], data[offset+2], data[offset+3]}
	}
	if got := pixel(overPixels, 2, 2); got != (color.RGBA{255, 0, 0, 255}) {
		check.err = fmt.Errorf("source-over transparent edge = %v, want red", got)
		return
	}
	if got := pixel(copyPixels, 2, 2); got != (color.RGBA{}) {
		check.err = fmt.Errorf("copy transparent edge = %v, want transparent", got)
		return
	}
	for _, data := range [][]byte{overPixels, copyPixels} {
		if got := pixel(data, 16, 16); got != (color.RGBA{0, 255, 0, 255}) {
			check.err = fmt.Errorf("opaque center = %v, want green", got)
			return
		}
	}
}

// TestMain runs the GPU check on macOS's required main thread.
func TestMain(m *testing.M) {
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	directory, err := os.MkdirTemp("", "dck-crt-blend-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var check *crtBlendCheck
	err = capture.Run(capture.Config{
		Directory: directory, Frames: []int{0}, Width: 32, Height: 32,
	}, func() (ebiten.Game, error) {
		source := ebiten.NewImage(32, 32)
		source.SubImage(image.Rect(8, 8, 24, 24)).(*ebiten.Image).Fill(color.RGBA{0, 255, 0, 255})
		over, err := NewCRTOverlay(CRTOverlayConfig{})
		if err != nil {
			return nil, err
		}
		copy, err := NewCRTOverlay(CRTOverlayConfig{Blend: ebiten.BlendCopy})
		if err != nil {
			return nil, err
		}
		check = &crtBlendCheck{
			source: source, overTarget: ebiten.NewImage(32, 32),
			copyTarget: ebiten.NewImage(32, 32), over: over, copy: copy,
		}
		return check, nil
	})
	if err == nil && check != nil {
		err = check.err
		if err == nil && !check.checked {
			err = fmt.Errorf("CRT blend check was not drawn")
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
