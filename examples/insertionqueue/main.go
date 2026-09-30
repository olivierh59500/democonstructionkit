// Command insertionqueue combines controlled mixed-font insertion and recycled sprites.
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
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
	"github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/democonstructionkit/scrolltext"
	"github.com/olivierh59500/democonstructionkit/sprites"
	xfont "golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const width, height = 640, 400

type particle struct {
	phase float64
	lane  int
}
type ballPose struct{ x, y, size float64 }
type game struct {
	text   *scrolling.Scrolling
	queue  *motion.RecycledQueue[particle, ballPose]
	balls  *sprites.ImageSlots
	images []*ebiten.Image
	tick   int
}

// face generates an atlas once, with a configurable cell size and character order.
func face(scale int, reverse bool, paint color.NRGBA) (scrolling.Face, error) {
	order := make([]rune, 64)
	for i := range order {
		order[i] = rune(32 + i)
		if reverse {
			order[i] = rune(95 - i)
		}
	}
	base := image.NewNRGBA(image.Rect(0, 0, 16*7, 4*13))
	drawer := xfont.Drawer{Dst: base, Src: image.NewUniform(paint), Face: basicfont.Face7x13}
	for i, r := range order {
		drawer.Dot = fixed.P(i%16*7, i/16*13+11)
		drawer.DrawString(string(r))
	}
	bitmap := image.NewNRGBA(image.Rect(0, 0, base.Bounds().Dx()*scale, base.Bounds().Dy()*scale))
	for y := 0; y < bitmap.Bounds().Dy(); y++ {
		for x := 0; x < bitmap.Bounds().Dx(); x++ {
			bitmap.SetNRGBA(x, y, base.NRGBAAt(x/scale, y/scale))
		}
	}
	metrics, err := font.NewGrid(font.Grid{Bounds: bitmap.Bounds(), Cell: image.Pt(7*scale, 13*scale), Columns: 16, Order: string(order), Blanks: " ", Fallback: '?'})
	if err != nil {
		return scrolling.Face{}, err
	}
	return scrolling.Face{Atlas: ebiten.NewImageFromImage(bitmap), Metrics: metrics}, nil
}

