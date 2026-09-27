//go:build dck_crt_origincheck

package effects

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
)

type crtOriginCheck struct {
	sourceRoot, source, manualBacking, manualView *ebiten.Image
	want, got                                     *ebiten.Image
	reference, padded                             *CRTOverlay
	frames                                        int
	err                                           error
}

func (c *crtOriginCheck) Layout(int, int) (int, int) { return 32, 32 }
func (c *crtOriginCheck) Update() error              { c.frames++; return nil }
func (c *crtOriginCheck) Draw(*ebiten.Image) {
	if c.err != nil {
		return
	}
	c.sourceRoot.Clear()
	paint := color.RGBA{R: 255, A: 255}
	if c.frames > 0 {
		paint = color.RGBA{G: 200, B: 80, A: 255}
	}
	c.source.Fill(paint)
	c.manualView.Clear()
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(1, 1) // Source bounds start at (3,2); the target starts at (4,3).
	c.manualView.DrawImage(c.source, &op)
	c.want.Fill(color.Black)
	c.got.Fill(color.Black)
	c.reference.DrawAt(c.want, c.manualView, 0, 9)
	c.padded.DrawAt(c.got, c.source, 0, 9)
	want, got := make([]byte, 32*32*4), make([]byte, 32*32*4)
	c.want.ReadPixels(want)
	c.got.ReadPixels(got)
	if !bytes.Equal(want, got) {
		c.err = fmt.Errorf("CRT source origin differs from an authored padded source at frame %d", c.frames)
	}
}

func TestMain(m *testing.M) {
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	directory, err := os.MkdirTemp("", "dck-crt-origin-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var check *crtOriginCheck
	err = capture.Run(capture.Config{Directory: directory, Frames: []int{0, 1}, Width: 32, Height: 32}, func() (ebiten.Game, error) {
		config := CRTOverlayConfig{Curvature: .15, ScanlineFrequency: 800,
			ScanlineAmplitude: .04, ChromaticShift: .002, Vignette: .5}
		reference, err := NewCRTOverlay(config)
		if err != nil {
			return nil, err
		}
		config.SourceOrigin = image.Pt(4, 3)
		padded, err := NewCRTOverlay(config)
		if err != nil {
			return nil, err
		}
		root := ebiten.NewImage(32, 16)
		manual := ebiten.NewImageWithOptions(image.Rect(0, 0, 28, 15), &ebiten.NewImageOptions{Unmanaged: true})
		check = &crtOriginCheck{
			sourceRoot: root, source: root.SubImage(image.Rect(3, 2, 27, 14)).(*ebiten.Image),
			manualBacking: manual, manualView: manual.SubImage(image.Rect(4, 3, 28, 15)).(*ebiten.Image),
			want: ebiten.NewImage(32, 32), got: ebiten.NewImage(32, 32),
			reference: reference, padded: padded,
		}
		return check, nil
	})
	if check != nil {
		check.reference.Close()
		check.padded.Close()
		check.sourceRoot.Deallocate()
		check.manualBacking.Deallocate()
		check.want.Deallocate()
		check.got.Deallocate()
		if err == nil {
			err = check.err
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
