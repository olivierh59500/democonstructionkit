package main

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// checkMagnifierRendering compares GPU pixels to an independent analytic model
// inside the active graphics context, including atlas origin and crop handling.
func checkMagnifierRendering() error {
	const size = 32
	cpu := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			cpu.SetRGBA(x, y, color.RGBA{byte(x * 3), byte(y * 3), byte((x + y) * 2), 255})
		}
	}
	atlas := ebiten.NewImageFromImage(cpu)
	defer atlas.Deallocate()
	source := atlas.SubImage(image.Rect(7, 9, 39, 41)).(*ebiten.Image)
	dst := ebiten.NewImage(size, size)
	defer dst.Deallocate()
	lens, err := composite.NewMagnifier()
	if err != nil {
		return err
	}
	defer lens.Close()
	base := composite.DefaultMagnifierOptions()
	base.CenterX, base.CenterY, base.Radius = 16.25, 16.75, 11.3
	base.Zoom, base.Falloff, base.Feather = 2.1, 1.5, 0
	crop := image.Rect(13, 14, 32, 34)
	cases := []composite.MagnifierOptions{base, base, base, base, base, base, base, base}
	cases[1].Crop = &crop
	cases[2].Opacity = .4
	cases[3].Feather = 4
	cases[4].Filter, cases[4].Crop = ebiten.FilterLinear, &crop
	cases[5].Opacity = 0
	cases[6].CenterX, cases[6].CenterY = 1, 2
	cases[7].Falloff = 0
	pixels := make([]byte, size*size*4)
	background := color.RGBA{3, 5, 7, 255}
	for test, options := range cases {
		dst.Fill(background)
		lens.Draw(dst, source, options)
		dst.ReadPixels(pixels)
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				want := magnifierExpectedPixel(cpu, source.Bounds(), options, background, float64(x)+.5, float64(y)+.5)
				for channel, value := range want {
					got := float64(pixels[(y*size+x)*4+channel])
					if math.Abs(got-value) > 2 {
						return fmt.Errorf("magnifier case %d pixel (%d,%d) channel %d: got %.0f, want %.3f", test, x, y, channel, got, value)
					}
				}
			}
		}
	}
	fmt.Println("OK magnifier rendering: live texture, atlas/crop, falloff, feather, opacity, linear filtering, clipping")
	return nil
}

func magnifierExpectedPixel(source *image.RGBA, bounds image.Rectangle, o composite.MagnifierOptions, background color.RGBA, x, y float64) [4]float64 {
	out := [4]float64{float64(background.R), float64(background.G), float64(background.B), float64(background.A)}
	dx, dy := x-o.CenterX, y-o.CenterY
	r := math.Hypot(dx, dy) / o.Radius
	if r >= 1 || o.Opacity <= 0 {
		return out
	}
	weight := 1.0
	if o.Falloff > 0 {
		weight = math.Pow(1-r, o.Falloff)
	}
	z := 1 + (o.Zoom-1)*weight
	sx, sy := float64(bounds.Min.X)+o.CenterX+dx/z, float64(bounds.Min.Y)+o.CenterY+dy/z
	crop := bounds
	if o.Crop != nil {
		crop = crop.Intersect(*o.Crop)
	}
	if sx < float64(crop.Min.X) || sy < float64(crop.Min.Y) || sx >= float64(crop.Max.X) || sy >= float64(crop.Max.Y) {
		return out
	}
	sample := func(x, y float64) [4]float64 {
		ix := max(crop.Min.X, min(crop.Max.X-1, int(math.Floor(x))))
		iy := max(crop.Min.Y, min(crop.Max.Y-1, int(math.Floor(y))))
		c := source.RGBAAt(ix, iy)
		return [4]float64{float64(c.R), float64(c.G), float64(c.B), float64(c.A)}
	}
	c := sample(sx, sy)
	if o.Filter == ebiten.FilterLinear {
		bx, by := math.Floor(sx-.5)+.5, math.Floor(sy-.5)+.5
		fx, fy := sx-bx, sy-by
		a, b, d, e := sample(bx, by), sample(bx+1, by), sample(bx, by+1), sample(bx+1, by+1)
		for i := range c {
			c[i] = (a[i]*(1-fx)+b[i]*fx)*(1-fy) + (d[i]*(1-fx)+e[i]*fx)*fy
		}
	}
	alpha := o.Opacity
	if o.Feather > 0 {
		alpha *= min(1, (1-r)*o.Radius/o.Feather)
	}
	for i := range out {
		out[i] = c[i]*alpha + out[i]*(1-alpha)
	}
	return out
}
