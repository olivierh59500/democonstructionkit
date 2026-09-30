//go:build dck_gpu_rendercheck

package effects

import (
	"errors"
	"image"
	"image/color"
	"math"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/fidelity/ebiten/testutil"
)

func indexedBankPixels(t *testing.T, target *ebiten.Image, want func(int, int) color.NRGBA) {
	t.Helper()
	bounds := target.Bounds()
	pixels := make([]byte, bounds.Dx()*bounds.Dy()*4)
	target.ReadPixels(pixels)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			expected := want(x, y)
			at := ((y-bounds.Min.Y)*bounds.Dx() + x - bounds.Min.X) * 4
			if pixels[at] != expected.R || pixels[at+1] != expected.G || pixels[at+2] != expected.B || pixels[at+3] != expected.A {
				t.Fatalf("indexed bank pixel %d,%d: got %v, want %v", x, y, pixels[at:at+4], expected)
			}
		}
	}
}

func TestIndexedImageBankCropsAtlasOriginsCopiesSlotsAndUpdatesLivePalettesGPU(t *testing.T) {
	testutil.RequireGPU(t)
	atlas := ebiten.NewImage(9, 4)
	defer atlas.Deallocate()
	a, b := [2][4]byte{{0, 1, 2, 3}, {3, 2, 1, 0}}, [2][3]byte{{3, 0, 1}, {1, 2, 0}}
	pixels := make([]byte, 9*4*4)
	for y := range 2 {
		for x, value := range a[y] {
			at := ((y+1)*9 + x + 1) * 4
			pixels[at], pixels[at+3] = value, 255
		}
		for x, value := range b[y] {
			at := ((y+1)*9 + x + 5) * 4
			pixels[at], pixels[at+3] = value, 255
		}
	}
	atlas.WritePixels(pixels)
	first := atlas.SubImage(image.Rect(1, 1, 5, 3)).(*ebiten.Image)
	second := atlas.SubImage(image.Rect(5, 1, 8, 3)).(*ebiten.Image)
	firstColors := []color.NRGBA{{R: 255, A: 255}, {G: 255, A: 255}, {B: 255, A: 255}, {R: 255, G: 255, B: 255, A: 255}}
	secondColors := []color.NRGBA{{R: 40, G: 80, B: 120, A: 255}, {R: 60, G: 100, B: 140, A: 255}, {R: 80, G: 120, B: 160, A: 255}, {R: 100, G: 140, B: 180, A: 255}}
	images := []*ebiten.Image{first, second}
	palettes := []IndexedImagePalette{{Name: "primary", Colors: firstColors}, {Name: "secondary", Colors: secondColors}}
	slots := []IndexedImageSlot{{Image: 0, Palette: 0, Crop: image.Rect(1, 0, 3, 2)}, {Image: 1, Palette: 1}, {Image: 0, Palette: 1, Crop: image.Rect(0, 0, 1, 2)}}
	slots[0].Options.GeoM.Scale(2, 2)
	slots[0].Options.GeoM.Translate(2, 2)
	slots[1].Options.GeoM.Translate(8, 1)
	slots[2].Options.GeoM.Translate(13, 2)
	slots[2].Options.ColorScale.Scale(.5, 1, 1, 1)
	retained := append([]IndexedImageSlot(nil), slots...)
	bank, err := NewIndexedImageBank(IndexedImageBankConfig{Images: images, Palettes: palettes, Slots: slots, Channel: composite.BitplaneRed, Scale: 255})
	if err != nil {
		t.Fatal(err)
	}
	defer bank.Close()
	images[0], palettes[1].Name, firstColors[0] = nil, "changed", color.NRGBA{}
	slots[0].Options.GeoM.Translate(100, 0)
	if index, found := bank.PaletteIndex("secondary"); !found || index != 1 {
		t.Fatal("input mutation changed the named palette bank", index, found)
	}
	if _, found := bank.PaletteIndex("changed"); found {
		t.Fatal("palette names stayed bound to caller storage")
	}
	target := ebiten.NewImage(17, 9)
	defer target.Deallocate()
	view := target.SubImage(image.Rect(1, 1, 16, 8)).(*ebiten.Image)
	verify := func(colors []color.NRGBA) {
		t.Helper()
		target.Clear()
		bank.Draw(view)
		if bank.Err() != nil {
			t.Fatal(bank.Err())
		}
		indexedBankPixels(t, target, func(x, y int) color.NRGBA {
			if x >= 2 && x < 6 && y >= 2 && y < 6 {
				if a[(y-2)/2][1+(x-2)/2] == 1 {
					return color.NRGBA{G: 255, A: 255}
				}
				return color.NRGBA{B: 255, A: 255}
			}
			if x >= 8 && x < 11 && y >= 1 && y < 3 {
				return colors[b[y-1][x-8]]
			}
			if x == 13 && y >= 2 && y < 4 {
				paint := colors[a[y-2][0]]
				paint.R /= 2
				return paint
			}
			return color.NRGBA{}
		})
	}
	verify(secondColors)
	updated := []color.NRGBA{{R: 120, G: 30, B: 40, A: 255}, {R: 140, G: 50, B: 60, A: 255}, {R: 160, G: 70, B: 80, A: 255}, {R: 180, G: 90, B: 100, A: 255}}
	expected := append([]color.NRGBA(nil), updated...)
	if err := bank.SetPalette(1, updated); err != nil {
		t.Fatal(err)
	}
	updated[0] = color.NRGBA{}
	verify(expected)
	if allocations := testing.AllocsPerRun(100, func() {
		if err := bank.SetSlots(retained); err != nil {
			panic(err)
		}
	}); allocations != 0 {
		t.Fatalf("unchanged indexed selection allocates: %v", allocations)
	}
	bank.Draw(first)
	if bank.Err() == nil {
		t.Fatal("drawing into borrowed artwork was accepted")
	}
	verify(expected)
	bank.Close()
	bank.Close()
	atlas.Fill(color.NRGBA{R: 91, A: 255})
	target.Clear()
	bank.Draw(target)
	indexedBankPixels(t, target, func(int, int) color.NRGBA { return color.NRGBA{} })
	target.DrawImage(first, nil)
	indexedBankPixels(t, target, func(x, y int) color.NRGBA {
		if x < 4 && y < 2 {
			return color.NRGBA{R: 91, A: 255}
		}
		return color.NRGBA{}
	})
	if bank.SetSlots(nil) == nil || bank.SetPalette(0, expected) == nil || bank.Update(kit.Frame{}) == nil {
		t.Fatal("closed indexed bank accepted state changes")
	}
}

