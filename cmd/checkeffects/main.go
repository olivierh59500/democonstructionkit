// Command checkeffects renders every bitmap font with every shared scroll mode.
// It checks visible output and repeated-draw stability on the native renderer.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/olivierh59500/democonstructionkit/assets"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/presets"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/democonstructionkit/scrolltext"
)

const cellWidth, cellHeight = 192, 164

type result struct {
	Font, Mode    string
	VisiblePixels int
}
type check struct {
	scrolls []*scrolling.Scrolling
	names   []string
	tile    *ebiten.Image
	results []result
	done    bool
	err     error
}

func (c *check) Update() error { return c.err }
func (c *check) Layout(int, int) (int, int) {
	return len(c.names) * cellWidth, len(c.scrolls) * cellHeight
}
func (c *check) Draw(dst *ebiten.Image) {
	if c.done {
		return
	}
	c.done = true
	dst.Fill(color.RGBA{16, 16, 28, 255})
	a, b := make([]byte, cellWidth*(cellHeight-16)*4), make([]byte, cellWidth*(cellHeight-16)*4)
	for row, s := range c.scrolls {
		for column, name := range c.names {
			state := scrolling.IdentityState()
			state.X = 12
			state.Y = 65
			state.Time = 1.37
			state.Shape = name
			c.tile.Clear()
			s.DrawAt(c.tile, state)
			c.tile.ReadPixels(a)
			c.tile.Clear()
			s.DrawAt(c.tile, state)
			c.tile.ReadPixels(b)
			if !bytes.Equal(a, b) {
				c.err = fmt.Errorf("%s/%s: repeated Draw changed output", presets.Fonts()[row].ID, name)
				return
			}
			visible := 0
			for i := 3; i < len(a); i += 4 {
				if a[i] > 0 {
					visible++
				}
			}
			if visible == 0 {
				c.err = fmt.Errorf("%s/%s: no visible glyph pixels", presets.Fonts()[row].ID, name)
				return
			}
			c.results = append(c.results, result{presets.Fonts()[row].ID, name, visible})
			op := ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(column*cellWidth), float64(row*cellHeight))
			dst.DrawImage(c.tile, &op)
			ebitenutil.DebugPrintAt(dst, presets.Fonts()[row].ID+" / "+name, column*cellWidth, row*cellHeight+cellHeight-14)
		}
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	demos := flag.String("demos", "../../demos", "original demo assets")
	out := flag.String("out", "captures/effects", "capture and report directory")
	flag.Parse()
	store := assets.New(os.DirFS(*demos))
	defer store.Close()
	dna, err := scrolling.NewDNA(scrolling.DNAConfig{RotationSpeed: 20, Twist: motion.Wave{Amplitude: 15, Spatial: .025}})
	if err != nil {
		return err
	}
	defer dna.Close()
	projection, err := scrolling.Perspective(scrolling.PerspectiveConfig{Focal: 250, Near: 1, Depth: 50, CenterX: 96, CenterY: 35, DepthWave: motion.Wave{Amplitude: 40, Spatial: .05, Speed: 2}})
	if err != nil {
		return err
	}
	modes := map[string]scrolling.Mode{"normal": scrolling.Normal(), "bounce": scrolling.Bounce(motion.Wave{Amplitude: 12, Speed: 2}), "sine": scrolling.Sine(motion.Wave{Amplitude: 12, Spatial: .04, Speed: 2}), "zoom": scrolling.Zoom(scrolling.ZoomConfig{BaseX: 1, BaseY: 1, Wave: motion.Wave{Amplitude: .2, Speed: 2}, PivotX: 96, PivotY: 35}), "3d": projection, "dna": dna.Mode()}
	names := []string{"normal", "bounce", "sine", "zoom", "3d", "dna"}
	for i, f := range presets.TCBScrollForms() {
		name := fmt.Sprintf("tcb%d", i)
		names = append(names, name)
		modes[name] = f.Mode(scrolling.PlaneProjection{Focal: 250, Depth: 150, CenterX: 96, CenterY: 72})
	}
	c := &check{names: names, tile: ebiten.NewImage(cellWidth, cellHeight-16)}
	defer c.tile.Deallocate()
	for _, spec := range presets.Fonts() {
		img, err := store.Texture(spec.Path)
		if err != nil {
			return err
		}
		metrics, err := spec.Build(img.Bounds())
		if err != nil {
			return err
		}
		scale := math.Min(35/metrics.LineHeight(), .65)
		face := scrolling.Face{Atlas: img, Metrics: metrics, ScaleX: scale, ScaleY: scale}
		// Mix two independently scaled faces in each test, including the DNA mode.
		other := face
		other.ScaleX *= .6
		other.ScaleY *= .6
		s, err := scrolling.New(scrolling.Config{Text: "AB{font:small}C", Controls: scrolltext.Braces, Fonts: map[string]scrolling.Face{"default": face, "small": other}, Modes: modes})
		if err != nil {
			return fmt.Errorf("%s: %w", spec.ID, err)
		}
		c.scrolls = append(c.scrolls, s)
	}
	width, height := c.Layout(0, 0)
	err = capture.Run(capture.Config{Directory: *out, Frames: []int{0}, Width: width, Height: height}, func() (ebiten.Game, error) { return c, nil })
	if err != nil {
		return err
	}
	if c.err != nil {
		return c.err
	}
	b, err := json.MarshalIndent(c.results, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(*out, "report.json"), append(b, '\n'), 0644); err != nil {
		return err
	}
	fmt.Printf("Verified %d font/mode combinations with mixed-size glyphs; repeated Draw is stable.\n", len(c.results))
	return nil
}
