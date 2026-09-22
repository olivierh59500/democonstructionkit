package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/effects"
	"github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/render"
)

// These pixel assertions need RunGame's active graphics context. They cover
// short/reverse scrolls, overhanging glyphs, viewport clipping and batch flushing.
func checkRenderingPrimitives() error {
	if err := checkRepeatRendering(); err != nil {
		return err
	}
	if err := checkMagnifierRendering(); err != nil {
		return err
	}
	atlas := ebiten.NewImage(4, 4)
	defer atlas.Deallocate()
	atlas.Fill(color.White)
	dst := render.NewSurface(16, 8)
	defer dst.Deallocate()
	pixels := make([]byte, 16*8*4)
	f, err := font.New(font.Config{Bounds: atlas.Bounds(), Glyphs: map[rune]font.Glyph{'A': {Rect: atlas.Bounds(), Advance: 4}, 'B': {Rect: atlas.Bounds(), Advance: 2}}, SpaceAdvance: 2, LineHeight: 4})
	if err != nil {
		return err
	}
	for _, message := range []string{"A", " B"} {
		for _, speed := range []float64{-5, 5} {
			s, err := effects.NewScroller(atlas, f, effects.ScrollConfig{Width: 8, Height: 4, Message: message, Speed: speed})
			if err != nil {
				return err
			}
			for _, seconds := range []float64{0, 1, 1000} {
				dst.Clear()
				if err = s.Update(kit.Frame{Time: seconds}); err != nil {
					return err
				}
				s.Draw(dst)
				dst.ReadPixels(pixels)
				for y := 0; y < 8; y++ {
					for x := 0; x < 16; x++ {
						want := byte(0)
						if x < 8 && y < 4 {
							want = 255
						}
						for c := 0; c < 4; c++ {
							if pixels[(y*16+x)*4+c] != want {
								return fmt.Errorf("scroller %q speed %g time %g: wrong pixel at %d,%d", message, speed, seconds, x, y)
							}
						}
					}
				}
			}
		}
	}
	dst.Clear()
	batch := render.NewBatch(2)
	batch.Begin(dst, atlas)
	for y := 0; y < 8; y++ {
		for x := 0; x < 16; x++ {
			batch.Rect(float64(x), float64(y), 1, 1, atlas.Bounds(), color.White)
		}
	}
	batch.Flush()
	dst.ReadPixels(pixels)
	for _, v := range pixels {
		if v != 255 {
			return fmt.Errorf("triangle batch flush lost pixels")
		}
	}
	fmt.Println("OK rendering primitives: reverse/short text, overhang, clipping, batch flush")
	return nil
}
