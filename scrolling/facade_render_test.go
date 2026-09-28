//go:build dck_scroll_facadecheck

package scrolling

import (
	"bytes"
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/fidelity/ebiten/testutil"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestMain(m *testing.M) { os.Exit(testutil.RunGPU(m)) }

func facadeFont(c color.Color, width, height int) BitmapGrid {
	image := ebiten.NewImage(width*4, height)
	image.Fill(c)
	return BitmapGrid{Image: image, Width: float64(width), Height: float64(height), Columns: 4, First: 'A', Filter: ebiten.FilterLinear}
}

func requireFacadePixels(t *testing.T, reference, candidate *ebiten.Image) {
	t.Helper()
	want, got := make([]byte, reference.Bounds().Dx()*reference.Bounds().Dy()*4), make([]byte, candidate.Bounds().Dx()*candidate.Bounds().Dy()*4)
	reference.ReadPixels(want)
	candidate.ReadPixels(got)
	if !bytes.Equal(want, got) {
		t.Fatal("facade changed complete rendered pixels")
	}
}

func TestFacadeRecycledOffsetAndVerticalCoordinates(t *testing.T) {
	font := facadeFont(color.White, 7, 5)
	defer font.Image.Deallocate()
	config := RingConfig{Font: font, Text: "AB^S2CDAB", Viewport: 35, Speed: 3, Controls: true,
		Waves: []RingWave{{Amplitude: 2, LetterStep: .13, TickStep: .03}}}
	for _, vertical := range []bool{false, true} {
		scroll, err := New(Config{X: 2, Y: 1, Recycled: &RecycledConfig{Ring: config, Vertical: vertical}})
		if err != nil {
			t.Fatal(err)
		}
		reference, err := NewRing(config)
		if err != nil {
			t.Fatal(err)
		}
		want, got := ebiten.NewImage(64, 64), ebiten.NewImage(64, 64)
		for tick := 0; tick < 900; tick++ {
			reference.Step()
			if err := scroll.Update(kit.Frame{Tick: uint64(tick)}); err != nil {
				t.Fatal(err)
			}
			if scroll.CursorRune() != reference.NextRune() || scroll.RecycledController().Cursor() != reference.Cursor() {
				t.Fatalf("recycled cursor changed at %d", tick)
			}
			if tick%29 != 0 {
				continue
			}
			want.Clear()
			got.Clear()
			if !vertical {
				reference.DrawAt(want, 2+3.25, 1+4.5)
			} else {
				for _, index := range reference.order {
					letter := reference.letters[index]
					region, ok := font.Region(letter.Rune)
					if !ok {
						continue
					}
					options := ebiten.DrawImageOptions{Filter: font.Filter}
					options.GeoM.Translate(2+letter.Y+3.25, 1+letter.X+4.5)
					composite.DrawRegion(want, font.Image, region, &options)
				}
			}
			if err := scroll.DrawOffset(got, 3.25, 4.5); err != nil {
				t.Fatal(err)
			}
			requireFacadePixels(t, want, got)
		}
		want.Deallocate()
		got.Deallocate()
		scroll.Close()
	}
}

func TestFacadeDualProfileWithIndependentFontsAndOutputPasses(t *testing.T) {
	front := facadeFont(color.RGBA{R: 220, A: 255}, 7, 5)
	back := facadeFont(color.RGBA{B: 220, A: 255}, 9, 6)
	raster := ebiten.NewImage(1, 40)
	raster.Fill(color.RGBA{G: 140, A: 255})
	defer front.Image.Deallocate()
	defer back.Image.Deallocate()
	defer raster.Deallocate()
	config := DualProfiledRingConfig{
		Front: RingConfig{Font: front, Text: "ABCDAB", Viewport: 48, Speed: 2},
		Back:  RingConfig{Font: back, Text: "DCBADC", Viewport: 48, Speed: 3},
		Profile: motion.SegmentedProfileConfig{Segments: []motion.ProfileSegment{{Samples: 64, Amplitude: 5}},
			Start: 0, Restart: 0, WrapMargin: 12, VisibleSamples: 12},
		Raster: composite.RasterOverlayConfig{Image: raster, ScaleX: 48, ScaleY: 1, Alpha: 1, Blend: ebiten.BlendSourceIn},
		Width:  48, RingHeight: 6, MaskHeight: 40, StripCount: 12, StripWidth: 4,
		SourceHeight: 6, BaseY: 12, BackFirstFilter: ebiten.FilterNearest,
		BackFilter: ebiten.FilterLinear, FrontFilter: ebiten.FilterLinear, Unmanaged: true,
	}
	reference, err := NewDualProfiledRing(config)
	if err != nil {
		t.Fatal(err)
	}
	defer reference.Close()
	passes := 0
	scroll, err := New(Config{DualProfiled: &config, Output: &OutputConfig{Width: 48, Height: 40,
		Passes: []kit.ImagePass{{Apply: func(dst, src *ebiten.Image, _ kit.Frame) { passes++; dst.DrawImage(src, nil) }}}}})
	if err != nil {
		t.Fatal(err)
	}
	defer scroll.Close()
	want, got := ebiten.NewImage(48, 40), ebiten.NewImage(48, 40)
	defer want.Deallocate()
	defer got.Deallocate()
	for tick := 0; tick < 900; tick++ {
		reference.Update(kit.Frame{})
		if err := scroll.Update(kit.Frame{Tick: uint64(tick)}); err != nil {
			t.Fatal(err)
		}
		if scroll.DualProfiledController().FrontRing().Cursor() != reference.FrontRing().Cursor() ||
			scroll.DualProfiledController().BackRing().Cursor() != reference.BackRing().Cursor() {
			t.Fatal("independent font cursors changed")
		}
		if tick%31 == 0 {
			want.Clear()
			got.Clear()
			reference.Draw(want)
			scroll.Draw(got)
			requireFacadePixels(t, want, got)
			got.Clear()
			scroll.Draw(got)
			requireFacadePixels(t, want, got)
		}
	}
	if passes == 0 {
		t.Fatal("selected transport bypassed the configured image passes")
	}
	if err := scroll.DrawOffset(got, 1, 1); err == nil {
		t.Fatal("offset silently bypassed post-processing")
	}
}

func TestFacadeCaptionAndRevealKeepTheirAuthoredClocks(t *testing.T) {
	font := facadeFont(color.White, 7, 5)
	defer font.Image.Deallocate()
	config := CaptionCarouselConfig{Lines: []string{"AB", "CD", "ABCD"}, Font: font,
		Motion:     motion.CaptionCycleConfig{Top: -7, Bottom: 2, StartY: -7, Speed: 1, InitialWait: 7, HoldWait: 3},
		Background: image.Rect(0, 0, 48, 12), BackgroundColor: color.Black, CenterX: 24, ScaleX: 1, ScaleY: 1}
	caption, err := NewCaptionCarousel(config)
	if err != nil {
		t.Fatal(err)
	}
	scroll, err := New(Config{Caption: &config})
	if err != nil {
		t.Fatal(err)
	}
	defer scroll.Close()
	want, got := ebiten.NewImage(48, 40), ebiten.NewImage(48, 40)
	defer want.Deallocate()
	defer got.Deallocate()
	for tick := 0; tick < 600; tick++ {
		caption.Step()
		scroll.Update(kit.Frame{Tick: uint64(tick)})
		if scroll.CaptionController().Pose() != caption.Pose() {
			t.Fatal("caption slide/hold/exit clock changed")
		}
		if tick%13 == 0 {
			want.Fill(color.RGBA{B: 80, A: 255})
			got.Fill(color.RGBA{B: 80, A: 255})
			caption.Draw(want)
			scroll.Draw(got)
			requireFacadePixels(t, want, got)
		}
	}
	revealConfig := RevealConfig{Font: font, Lines: []string{"ABCD", "DCBA"}, Columns: 4,
		X: 2, Y: 2, FromY: 30, Delay: 3, Duration: 11, Order: RevealColumnsBottomFirst}
	reveal, err := NewReveal(revealConfig)
	if err != nil {
		t.Fatal(err)
	}
	clockSamples := 0
	animated, err := New(Config{Reveal: &RevealTransportConfig{Reveal: revealConfig,
		TimeAt: func(frame kit.Frame) float64 { clockSamples++; return float64(frame.Tick) * .7 }}})
	if err != nil {
		t.Fatal(err)
	}
	defer animated.Close()
	for tick := 0; tick < 100; tick++ {
		animated.Update(kit.Frame{Tick: uint64(tick)})
		want.Clear()
		got.Clear()
		reveal.DrawAt(want, float64(tick)*.7)
		animated.Draw(got)
		requireFacadePixels(t, want, got)
		got.Clear()
		animated.Draw(got)
		requireFacadePixels(t, want, got)
	}
	if clockSamples != 100 {
		t.Fatal("drawing reevaluated the reveal clock", clockSamples)
	}
}