func newGame() (_ *game, err error) {
	g := &game{}
	defer func() {
		if err != nil {
			g.Close()
		}
	}()
	fonts := make(map[string]scrolling.Face, 2)
	for i, name := range []string{"small", "large"} {
		paint := color.NRGBA{R: 85, G: 215, B: 255, A: 255}
		if i == 1 {
			paint = color.NRGBA{R: 255, G: 195, B: 85, A: 255}
		}
		f, e := face(2+i, i == 1, paint)
		if e != nil {
			return nil, e
		}
		fonts[name] = f
		g.images = append(g.images, f.Atlas)
	}
	var tokens []scrolltext.InsertionToken
	appendText := func(text, name string) {
		for _, r := range text {
			metric, _ := fonts[name].Metrics.Glyph(r)
			tokens = append(tokens, scrolltext.InsertionToken{Glyph: true, Rune: r, Font: name, Advance: int(metric.Advance)})
		}
	}
	appendText("DCK INSERTS EACH PIXEL   ", "small")
	tokens = append(tokens, scrolltext.InsertionToken{Command: 's', Payload: []byte{8}})
	appendText("ANOTHER FONT AND SPEED   ", "large")
	tokens = append(tokens, scrolltext.InsertionToken{Command: 'p', Payload: []byte{35}}, scrolltext.InsertionToken{Command: 's', Payload: []byte{4}})
	appendText("TEXT CAN PAUSE WHILE SPRITES KEEP MOVING   ", "small")
	// Drain the visible message before resetting this finite program. Ordinary
	// seamless repeats can instead use scrolling.Config.Repeat on the regular transport.
	for range 48 {
		appendText(" ", "small")
	}
	g.text, err = scrolling.New(scrolling.Config{Insertion: &scrolling.InsertionConfig{
		Fonts: fonts, Y: 325,
		Program: scrolltext.InsertionProgramConfig{Tokens: tokens, Speed: 4, TargetSpeed: 4, Entry: width, RetireBefore: -64,
			OnCommand: func(p *scrolltext.InsertionProgram, t scrolltext.InsertionToken) error {
				if t.Command == 'p' {
					return p.SetPause(int(t.Payload[0]))
				}
				return p.SetTargetSpeed(int(t.Payload[0]))
			}},
	}})
	if err != nil {
		return nil, err
	}
	bitmap := image.NewNRGBA(image.Rect(0, 0, 24, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			r := math.Hypot(float64(x)-11.5, float64(y)-11.5) / 11.5
			if r <= 1 {
				v := byte(90 + 165*math.Sqrt(1-r*r))
				bitmap.SetNRGBA(x, y, color.NRGBA{R: v, G: v / 2, B: 230, A: 255})
			}
		}
	}
	ball := ebiten.NewImageFromImage(bitmap)
	g.images = append(g.images, ball)
	g.queue, err = motion.NewRecycledQueue(motion.RecycledQueueConfig[particle, ballPose]{Items: make([]particle, 12), Count: 1, Step: 12, Spacing: 100, DepthWrap: 1200, Grow: true,
		Recycle: func(tick int, p *particle) { p.phase, p.lane = 0, tick%7-3 },
		Project: func(_ int, depth int, p *particle) (ballPose, error) {
			p.phase += .055
			scale := 440 / float64(180+depth)
			return ballPose{x: 320 + float64(p.lane)*48*scale, y: 170 + math.Sin(p.phase)*38*scale, size: 24 * scale}, nil
		},
	})
	if err != nil {
		return nil, err
	}
	g.balls, err = sprites.NewImageSlots(sprites.ImageSlotsConfig{Images: []*ebiten.Image{ball}, MaxSlots: 12,
		Select: func(_ kit.Frame, slots []sprites.ImageSlot) (int, error) {
			poses := g.queue.Poses()
			// The camera makes larger depth farther away: paint those slots first.
			for i := range poses {
				p := poses[len(poses)-1-i]
				slots[i] = sprites.ImageSlot{X: p.x - p.size/2, Y: p.y - p.size/2, Width: p.size, Height: p.size}
			}
			return len(poses), nil
		},
	})
	return g, err
}

func (g *game) Update() error {
	g.tick++
	f := kit.Frame{Tick: uint64(g.tick), Time: float64(g.tick) / 50, Delta: 1.0 / 50}
	if g.text.InsertionController().State().Finished {
		g.text.InsertionController().Reset()
	}
	if err := g.text.Update(f); err != nil {
		return err
	}
	if err := g.queue.Step(g.tick); err != nil {
		return err
	}
	return g.balls.Update(f)
}
func (g *game) Draw(dst *ebiten.Image) {
	dst.Fill(color.NRGBA{R: 6, G: 12, B: 24, A: 255})
	g.balls.Draw(dst)
	g.text.Draw(dst)
	ebitenutil.DebugPrintAt(dst, "DCK / MIXED FONT INSERTION + RECYCLED SPRITE QUEUE", 16, 18)
}
func (*game) Layout(int, int) (int, int) { return width, height }
func (g *game) Close() {
	if g.text != nil {
		g.text.Close()
	}
	if g.balls != nil {
		g.balls.Close()
	}
	for _, img := range g.images {
		img.Deallocate()
	}
}
func main() {
	directory := flag.String("capture", "", "write native PNG frames")
	first := flag.Int("frame", 120, "first update tick")
	count := flag.Int("frames", 1, "consecutive frames (1..1500)")
	flag.Parse()
	if *directory != "" {
		if *first < 0 || *count < 1 || *count > 1500 || *first > int(^uint(0)>>1)-*count {
			log.Fatal("invalid capture range")
		}
		frames := make([]int, *count)
		for i := range frames {
			frames[i] = *first + i
		}
		var g *game
		err := capture.Run(capture.Config{Directory: *directory, Frames: frames, Width: width, Height: height}, func() (ebiten.Game, error) { var e error; g, e = newGame(); return g, e })
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
	ebiten.SetWindowTitle("DCK Insertion and Recycled Sprites")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