func TestIndexedImageBankSelectorClockAndFailedWindowsAreAtomicGPU(t *testing.T) {
	testutil.RequireGPU(t)
	source := indexedImageSource(2, 1, []byte{0, 1})
	defer source.Deallocate()
	images := []*ebiten.Image{source.SubImage(image.Rect(0, 0, 1, 1)).(*ebiten.Image), source.SubImage(image.Rect(1, 0, 2, 1)).(*ebiten.Image)}
	palettes := []IndexedImagePalette{{Colors: []color.NRGBA{{R: 255, A: 255}, {B: 255, A: 255}}}, {Colors: []color.NRGBA{{G: 255, A: 255}, {R: 255, G: 255, A: 255}}}}
	calls := 0
	wantError := errors.New("invalid authored cue")
	slots := []IndexedImageSlot{{Image: 0, Palette: 0}, {Image: 1, Palette: 1, Hidden: true}}
	slots[1].Options.GeoM.Translate(1, 0)
	bank, err := NewIndexedImageBank(IndexedImageBankConfig{Images: images, Palettes: palettes, Slots: slots, MaxSlots: 2,
		Channel: composite.BitplaneRed, Scale: 255, Select: func(frame kit.Frame, output []IndexedImageSlot) error {
			calls++
			output[0].Image = int(frame.Tick % 2)
			output[1].Hidden = frame.Tick < 2
			if frame.Tick == 3 {
				output[1].Image = 99
			}
			if frame.Tick == 4 {
				return wantError
			}
			return nil
		}})
	if err != nil {
		t.Fatal(err)
	}
	defer bank.Close()
	target := ebiten.NewImage(2, 1)
	defer target.Deallocate()
	if err := bank.Update(kit.Frame{Tick: 2, Time: .04}); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		target.Clear()
		bank.Draw(target)
		indexedBankPixels(t, target, func(x, _ int) color.NRGBA {
			if x == 0 {
				return color.NRGBA{R: 255, A: 255}
			}
			return color.NRGBA{R: 255, G: 255, A: 255}
		})
	}
	if calls != 1 {
		t.Fatal("Draw advanced the authored selector", calls)
	}
	if bank.Update(kit.Frame{Tick: 3, Time: .06}) == nil || !errors.Is(bank.Update(kit.Frame{Tick: 4, Time: .08}), wantError) || bank.Update(kit.Frame{Time: math.NaN()}) == nil {
		t.Fatal("invalid selection or selector error was lost")
	}
	if calls != 3 {
		t.Fatal("invalid clock called the selector", calls)
	}
	target.Clear()
	bank.Draw(target)
	indexedBankPixels(t, target, func(x, _ int) color.NRGBA {
		if x == 0 {
			return color.NRGBA{R: 255, A: 255}
		}
		return color.NRGBA{R: 255, G: 255, A: 255}
	})
	if err := bank.SetSlots([]IndexedImageSlot{{Image: -1, Palette: -1, Hidden: true}}); err != nil {
		t.Fatal("hidden authored slots required ready artwork", err)
	}
	target.Clear()
	bank.Draw(target)
	indexedBankPixels(t, target, func(int, int) color.NRGBA { return color.NRGBA{} })
}

