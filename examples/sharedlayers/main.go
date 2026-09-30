// Command sharedlayers combines indexed artwork, sparse point planes and meters.
package main

import (
	"flag"
	"image"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/effects"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sound"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

const width, height = 640, 360

type game struct {
	art    *ebiten.Image
	bank   *effects.IndexedImageBank
	planes [2]*sprites.IndexedPointPlane
	points []motion.WrappedPoint
	offset [3]int16
	meter  *sound.YMPeriodMeter
	bars   *composite.GradientBars
	tick   int
}

func newGame() (_ *game, err error) {
	g := &game{points: make([]motion.WrappedPoint, 320)}
	defer func() {
		if err != nil {
			g.Close()
		}
	}()
	bitmap := image.NewNRGBA(image.Rect(0, 0, 64, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 64; x++ {
			index := byte((x/8 + y/5) % 8)
			bitmap.SetNRGBA(x, y, color.NRGBA{R: index * 17, A: 255})
		}
	}
	g.art = ebiten.NewImageFromImage(bitmap)
	palettes := make([]effects.IndexedImagePalette, 2)
	for p := range palettes {
		palettes[p].Colors = make([]color.NRGBA, 8)
		for i := range 8 {
			palettes[p].Colors[i] = color.NRGBA{R: byte(12 + i*7 + p*8), G: byte(18 + i*9), B: byte(30 + i*15), A: 255}
		}
	}
	slots := make([]effects.IndexedImageSlot, 2)
	for i := range slots {
		slots[i].Palette = i
		slots[i].Options.GeoM.Scale(4.5, 5.5)
		slots[i].Options.GeoM.Translate(float64(16+i*320), 45)
	}
	g.bank, err = effects.NewIndexedImageBank(effects.IndexedImageBankConfig{Images: []*ebiten.Image{g.art}, Palettes: palettes, Slots: slots, Channel: composite.BitplaneRed, Scale: 15})
	if err != nil {
		return nil, err
	}
	for i := range g.points {
		seed := i
		if i%5 == 0 {
			seed = max(0, i-1)
		}
		g.points[i] = motion.WrappedPoint{X: int16(seed * 79), Y: int16(seed * 131), Z: int16(seed * 31)}
	}
	projection, err := motion.NewWrappedPointProjection(motion.WrappedPointProjectionConfig{Mask: [3]uint16{511, 511, 2047}, Bias: [2]int16{-256, -256}, Numerator: 160000, DepthBias: 180, Shift: 9, Center: image.Pt(144, 108), Bounds: image.Rect(0, 0, 288, 220)})
	if err != nil {
		return nil, err
	}
	colors := []color.NRGBA{{}, {R: 255, G: 240, B: 150, A: 255}, {R: 80, G: 190, B: 255, A: 255}, {R: 255, G: 120, B: 200, A: 255}}
	for i := range g.planes {
		collision := sprites.PointPlaneOR
		if i == 1 {
			collision = sprites.PointPlaneXOR
		}
		g.planes[i], err = sprites.NewIndexedPointPlane(sprites.IndexedPointPlaneConfig{Width: 288, Height: 220, Count: len(g.points), Collision: collision, Offset: image.Pt(16+i*320, 45), Palette: colors,
			Sample: func(index int) (sprites.PointPlaneSample, bool) {
				p, visible := projection.Project(g.points[index], g.offset)
				mask := byte(2)
				if p.Depth < 900 {
					mask = 1
				}
				return sprites.PointPlaneSample{X: p.X, Y: p.Y, Mask: mask}, visible
			},
		})
		if err != nil {
			return nil, err
		}
	}
	var levels [16]float64
	for i := range levels {
		levels[i] = float64(i) * 5
	}
	g.meter, err = sound.NewYMPeriodMeter(sound.YMPeriodMeterConfig{Columns: 72, Levels: levels, Gain: 1, Decay: 1.5, PeriodMin: 1, PeriodMaxExclusive: 4096, FrequencyScale: 8000, Envelope: true, EnvelopeShift: 4, EnvelopeLevel: 15, EnvelopeGain: .7, Gating: sound.YMPeriodGateAll})
	if err != nil {
		return nil, err
	}
	g.bars, err = composite.NewGradientBars(composite.GradientBarsConfig{Columns: 72, X: 32, Baseline: 335, Step: 8, Width: 6, Level: g.meter.Level, Top: color.NRGBA{R: 255, G: 180, B: 90, A: 255}, Bottom: color.NRGBA{R: 60, G: 190, B: 210, A: 255}})
	return g, err
}

func (g *game) Update() error {
	g.tick++
	seconds := float64(g.tick) / 50
	g.offset = [3]int16{int16(g.tick * 3), int16(g.tick * 2), int16(-g.tick * 8)}
	frame := kit.Frame{Tick: uint64(g.tick), Time: seconds, Delta: 1.0 / 50}
	for _, plane := range g.planes {
		if err := plane.Update(frame); err != nil {
			return err
		}
	}
	var registers [14]uint8
	for i := range 3 {
		period := int(130 + 90*math.Sin(seconds*(.7+float64(i)*.3)+float64(i)*2))
		registers[i*2], registers[i*2+1] = uint8(period), uint8(period>>8)
		registers[8+i] = uint8(10 + i*2)
	}
	return g.meter.Step(registers)
}
func (g *game) Draw(dst *ebiten.Image) {
	dst.Fill(color.NRGBA{R: 6, G: 12, B: 24, A: 255})
	g.bank.Draw(dst)
	for _, plane := range g.planes {
		plane.Draw(dst)
	}
	g.bars.Draw(dst)
	ebitenutil.DebugPrintAt(dst, "INDEXED ARTWORK + PROJECTED POINTS: OR / XOR", 16, 15)
	ebitenutil.DebugPrintAt(dst, "SHARED PEAKS / GRADIENT BARS - SYNTHETIC YM REGISTER SIGNALS", 16, 345)
}
func (*game) Layout(int, int) (int, int) { return width, height }
func (g *game) Close() {
	if g.bank != nil {
		g.bank.Close()
	}
	for _, p := range g.planes {
		if p != nil {
			p.Close()
		}
	}
	if g.bars != nil {
		g.bars.Close()
	}
	if g.art != nil {
		g.art.Deallocate()
	}
}
func main() {
	directory := flag.String("capture", "", "write native PNG frames")
	frame := flag.Int("frame", 200, "first update tick to capture")
	count := flag.Int("frames", 1, "consecutive frames to capture (1..1500)")
	flag.Parse()
	if *directory != "" {
		if *frame < 0 || *count < 1 || *count > 1500 || *frame > int(^uint(0)>>1)-*count {
			log.Fatal("invalid capture interval")
		}
		frames := make([]int, *count)
		for i := range frames {
			frames[i] = *frame + i
		}
		var g *game
		err := capture.Run(capture.Config{Directory: *directory, Frames: frames, Width: width, Height: height}, func() (ebiten.Game, error) { var err error; g, err = newGame(); return g, err })
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
	ebiten.SetTPS(50)
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle("DCK Shared Layers")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
