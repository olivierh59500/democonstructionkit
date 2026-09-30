package composite

import (
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestBitplanePaletteValidationAndColorUpdates(t *testing.T) {
	colors := []color.NRGBA{{R: 200, A: 255}, {G: 180, A: 128}}
	for _, config := range []BitplanePaletteConfig{
		{},
		{Width: 8, Height: 8, Planes: 0, Palette: colors},
		{Width: 8, Height: 8, Planes: 7, Palette: colors},
		{Width: 8, Height: 8, Planes: 1, Palette: colors[:1]},
		{Width: 8192, Height: 8192, Planes: 1, Palette: colors},
		{Width: 8, Height: 8, Planes: 1, Palette: colors, Threshold: float32(math.NaN())},
		{Width: 8, Height: 8, Planes: 1, Palette: colors, Threshold: 1.1},
	} {
		if effect, err := NewBitplanePalette(config); err == nil {
			effect.Close()
			t.Fatalf("accepted invalid bitplane config %+v", config)
		}
	}
	effect, err := NewBitplanePalette(BitplanePaletteConfig{Width: 8, Height: 8, Planes: 1, Palette: colors})
	if err != nil {
		t.Fatal(err)
	}
	defer effect.Close()
	if effect.threshold != 0.5 || effect.packed != nil {
		t.Fatal("single plane should use the default threshold and one draw pass")
	}
	if effect.palette[0] != float32(200)/255 || effect.palette[3] != 1 ||
		math.Abs(float64(effect.palette[5]-float32(180)/255*(float32(128)/255))) > 1e-6 ||
		effect.palette[7] != float32(128)/255 {
		t.Fatal("palette did not convert to premultiplied shader colors")
	}
	if err := effect.SetPalette(colors[:1]); err == nil {
		t.Fatal("accepted incomplete palette update")
	}
	if err := effect.Draw(nil, nil); err == nil {
		t.Fatal("accepted missing target")
	}
	target := ebiten.NewImage(8, 8)
	defer target.Deallocate()
	if err := effect.Draw(target, []*ebiten.Image{nil}); err == nil {
		t.Fatal("accepted nil source plane")
	}
	wrongSize := ebiten.NewImage(7, 8)
	defer wrongSize.Deallocate()
	if err := effect.Draw(target, []*ebiten.Image{wrongSize}); err == nil {
		t.Fatal("accepted wrong-sized source plane")
	}
	if err := effect.Close(); err != nil {
		t.Fatal(err)
	}
	if err := effect.SetPalette(colors); err == nil {
		t.Fatal("updated a closed palette")
	}
}