func TestIndexedImageBankSourcePaletteAndSlotAlphaWithCopyBlendGPU(t *testing.T) {
	testutil.RequireGPU(t)
	source := ebiten.NewImage(1, 1)
	defer source.Deallocate()
	source.WritePixels([]byte{1, 0, 0, 128})
	slot := IndexedImageSlot{}
	slot.Options.Blend = ebiten.BlendCopy
	slot.Options.ColorScale.ScaleAlpha(.5)
	bank, err := NewIndexedImageBank(IndexedImageBankConfig{Images: []*ebiten.Image{source},
		Palettes: []IndexedImagePalette{{Colors: []color.NRGBA{{}, {R: 255, G: 128, B: 64, A: 128}}}},
		Slots:    []IndexedImageSlot{slot}, Channel: composite.BitplaneRed, Scale: 255, SourceAlpha: true})
	if err != nil {
		t.Fatal(err)
	}
	defer bank.Close()
	target := ebiten.NewImage(1, 1)
	defer target.Deallocate()
	target.Fill(color.NRGBA{B: 255, A: 255})
	bank.Draw(target)
	// Independent premultiplication: 128/255 palette alpha, 128/255 source
	// alpha and one-half slot opacity. Copy must replace the blue destination.
	indexedImagePixel(t, target, [4]byte{32, 16, 8, 32})
}

func TestIndexedImageBankRejectsInvalidBanksAndKeepsPreviousSlotsGPU(t *testing.T) {
	testutil.RequireGPU(t)
	source := indexedImageSource(1, 1, []byte{0})
	defer source.Deallocate()
	palette := IndexedImagePalette{Name: "one", Colors: []color.NRGBA{{R: 255, A: 255}}}
	valid := IndexedImageBankConfig{Images: []*ebiten.Image{source}, Palettes: []IndexedImagePalette{palette}, Slots: []IndexedImageSlot{{}}}
	for _, config := range []IndexedImageBankConfig{
		{}, {Images: []*ebiten.Image{nil}, Palettes: valid.Palettes}, {Images: valid.Images},
		{Images: make([]*ebiten.Image, 65537), Palettes: valid.Palettes}, {Images: valid.Images, Palettes: make([]IndexedImagePalette, 257)},
		{Images: valid.Images, Palettes: []IndexedImagePalette{{}}}, {Images: valid.Images, Palettes: []IndexedImagePalette{{Colors: make([]color.NRGBA, 257)}}},
		{Images: valid.Images, Palettes: []IndexedImagePalette{palette, palette}},
		{Images: valid.Images, Palettes: []IndexedImagePalette{palette, {Colors: []color.NRGBA{{}, {}}}}},
		{Images: valid.Images, Palettes: []IndexedImagePalette{{Name: strings.Repeat("x", 1025), Colors: palette.Colors}}},
		{Images: valid.Images, Palettes: valid.Palettes, MaxSlots: -1}, {Images: valid.Images, Palettes: valid.Palettes, MaxSlots: 65537},
		{Images: valid.Images, Palettes: valid.Palettes, MaxSlots: 1, Slots: []IndexedImageSlot{{}, {}}},
		{Images: valid.Images, Palettes: valid.Palettes, Channel: composite.BitplaneBlue + 1},
		{Images: valid.Images, Palettes: valid.Palettes, Scale: -1}, {Images: valid.Images, Palettes: valid.Palettes, Offset: float32(math.NaN())},
	} {
		if bank, err := NewIndexedImageBank(config); err == nil {
			bank.Close()
			t.Fatal("accepted invalid indexed artwork/palette/slot bounds")
		}
	}
	bank, err := NewIndexedImageBank(valid)
	if err != nil {
		t.Fatal(err)
	}
	defer bank.Close()
	nan, huge := IndexedImageSlot{}, IndexedImageSlot{}
	nan.Options.GeoM.Translate(math.NaN(), 0)
	huge.Options.GeoM.Translate(math.MaxFloat32*2, 0)
	for _, slot := range []IndexedImageSlot{{Image: -1}, {Image: 1}, {Palette: -1}, {Palette: 1}, {Crop: image.Rect(-1, 0, 1, 1)}, {Crop: image.Rect(0, 0, 2, 1)}, nan, huge} {
		if bank.SetSlots([]IndexedImageSlot{{Hidden: true}, slot}) == nil {
			t.Fatal("accepted invalid second slot or partially replaced the old window", slot)
		}
	}
	if bank.SetPalette(-1, palette.Colors) == nil || bank.SetPalette(0, nil) == nil || bank.SetPalette(1, palette.Colors) == nil {
		t.Fatal("accepted invalid live palette update")
	}
	target := ebiten.NewImage(1, 1)
	defer target.Deallocate()
	bank.Draw(target)
	indexedImagePixel(t, target, [4]byte{255, 0, 0, 255})
	if err := bank.SetSlots(nil); err != nil {
		t.Fatal(err)
	}
	target.Clear()
	bank.Draw(target)
	indexedImagePixel(t, target, [4]byte{})
}
