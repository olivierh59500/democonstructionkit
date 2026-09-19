// A self-contained demo: no external image, font or music files are required.
package main

import (
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/effects"
	bitmap "github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/motion"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
func run() error {
	const order = " ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!?.-"
	sheet := image.NewRGBA(image.Rect(0, 0, 16*8, 3*16))
	drawer := font.Drawer{Dst: sheet, Src: image.NewUniform(color.White), Face: basicfont.Face7x13}
	for i, r := range order {
		drawer.Dot = fixed.P(i%16*8, i/16*16+13)
		drawer.DrawString(string(r))
	}
	atlas := ebiten.NewImageFromImage(sheet)
	defer atlas.Deallocate()
	metrics, err := bitmap.NewGrid(bitmap.Grid{Bounds: sheet.Bounds(), Cell: image.Pt(8, 16), Columns: 16, Order: order, Uppercase: true})
	if err != nil {
		return err
	}
	scroll, err := effects.NewScroller(atlas, metrics, effects.ScrollConfig{Width: 640, Height: 400, Message: "  DEMOCONSTRUCTIONKIT - SHARED EFFECTS IN GO!  ", Speed: 100, Gap: 40, Scale: 3, Y: 320, Wave: motion.Waves{{Amplitude: 18, Spatial: .02, Speed: 2}}})
	if err != nil {
		return err
	}
	stars, err := effects.NewStarfield(640, 400, 300, 42)
	if err != nil {
		return err
	}
	mesh, err := effects.NewMesh(effects.Cube(150, geometry.Vec2{X: 1, Y: 1}, color.NRGBA{R: 180, G: 120, B: 255, A: 180}), nil, geometry.Camera{Center: geometry.Vec2{X: 320, Y: 170}, Focal: 350, Near: 10})
	if err != nil {
		stars.Close()
		return err
	}
	mesh.Animate = func(t float64) effects.Transform {
		return effects.Transform{Position: geometry.Vec3{Z: 420}, Rotation: geometry.Vec3{X: t * .6, Y: t, Z: t * .2}, Scale: 1}
	}
	game, err := kit.NewGame(kit.Group{&effects.Solid{Color: color.NRGBA{5, 7, 20, 255}}, stars, mesh, scroll}, kit.Config{Width: 640, Height: 400, TPS: 60})
	if err != nil {
		stars.Close()
		mesh.Close()
		return err
	}
	defer game.Close()
	ebiten.SetWindowSize(960, 600)
	ebiten.SetWindowTitle("democonstructionkit")
	ebiten.SetTPS(60)
	return ebiten.RunGame(game)
}
