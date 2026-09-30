// Command contourscroll composes mixed vector fonts, polar text and perspective text.
package main

import (
	"flag"
	"image/color"
	"log"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	kit "github.com/olivierh59500/democonstructionkit"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
	"github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/scrolling"
)

const width, height = 960, 540

type game struct {
	white *ebiten.Image
	lanes [3]*scrolling.Scrolling
	tick  int
}

func newGame() (_ *game, err error) {
	g := &game{white: ebiten.NewImage(1, 1)}
	g.white.Fill(color.White)
	defer func() {
		if err != nil {
			g.Close()
		}
	}()
	square, err := makeFont(false)
	if err != nil {
		return nil, err
	}
	beveled, err := makeFont(true)
	if err != nil {
		return nil, err
	}
	fonts := map[string]*font.ContourBank{"square": square, "beveled": beveled}
	base := scrolling.ContourPainterConfig{Fonts: fonts, Font: "square", Texture: g.white}
	var glyphs []scrolling.Glyph
	for i, character := range "DCK EFFECTS - MIXED FONTS - WAVES - ZOOM - " {
		name := "square"
		if i/6%2 == 1 {
			name = "beveled"
		}
		glyphs = append(glyphs, scrolling.Glyph{Rune: character, Font: name,
			Advance: 24, ScaleX: 4, ScaleY: 4})
	}
	zoomAndTint := func(s scrolling.Sample, op *ebiten.DrawImageOptions) bool {
		zoom := 1 + .18*math.Sin(s.Time*1.7+s.Glyph.Offset*.018)
		op.GeoM.Translate(-s.X, -s.Y)
		op.GeoM.Scale(zoom, zoom)
		op.GeoM.Translate(s.X, s.Y)
		if s.Glyph.Font == "beveled" {
			op.ColorScale.Scale(.35, 1, .85, 1)
		} else {
			op.ColorScale.Scale(1, .78, .3, 1)
		}
		return true
	}
	g.lanes[0], err = scrolling.New(scrolling.Config{
		Glyphs: glyphs, Speed: 110, Gap: 24, Y: 76, Repeat: true, Shape: "flat",
		Modes: map[string]scrolling.Mode{"flat": {
			Contours: &base,
			Map: scrolling.Chain(scrolling.Sine(motion.Wave{
				Amplitude: 14, Spatial: .022, Speed: 2,
			}).Map, zoomAndTint),
		}},
	})
	if err != nil {
		return nil, err
	}

	// A power-of-two table replaces per-point trigonometry. Font X becomes
	// angular position and font Y becomes radius; the source alphabet stays shared.
	wave := make([]int16, 1024)
	for i := range wave {
		wave[i] = int16(math.Round(32767 * math.Sin(float64(i)*2*math.Pi/1024)))
	}
	polar, err := motion.NewTablePolar(motion.TablePolarConfig{
		Wave: wave, Shift: 15, Phases: [2]int{256, 0}, Center: [2]int64{240, 355},
	})
	if err != nil {
		return nil, err
	}
	ringText := []rune("DCK EFFECTS     ")
	polarPaint := base
	polarPaint.Map = func(s scrolling.ContourSample, _ ebiten.GeoM) (geometry.Vec2, bool) {
		angle := -int64(math.Round(s.Time*65)) - int64(s.Index*64) - int64(math.Round(s.Point.X*7))
		radius := int64(math.Round(125 + 10*math.Sin(s.Time*1.3) - s.Point.Y*5))
		point, ok := polar.Point(angle, radius)
		return geometry.Vec2{X: point.X, Y: point.Y}, ok
	}
	ringColor := color.NRGBA{R: 255, G: 100, B: 155, A: 255}
	polarPaint.Color = &ringColor
	g.lanes[1], err = scrolling.New(scrolling.Config{
		GlyphWindow: &scrolling.GlyphWindowConfig{Count: len(ringText), Advance: 6,
			Glyph: func(slot int) scrolling.Glyph { return scrolling.Glyph{Rune: ringText[slot], Font: "beveled"} }},
		Shape: "polar", Modes: map[string]scrolling.Mode{"polar": {Contours: &polarPaint}},
	})
	if err != nil {
		return nil, err
	}

	// Change the affine coefficients to tilt the plane or choose another vanishing
	// point. Integer division belongs to the projection, not to the font artwork.
	projection, err := motion.NewRationalGrid(motion.RationalGridConfig{
		X: [3]int64{0, 0, 420}, Y: [3]int64{-5000, 900, 0},
		Denominator: [3]int64{380, 6, 1}, Center: [2]int64{740, 355},
	})
	if err != nil {
		return nil, err
	}
	planeText := []rune(" CONTOURS ")
	planePaint := base
	planePaint.Map = func(s scrolling.ContourSample, _ ebiten.GeoM) (geometry.Vec2, bool) {
		column := int64(math.Round((float64(s.Index*6)+s.Point.X-float64(len(planeText)*3))*4 - 25*math.Sin(s.Time*.8)))
		row := int64(math.Round(s.Point.Y * 8))
		point, ok := projection.Point(column, row)
		return geometry.Vec2{X: point.X, Y: point.Y + 18*math.Sin(s.Time*.6)}, ok
	}
	planeColor := color.NRGBA{R: 110, G: 175, B: 255, A: 255}
	planePaint.Color = &planeColor
	g.lanes[2], err = scrolling.New(scrolling.Config{
		GlyphWindow: &scrolling.GlyphWindowConfig{Count: len(planeText), Advance: 6,
			Glyph: func(slot int) scrolling.Glyph { return scrolling.Glyph{Rune: planeText[slot], Font: "square"} }},
		Shape: "perspective", Modes: map[string]scrolling.Mode{"perspective": {Contours: &planePaint}},
	})
	if err != nil {
		return nil, err
	}
	return g, nil
}

