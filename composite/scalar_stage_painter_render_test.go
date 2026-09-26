//go:build dck_scalar_stage_rendercheck && !dck_tiled_wave_rendercheck && !dck_copper_title_rendercheck && !dck_window_bank_rendercheck && !dck_background_rendercheck && !dck_water_rendercheck

package composite_test

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/presets"
	"github.com/olivierh59500/democonstructionkit/timeline"
)

// Opt-in pixel comparison against the previous ordered Phenomena intro/outro.
func TestMain(m *testing.M) {
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	game := &scalarStageRenderCheck{}
	ebiten.SetWindowSize(320, 240)
	if err := ebiten.RunGame(game); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if game.err != nil {
		fmt.Fprintln(os.Stderr, game.err)
		os.Exit(1)
	}
}

type scalarStageRenderCheck struct {
	done bool
	err  error
}

func (*scalarStageRenderCheck) Layout(int, int) (int, int) { return 640, 480 }
func (game *scalarStageRenderCheck) Update() error {
	if game.done {
		return ebiten.Termination
	}
	return nil
}
func (game *scalarStageRenderCheck) Draw(*ebiten.Image) {
	if game.done {
		return
	}
	game.done = true
	game.err = compareScalarStagePixels()
}

func stageTestImage(width, height, seed int) *ebiten.Image {
	pixels := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			alpha := uint8(255)
			if (x/9+y/7+seed)%11 == 0 {
				alpha = 0
			}
			pixels.SetNRGBA(x, y, color.NRGBA{R: uint8(x*3 + seed), G: uint8(y*5 + seed), B: uint8(x + y + seed*13), A: alpha})
		}
	}
	return ebiten.NewImageFromImage(pixels)
}

func compareScalarStagePixels() error {
	images := presets.PhenomenaStageImages{
		RasterBar: stageTestImage(640, 12, 1), Page1: stageTestImage(640, 480, 2),
		Page2: stageTestImage(640, 480, 3), Logo: stageTestImage(640, 300, 4),
		LogoMask: stageTestImage(640, 300, 5), Middle: stageTestImage(640, 300, 6),
		Raster: stageTestImage(640, 12, 7), Photon: stageTestImage(70, 15, 8),
	}
	for _, img := range []*ebiten.Image{images.RasterBar, images.Page1, images.Page2, images.Logo,
		images.LogoMask, images.Middle, images.Raster, images.Photon} {
		defer img.Deallocate()
	}
	actual := ebiten.NewImageWithOptions(image.Rect(0, 0, 640, 480), &ebiten.NewImageOptions{Unmanaged: true})
	expected := ebiten.NewImageWithOptions(image.Rect(0, 0, 640, 480), &ebiten.NewImageOptions{Unmanaged: true})
	defer actual.Deallocate()
	defer expected.Deallocate()
	got, want := make([]byte, 640*480*4), make([]byte, 640*480*4)
	samples := []struct {
		stage                       int
		value, direction, secondary float64
	}{
		{0, -40, 1, 184}, {0, 75, 1, 184},
		{1, 0, 1, 184}, {1, 100, 1, 184}, {1, 101, -1, 184}, {1, 200, -1, 184},
		{2, 0, 1, 184}, {2, 100, 1, 184}, {2, 101, 1, 184}, {2, 200, 1, 184},
		{3, 0, 1, 184}, {3, 100, 1, 184},
		{4, 0, 1, 184}, {4, 100, 1, 184},
		{5, 0, 1, 184}, {5, 0, 1, 445},
		{6, 100, 1, 445}, {6, 50, 1, 445},
		{8, 50, 1, 445}, {8, 50, -1, 445},
		{9, 100, -1, 445}, {9, 0, -1, 445},
		{10, 100, -1, 445}, {10, 0, -1, 445},
		{11, 0, -1, 445},
	}
	for _, sample := range samples {
		stages := make([]timeline.ScalarStage, presets.PhenomenaEnd+1)
		for i := range stages {
			stages[i].Name = fmt.Sprintf("stage-%d", i)
		}
		director, err := timeline.NewScalarStages(timeline.ScalarStagesConfig{
			Stages: stages, InitialStage: sample.stage,
			InitialValue: sample.value, InitialDirection: sample.direction,
		})
		if err != nil {
			return err
		}
		painter, err := composite.NewScalarStagePainter(presets.PhenomenaStageMaterials(director, images))
		if err != nil {
			return err
		}
		actual.Clear()
		expected.Clear()
		painter.Draw(actual, sample.secondary)
		drawOriginalStage(expected, images, sample.stage, sample.value, sample.direction, sample.secondary)
		actual.ReadPixels(got)
		expected.ReadPixels(want)
		if !bytes.Equal(got, want) {
			pixel := 0
			for got[pixel] == want[pixel] {
				pixel++
			}
			return fmt.Errorf("stage %d value %g direction %g differs at (%d,%d): got %v want %v",
				sample.stage, sample.value, sample.direction, (pixel/4)%640, pixel/(4*640),
				got[pixel/4*4:pixel/4*4+4], want[pixel/4*4:pixel/4*4+4])
		}
	}
	fmt.Printf("Scalar stage materials match the previous renderer in %d sampled states\n", len(samples))
	return nil
}

