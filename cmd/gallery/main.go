// Command gallery runs the source-asset effect recipes and optional audio.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/assets"
	"github.com/olivierh59500/democonstructionkit/presets"
	"github.com/olivierh59500/democonstructionkit/recipes"
	"github.com/olivierh59500/democonstructionkit/sound"
	playback "github.com/olivierh59500/democonstructionkit/sound/ebiten"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	root := flag.String("demos", "../../demos", "path to the original demo repositories")
	name := flag.String("demo", "bilizir-demo", "recipe name, or all for a contact sheet")
	list := flag.Bool("list", false, "list available recipes")
	withAudio := flag.Bool("audio", false, "play the recipe's original music")
	musicPath := flag.String("music", "", "external YM/MOD/XM/S3M/IT file, overrides recipe music")
	capture := flag.String("capture", "", "write a PNG and exit")
	sampleTime := flag.Float64("time", 3, "time in seconds for PNG captures")
	frames := flag.Int("frames", 0, "exit after this many rendered frames; zero runs until Escape")
	check := flag.Bool("check", false, "verify every recipe using the real graphics backend")
	flag.Parse()
	if *list {
		for _, r := range recipes.Catalog() {
			fmt.Printf("%-31s %s\n", r.Name, r.Description)
		}
		return nil
	}
	if *sampleTime < 0 || *frames < 0 {
		return fmt.Errorf("time and frame count must be nonnegative")
	}
	store := assets.New(os.DirFS(*root))
	defer store.Close()
	if *check || *capture != "" || *frames > 0 {
		ebiten.SetRunnableOnUnfocused(true)
	}
	if *check {
		return checkRender(store)
	}
	var effect kit.Effect
	var err error
	width, height := recipes.Width, recipes.Height
	if *name == "all" {
		group := kit.Group{}
		width, height = 1600, 800
		for i, r := range recipes.Catalog() {
			child, buildErr := recipes.Build(r.Name, store)
			if buildErr != nil {
				group.Close()
				return buildErr
			}
			x, y := i%5*320, i/5*200
			v, buildErr := kit.NewViewport(child, recipes.Width, recipes.Height, image.Rect(x, y, x+320, y+200))
			if buildErr != nil {
				kit.Close(child)
				group.Close()
				return buildErr
			}
			group = append(group, v)
		}
		effect = group
	} else {
		effect, err = recipes.Build(*name, store)
		if err != nil {
			return err
		}
	}
	defer kit.Close(effect)
	app := &gallery{effect: effect, store: store, name: *name, width: width, height: height, audio: *withAudio, music: *musicPath, capture: *capture, sampleTime: *sampleTime, limit: *frames}
	defer func() {
		if app.player != nil {
			app.player.Close()
		}
	}()
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle("democonstructionkit / " + *name)
	ebiten.SetTPS(60)
	if err := ebiten.RunGame(app); err != nil {
		return err
	}
	return app.err
}

type gallery struct {
	effect                    kit.Effect
	store                     *assets.Store
	name                      string
	width, height             int
	audio                     bool
	music, capture            string
	sampleTime                float64
	time                      float64
	limit, drawn              int
	initialized, paused, done bool
	err                       error
	player                    *playback.Player
}

