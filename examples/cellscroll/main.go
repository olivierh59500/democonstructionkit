// Command cellscroll composes flat and cuboid fonts through scrolling.New.
package main

import (
	"flag"
	"image"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
	bitmap "github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	textfont "golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const width, height = 640, 360

type game struct {
	atlas *ebiten.Image
	lanes [2]*scrolling.Scrolling
	tick  int
}

func newGame() (*game, error) {
	var chars []rune
	for r := 32; r < 127; r++ {
		chars = append(chars, rune(r))
	}
	pixels := image.NewNRGBA(image.Rect(0, 0, 16*8, 6*16))
	drawer := textfont.Drawer{Dst: pixels, Src: image.NewUniform(color.White), Face: basicfont.Face7x13}
	for i, r := range chars {
		drawer.Dot = fixed.P((i%16)*8, (i/16)*16+13)
		drawer.DrawString(string(r))
	}
	metrics, err := bitmap.NewGrid(bitmap.Grid{Bounds: pixels.Bounds(), Cell: image.Pt(8, 16), Columns: 16, Order: string(chars), Blanks: " ", Advance: 8, LineHeight: 16})
	if err != nil {
		return nil, err
	}
	cells, err := bitmap.NewImageCellBank(pixels, metrics, string(chars), 0)
	if err != nil {
		return nil, err
	}
	g := &game{atlas: ebiten.NewImageFromImage(pixels)}
	face := scrolling.Face{Atlas: g.atlas, Metrics: metrics, ScaleX: 4, ScaleY: 4}
	flat := scrolling.CellPainterConfig{Fonts: map[string]*bitmap.CellBank{"main": cells},
		Flat: scrolling.FlatCellConfig{Size: geometry.Vec2{X: .85, Y: .85}},
		Rows: func(_ string, row int, t float64) geometry.Vec3 {
			return geometry.Vec3{X: 12 * math.Sin(t*2+float64(row)*.25), Y: 8 * math.Sin(t*3+float64(row)*.3)}
		},
	}
	palette := []color.NRGBA{{255, 255, 255, 255}, {255, 255, 255, 255}, {255, 255, 255, 255}, {255, 255, 255, 255},
		{255, 90, 80, 255}, {255, 220, 40, 255}, {40, 90, 255, 255}, {40, 255, 255, 255}}
	cuboid := scrolling.CellPainterConfig{Fonts: map[string]*bitmap.CellBank{"main": cells}, Shape: scrolling.CellCuboid,
		Cuboid: scrolling.CuboidCellConfig{Size: geometry.Vec3{X: 3, Y: 3, Z: 3}, Origin: geometry.Vec3{X: -320, Z: 400},
			Camera: geometry.Camera{Center: geometry.Vec2{X: 320, Y: 260}, Focal: 400, Near: 1}, Colors: palette, IgnoreMappedY: true, FlipY: true},
		Rows: func(_ string, row int, t float64) geometry.Vec3 {
			return geometry.Vec3{Y: 40 - float64(row)*4 + 10*math.Sin(t*1.2+float64(row)*.25), Z: 35 * math.Sin(t+float64(row)*.35)}
		},
		Pose: func(sample scrolling.CellSample, pose scrolling.CellPose) scrolling.CellPose {
			pose.Rotation.Y = .2 * math.Sin(sample.Time+float64(sample.Cell.X)*.3)
			return pose
		},
	}
	for i, c := range []scrolling.CellPainterConfig{flat, cuboid} {
		text, y := "ONE SCROLLING PIPELINE - BITMAP CELLS - WAVES - ROTATION - REPEAT ", 70.0
		if i == 1 {
			text = "CUBOID CELLS - ANY FONT - PER VERTEX COLORS - DEPTH AND CUSTOM PATHS "
			y = 0
		}
		g.lanes[i], err = scrolling.New(scrolling.Config{Text: text, Fonts: map[string]scrolling.Face{"main": face}, Font: "main",
			X: width, Y: y, Speed: 100, Repeat: true, Gap: 40, Shape: "cells", Modes: map[string]scrolling.Mode{"cells": {Cells: &c}}})
		if err != nil {
			g.Close()
			return nil, err
		}
	}
	return g, nil
}

func (g *game) Update() error {
	g.tick++
	f := kit.Frame{Tick: uint64(g.tick), Time: float64(g.tick) / 60, Delta: 1.0 / 60}
	for _, lane := range g.lanes {
		if err := lane.Update(f); err != nil {
			return err
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		for _, lane := range g.lanes {
			lane.CellPainterController("cells").SetOutlined(true)
		}
	} else {
		for _, lane := range g.lanes {
			lane.CellPainterController("cells").SetOutlined(false)
		}
	}
	return nil
}
func (g *game) Draw(dst *ebiten.Image) {
	dst.Fill(color.NRGBA{R: 8, G: 16, B: 30, A: 255})
	for _, lane := range g.lanes {
		lane.Draw(dst)
	}
}
func (*game) Layout(int, int) (int, int) { return width, height }
func (g *game) Close() {
	for _, lane := range g.lanes {
		if lane != nil {
			lane.Close()
		}
	}
	if g.atlas != nil {
		g.atlas.Deallocate()
	}
}

func main() {
	dir := flag.String("capture", "", "write a deterministic PNG")
	frame := flag.Int("frame", 360, "update tick to capture")
	flag.Parse()
	if *dir != "" {
		if *frame < 0 {
			log.Fatal("-frame must be nonnegative")
		}
		var g *game
		err := capture.Run(capture.Config{Directory: *dir, Frames: []int{*frame}, Width: width, Height: height}, func() (ebiten.Game, error) { var err error; g, err = newGame(); return g, err })
		if g != nil {
			g.Close()
		}
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	g, err := newGame()
	if err != nil {
		log.Fatal(err)
	}
	defer g.Close()
	ebiten.SetTPS(60)
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle("DCK Cell Scrolling - hold Space for wireframe")
	if err = ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
