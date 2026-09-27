//go:build dck_crt_storagecheck

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

type crtStorageCheck struct {
	pass    *CRTOverlay
	checked bool
	err     error
}

func (check *crtStorageCheck) Layout(int, int) (int, int) { return 800, 600 }
func (check *crtStorageCheck) Update() error              { return nil }
func (check *crtStorageCheck) Draw(*ebiten.Image) {
	if check.checked {
		return
	}
	check.checked = true
	managed := ebiten.NewImage(800, 72)
	unmanaged := ebiten.NewImageWithOptions(image.Rect(0, 0, 800, 72), &ebiten.NewImageOptions{Unmanaged: true})
	first := ebiten.NewImageWithOptions(image.Rect(0, 0, 800, 600), &ebiten.NewImageOptions{Unmanaged: true})
	second := ebiten.NewImageWithOptions(image.Rect(0, 0, 800, 600), &ebiten.NewImageOptions{Unmanaged: true})
	defer managed.Deallocate()
	defer unmanaged.Deallocate()
	defer first.Deallocate()
	defer second.Deallocate()
	for material := 0; material < 3; material++ {
		managed.Clear()
		unmanaged.Clear()
		switch material {
		case 1:
			managed.SubImage(image.Rect(50, 20, 150, 50)).(*ebiten.Image).Fill(color.RGBA{0, 220, 100, 255})
			unmanaged.SubImage(image.Rect(50, 20, 150, 50)).(*ebiten.Image).Fill(color.RGBA{0, 220, 100, 255})
		case 2:
			managed.Fill(color.RGBA{255, 120, 20, 255})
			unmanaged.Fill(color.RGBA{255, 120, 20, 255})
		}
		first.Fill(color.Black)
		second.Fill(color.Black)
		check.pass.DrawAt(first, managed, 0, 264)
		check.pass.DrawAt(second, unmanaged, 0, 264)
		a, b := make([]byte, 800*600*4), make([]byte, 800*600*4)
		first.ReadPixels(a)
		second.ReadPixels(b)
		if !bytes.Equal(a, b) {
			check.err = fmt.Errorf("normalized CRT depends on source storage for material %d", material)
			return
		}
	}
}

func TestMain(m *testing.M) {
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	directory, err := os.MkdirTemp("", "dck-crt-storage-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var check *crtStorageCheck
	err = capture.Run(capture.Config{Directory: directory, Frames: []int{0}, Width: 800, Height: 600}, func() (ebiten.Game, error) {
		pass, err := NewCRTOverlay(CRTOverlayConfig{
			Curvature: .15, ScanlineFrequency: 800, ScanlineAmplitude: .04,
			ChromaticShift: .002, Vignette: .5, NormalizeSource: true,
		})
		if err != nil {
			return nil, err
		}
		check = &crtStorageCheck{pass: pass}
		return check, nil
	})
	if check != nil {
		_ = check.pass.Close()
		if err == nil {
			err = check.err
		}
		if err == nil && !check.checked {
			err = fmt.Errorf("CRT storage check was not drawn")
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