func (g *gallery) Update() error {
	if g.done || (g.limit > 0 && g.drawn >= g.limit) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if !g.initialized {
		g.initialized = true
		if g.audio || g.music != "" {
			if err := g.startAudio(); err != nil {
				return err
			}
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.paused = !g.paused
		if g.player != nil {
			if g.paused {
				g.player.Pause()
			} else {
				g.player.Play()
			}
		}
	}
	t := g.time
	if g.capture != "" {
		t = g.sampleTime
	}
	delta := 1.0 / 60
	if g.paused {
		delta = 0
	}
	if err := g.effect.Update(kit.Frame{Time: t, Delta: delta}); err != nil {
		return err
	}
	g.time += delta
	return nil
}
func (g *gallery) Draw(dst *ebiten.Image) {
	g.effect.Draw(dst)
	if g.name == "all" {
		for i, r := range recipes.Catalog() {
			ebitenutil.DebugPrintAt(dst, r.Name, i%5*320+5, i/5*200+5)
		}
	}
	g.drawn++
	if g.capture != "" && !g.done {
		g.err = savePNG(g.capture, dst)
		g.done = true
	}
}
func (g *gallery) Layout(int, int) (int, int) { return g.width, g.height }
func (g *gallery) startAudio() error {
	var data []byte
	var err error
	module := false
	if g.music != "" {
		data, err = os.ReadFile(g.music)
		module = !strings.EqualFold(filepath.Ext(g.music), ".ym")
	} else {
		r, ok := recipes.Find(g.name)
		if !ok {
			return fmt.Errorf("select one recipe or pass -music for audio")
		}
		data, err = fs.ReadFile(g.store.Files, r.Music)
		module = r.Module
		if err == nil && module {
			data, err = presets.RealityModule(data, 1)
		}
	}
	if err != nil {
		return err
	}
	var stream *sound.Stream
	if module {
		stream, err = sound.NewModule(data, sound.ModuleOptions{SampleRate: 48000, Interpolation: true})
	} else {
		stream, err = sound.NewYM(data, sound.YMOptions{SampleRate: 48000, Loop: true})
	}
	if err != nil {
		return err
	}
	ctx := audio.CurrentContext()
	if ctx == nil {
		ctx = audio.NewContext(48000)
	}
	g.player, err = playback.NewPlayer(ctx, stream)
	if err != nil {
		stream.Close()
		return err
	}
	if err = g.player.SetVolume(.5); err != nil {
		return err
	}
	g.player.Play()
	return nil
}
func savePNG(path string, src *ebiten.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	img := image.NewRGBA(src.Bounds())
	src.ReadPixels(img.Pix)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	encodeErr := png.Encode(f, img)
	closeErr := f.Close()
	if encodeErr != nil {
		return encodeErr
	}
	return closeErr
}

// The render check exercises actual GPU output, not just constructor compilation.
type checker struct {
	store             *assets.Store
	catalog           []recipes.Recipe
	index, sample     int
	effect            kit.Effect
	first             *ebiten.Image
	a, b              []byte
	err               error
	finished          bool
	prepared          bool
	primitivesChecked bool
	hashes            map[string][32]byte
}

var sampleTimes = []float64{0, 1.5, 7, 13, 19, 25, 31}

func checkRender(store *assets.Store) error {
	c := &checker{store: store, catalog: recipes.Catalog(), first: ebiten.NewImage(recipes.Width, recipes.Height), a: make([]byte, recipes.Width*recipes.Height*4), b: make([]byte, recipes.Width*recipes.Height*4), hashes: map[string][32]byte{}}
	defer c.first.Deallocate()
	defer func() { kit.Close(c.effect) }()
	ebiten.SetWindowSize(640, 400)
	ebiten.SetWindowTitle("democonstructionkit rendering checks")
	ebiten.SetTPS(60)
	if err := ebiten.RunGame(c); err != nil {
		return err
	}
	return c.err
}
func (c *checker) Layout(int, int) (int, int) { return recipes.Width, recipes.Height }
func (c *checker) Update() error {
	if c.finished {
		return ebiten.Termination
	}
	if c.effect == nil {
		e, err := recipes.Build(c.catalog[c.index].Name, c.store)
		if err != nil {
			return err
		}
		c.effect = e
	}
	err := c.effect.Update(kit.Frame{Time: sampleTimes[c.sample], Delta: 1.0 / 60})
	c.prepared = err == nil
	return err
}
func (c *checker) Draw(screen *ebiten.Image) {
	if c.finished || c.effect == nil || !c.prepared {
		return
	}
	c.prepared = false
	if !c.primitivesChecked {
		c.primitivesChecked = true
		if err := checkRenderingPrimitives(); err != nil {
			c.err = err
			c.finished = true
			return
		}
	}
	c.first.Clear()
	c.effect.Draw(c.first)
	c.first.ReadPixels(c.a)
	// Reuse the same target to avoid differences in GPU atlas placement.
	c.first.Clear()
	c.effect.Draw(c.first)
	c.first.ReadPixels(c.b)
	name := c.catalog[c.index].Name
	if !bytes.Equal(c.a, c.b) {
		different, maxDelta := 0, 0
		for i, a := range c.a {
			delta := int(a) - int(c.b[i])
			if delta < 0 {
				delta = -delta
			}
			if delta > 0 {
				different++
			}
			maxDelta = max(maxDelta, delta)
		}
		c.err = fmt.Errorf("%s: Draw changes output (%d bytes, maximum difference %d)", name, different, maxDelta)
		c.finished = true
		return
	}
	colors := map[uint32]struct{}{}
	for i := 0; i < len(c.a); i += 4 {
		colors[binary.LittleEndian.Uint32(c.a[i:])] = struct{}{}
		if len(colors) > 1 {
			break
		}
	}
	if len(colors) <= 1 {
		c.err = fmt.Errorf("%s: output has too few colors (%d)", name, len(colors))
		c.finished = true
		return
	}
	hash := sha256.Sum256(c.a)
	if c.sample == 0 {
		c.hashes[name] = hash
	} else if c.sample == 1 && c.hashes[name] == hash {
		c.err = fmt.Errorf("%s: output does not animate", name)
		c.finished = true
		return
	}
	screen.DrawImage(c.first, nil)
	c.sample++
	if c.sample == len(sampleTimes) {
		fmt.Printf("OK %-31s %d samples, repeatable Draw, animated output\n", name, len(sampleTimes))
		kit.Close(c.effect)
		c.effect = nil
		c.sample = 0
		c.index++
		if c.index == len(c.catalog) {
			c.finished = true
		}
	}
}