func drawOriginalStage(dst *ebiten.Image, images presets.PhenomenaStageImages, stage int, value, direction, secondary float64) {
	place := func(img *ebiten.Image, x, y float64) {
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(x, y)
		dst.DrawImage(img, &op)
	}
	fade := func(img *ebiten.Image, x, y, alpha float64) {
		var op ebiten.DrawImageOptions
		op.ColorScale.ScaleAlpha(float32(alpha))
		op.GeoM.Translate(x, y)
		dst.DrawImage(img, &op)
	}
	brightness := func(img *ebiten.Image, x, y, r, g, b float64) {
		var op ebiten.DrawImageOptions
		op.ColorScale.Scale(float32(r), float32(g), float32(b), 1)
		op.GeoM.Translate(x, y)
		dst.DrawImage(img, &op)
	}
	dst.Fill(color.Black)
	switch stage {
	case presets.PhenomenaTextPage1:
		place(images.RasterBar, 0, value)
		dst.DrawImage(images.Page1, nil)
	case presets.PhenomenaTextPage2:
		v := value / 100
		if v > 1 {
			v = 2 - v
		}
		brightness(images.Page2, 0, 0, v, v, v)
	case presets.PhenomenaShowLogo:
		dst.Fill(color.RGBA{R: 0, G: 1, B: 17, A: 255})
		if value <= 100 {
			v := value / 100
			brightness(images.Logo, 0, 0, v, v, v)
		} else {
			dst.DrawImage(images.Logo, nil)
			fade(images.LogoMask, 0, 0, (200-value)/100)
		}
	case presets.PhenomenaShowUpperRaster, presets.PhenomenaShowLowerRaster,
		presets.PhenomenaDropPhoton, presets.PhenomenaPhotonFade:
		place(images.Middle, 0, 130)
		dst.DrawImage(images.Logo, nil)
		alpha := 1.0
		if stage == presets.PhenomenaShowUpperRaster {
			alpha = value / 100
		}
		fade(images.Raster, 0, 129, alpha)
		if stage >= presets.PhenomenaShowLowerRaster {
			alpha = 1
			if stage == presets.PhenomenaShowLowerRaster {
				alpha = value / 100
			}
			fade(images.Raster, 0, 430, alpha)
		}
		if stage == presets.PhenomenaDropPhoton {
			place(images.Photon, 285, secondary)
		} else if stage == presets.PhenomenaPhotonFade {
			lightness := value / 100
			brightness(images.Photon, 285, 445, lightness, lightness*.5, lightness*.5)
		}
	case presets.PhenomenaHideLogo:
		if direction > 0 {
			dst.DrawImage(images.Logo, nil)
		}
		fade(images.LogoMask, 0, 0, value/100)
	case presets.PhenomenaHideLowerRaster:
		fade(images.Raster, 0, 430, value/100)
	case presets.PhenomenaHideUpperRaster:
		fade(images.Raster, 0, 129, value/100)
	}
}