// Each lit cell becomes a polygon once. The second bank chamfers its corners;
// both banks have their own artwork and retain the same six-unit pen advance.
func makeFont(beveled bool) (*font.ContourBank, error) {
	patterns := map[rune]string{
		'A': "01110/10001/10001/11111/10001/10001/10001", 'C': "01111/10000/10000/10000/10000/10000/01111",
		'D': "11110/10001/10001/10001/10001/10001/11110", 'E': "11111/10000/10000/11110/10000/10000/11111",
		'F': "11111/10000/10000/11110/10000/10000/10000", 'I': "11111/00100/00100/00100/00100/00100/11111",
		'K': "10001/10010/10100/11000/10100/10010/10001", 'M': "10001/11011/10101/10101/10001/10001/10001",
		'N': "10001/11001/10101/10011/10001/10001/10001", 'O': "01110/10001/10001/10001/10001/10001/01110",
		'R': "11110/10001/10001/11110/10100/10010/10001", 'S': "01111/10000/10000/01110/00001/00001/11110",
		'T': "11111/00100/00100/00100/00100/00100/00100", 'U': "10001/10001/10001/10001/10001/10001/01110",
		'V': "10001/10001/10001/10001/10001/01010/00100", 'W': "10001/10001/10001/10101/10101/10101/01010",
		'X': "10001/10001/01010/00100/01010/10001/10001", 'Z': "11111/00001/00010/00100/01000/10000/11111",
		'-': "00000/00000/00000/11111/00000/00000/00000", ' ': "00000/00000/00000/00000/00000/00000/00000",
	}
	glyphs := make(map[rune]font.ContourGlyph, len(patterns))
	for character, pattern := range patterns {
		glyph := font.ContourGlyph{Advance: 6}
		for y, row := range strings.Split(pattern, "/") {
			for x, value := range row {
				if value != '1' {
					continue
				}
				left, top := float64(x), float64(y)
				var contour []geometry.Vec2
				if beveled {
					contour = []geometry.Vec2{
						{X: left + .22, Y: top}, {X: left + .78, Y: top},
						{X: left + 1, Y: top + .22}, {X: left + 1, Y: top + .78},
						{X: left + .78, Y: top + 1}, {X: left + .22, Y: top + 1},
						{X: left, Y: top + .78}, {X: left, Y: top + .22},
					}
				} else {
					contour = []geometry.Vec2{{X: left, Y: top}, {X: left + 1, Y: top},
						{X: left + 1, Y: top + 1}, {X: left, Y: top + 1}}
				}
				glyph.Contours = append(glyph.Contours, contour)
			}
		}
		glyphs[character] = glyph
	}
	return font.NewContourBank(font.ContourBankConfig{Glyphs: glyphs, Fallback: ' '})
}

func (g *game) Update() error {
	g.tick++
	f := kit.Frame{Tick: uint64(g.tick), Time: float64(g.tick) / 50, Delta: 1.0 / 50}
	for _, lane := range g.lanes {
		if err := lane.Update(f); err != nil {
			return err
		}
	}
	return nil
}

func (g *game) Draw(dst *ebiten.Image) {
	dst.Fill(color.NRGBA{R: 8, G: 16, B: 30, A: 255})
	for _, lane := range g.lanes {
		lane.Draw(dst)
		if err := lane.Err(); err != nil {
			panic(err)
		}
	}
	ebitenutil.DebugPrintAt(dst, "ONE SCROLLING PIPELINE: MIXED VECTOR FONTS + SINE + ZOOM + CONTINUOUS REPEAT", 18, 22)
	ebitenutil.DebugPrintAt(dst, "TABLE POLAR: ANGLE + RADIUS + PHASE", 18, 202)
	ebitenutil.DebugPrintAt(dst, "RATIONAL GRID: INTEGER PERSPECTIVE + CUSTOM PATH", width/2+18, 202)
	ebitenutil.DebugPrintAt(dst, "Font polygons and the wave table are prepared once; draw reuses DCK geometry buffers.", 18, height-30)
}

func (*game) Layout(int, int) (int, int) { return width, height }
func (g *game) Close() {
	for _, lane := range g.lanes {
		if lane != nil {
			lane.Close()
		}
	}
	if g.white != nil {
		g.white.Deallocate()
		g.white = nil
	}
}

func main() {
	dir := flag.String("capture", "", "write deterministic native PNG frames")
	frame := flag.Int("frame", 250, "update tick to capture")
	count := flag.Int("frames", 1, "consecutive capture frame count (1..1500)")
	flag.Parse()
	if *dir != "" {
		if *frame < 0 {
			log.Fatal("-frame must be nonnegative")
		}
		if *count < 1 || *count > 1500 || *frame > int(^uint(0)>>1)-(*count-1) {
			log.Fatal("-frames must be 1..1500 without overflowing the final capture tick")
		}
		frames := make([]int, *count)
		for i := range frames {
			frames[i] = *frame + i
		}
		var g *game
		err := capture.Run(capture.Config{Directory: *dir, Frames: frames, Width: width, Height: height}, func() (ebiten.Game, error) {
			var err error
			g, err = newGame()
			return g, err
		})
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
	ebiten.SetWindowTitle("DCK Contour Fonts: Waves, Polar and Perspective")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
