//go:build dck_gpu_rendercheck

package scrolling

import (
	"bytes"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/scrolltext"
)

func TestInsertionGPUMixedFontsBearingsAndDrawClock(t *testing.T) {
	a, b := ebiten.NewImage(2, 2), ebiten.NewImage(3, 3)
	defer a.Deallocate()
	defer b.Deallocate()
	a.Fill(color.NRGBA{R: 255, A: 255})
	b.Fill(color.NRGBA{B: 255, A: 255})
	ma, err := font.New(font.Config{Bounds: a.Bounds(), LineHeight: 2, SpaceAdvance: 3, Glyphs: map[rune]font.Glyph{'A': {Rect: a.Bounds(), Advance: 3, OffsetX: 1}}})
	if err != nil {
		t.Fatal(err)
	}
	mb, err := font.New(font.Config{Bounds: b.Bounds(), LineHeight: 3, SpaceAdvance: 4, Glyphs: map[rune]font.Glyph{'B': {Rect: b.Bounds(), Advance: 4, OffsetY: 1}}})
	if err != nil {
		t.Fatal(err)
	}
	s, err := New(Config{Insertion: &InsertionConfig{Program: scrolltext.InsertionProgramConfig{Tokens: []scrolltext.InsertionToken{{Glyph: true, Rune: 'A', Font: "small", Advance: 3}, {Glyph: true, Rune: 'B', Font: "large", Advance: 4}}, Speed: 1, TargetSpeed: 1, Entry: 8}, Fonts: map[string]Face{"small": {Atlas: a, Metrics: ma}, "large": {Atlas: b, Metrics: mb}}, Y: 1}})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for tick := 0; tick < 4; tick++ {
		if err := s.Update(kit.Frame{Tick: uint64(tick)}); err != nil {
			t.Fatal(err)
		}
	}
	dst := ebiten.NewImage(16, 8)
	defer dst.Deallocate()
	before := s.InsertionController().State()
	s.Draw(dst)
	pixels := make([]byte, 16*8*4)
	dst.ReadPixels(pixels)
	// At tick 4 A's pen is 4 plus its one-pixel bearing; B enters at 7 with a
	// one-pixel Y bearing. Rendering must keep those independent metrics.
	for y := 0; y < 8; y++ {
		for x := 0; x < 16; x++ {
			want := [4]byte{}
			if x >= 5 && x < 7 && y >= 1 && y < 3 {
				want = [4]byte{255, 0, 0, 255}
			}
			if x >= 7 && x < 10 && y >= 2 && y < 5 {
				want = [4]byte{0, 0, 255, 255}
			}
			at := (y*16 + x) * 4
			if !bytes.Equal(pixels[at:at+4], want[:]) {
				t.Fatal("mixed insertion pixel", x, y, pixels[at:at+4], want)
			}
		}
	}
	dst.Clear()
	s.Draw(dst)
	again := make([]byte, len(pixels))
	dst.ReadPixels(again)
	after := s.InsertionController().State()
	if before.Distance != after.Distance || before.Cursor != after.Cursor || !bytes.Equal(pixels, again) {
		t.Fatal("Draw advanced insertion")
	}
	s.Close()
	a.Fill(color.White)
	dst.Clear()
	dst.DrawImage(a, nil)
	dst.ReadPixels(again)
	if again[0] != 255 {
		t.Fatal("close invalidated borrowed font")
	}
	if _, err := New(Config{Insertion: &InsertionConfig{Fonts: map[string]Face{"default": {Atlas: a, Metrics: ma}}}, Text: "A"}); err == nil {
		t.Fatal("mixedregular/insertion config accepted")
	}
}
