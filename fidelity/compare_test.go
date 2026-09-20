package fidelity

import (
	"image"
	"image/color"
	"testing"
)

func TestComparisonDetectsAChangedPixelAndNormalizesBounds(t *testing.T) {
	a := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	b := image.NewNRGBA(image.Rect(10, 20, 12, 21))
	for x := 0; x < 2; x++ {
		a.SetNRGBA(x, 0, color.NRGBA{10, 20, 30, 255})
		b.SetNRGBA(x+10, 20, color.NRGBA{10, 20, 30, 255})
	}
	r, _, err := Compare(a, b)
	if err != nil || r.DifferentPixels != 0 {
		t.Fatal(r, err)
	}
	b.SetNRGBA(11, 20, color.NRGBA{15, 20, 30, 255})
	r, diff, err := Compare(a, b)
	if err != nil || r.DifferentPixels != 1 || r.MaxChannelError != 5 || r.MeanAbsoluteError != .625 || diff.RGBAAt(1, 0).R != 40 {
		t.Fatal(r, err)
	}
	if _, _, err = Compare(a, image.NewRGBA(image.Rect(0, 0, 3, 1))); err == nil {
		t.Fatal("dimension mismatch was accepted")
	}
}
