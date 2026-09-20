// Command scrollmodes applies the same independent effects to any audited font.
package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/assets"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/presets"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/democonstructionkit/scrolltext"
)

type demo struct {
	scroll *scrolling.Scrolling
	dna    *scrolling.DNA
	tick   uint64
	mode   string
	faces  string
}

func (d *demo) Update() error {
	d.tick++
	return d.scroll.Update(kit.Frame{Tick: d.tick, Time: float64(d.tick) / 60, Delta: 1.0 / 60})
}
func (d *demo) Draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{12, 14, 30, 255})
	d.scroll.Draw(dst)
	state := d.scroll.StateAt(float64(d.tick) / 60)
	ebitenutil.DebugPrintAt(dst, "DCK / "+state.Shape+"\n"+d.faces, 20, 20)
}
func (d *demo) Layout(int, int) (int, int) { return 800, 480 }

func main() {
	demos := flag.String("demos", "../../demos", "original demo assets directory")
	first := flag.String("font", "phenomena-dna-scroll-intro", "first font preset")
	second := flag.String("font2", "bilizir-demo", "second font preset")
	mode := flag.String("mode", "sequence", "normal, bounce, sine, zoom, 3d, dna or sequence")
	controls := flag.Bool("controls", false, "select effects with text commands instead of seconds")
	flag.Parse()
	store := assets.New(os.DirFS(*demos))
	defer store.Close()
	load := func(id string) scrolling.Face {
		spec, ok := presets.FindFont(id)
		if !ok {
			log.Fatalf("unknown font %q", id)
		}
		img, err := store.Texture(spec.Path)
		if err != nil {
			log.Fatal(err)
		}
		metrics, err := spec.Build(img.Bounds())
		if err != nil {
			log.Fatal(err)
		}
		return scrolling.Face{Atlas: img, Metrics: metrics, ScaleX: 2, ScaleY: 2}
	}
	faces := map[string]scrolling.Face{"default": load(*first), "alternate": load(*second)}
	dna, err := scrolling.NewDNA(scrolling.DNAConfig{RotationSpeed: 20, SliceWidth: 2, Twist: motion.Wave{Amplitude: 15, Spatial: .025}, Baseline: motion.Wave{Amplitude: 12, Spatial: .025, Speed: 1.2}})
	if err != nil {
		log.Fatal(err)
	}
	defer dna.Close()
	forms := presets.TCBScrollForms()
	perspective := forms[5]
	perspective.DepthSpeed *= 1.2
	perspective.VerticalSpeed *= 1.2
	modes := map[string]scrolling.Mode{
		"normal": scrolling.Normal(),
		"bounce": scrolling.Bounce(motion.Wave{Amplitude: 65, Speed: 2.4}),
		"sine":   scrolling.Sine(motion.Wave{Amplitude: 45, Spatial: .015, Speed: 2.4}),
		"zoom":   scrolling.Zoom(scrolling.ZoomConfig{BaseX: 1.2, BaseY: 1.2, Wave: motion.Wave{Amplitude: .5, Speed: 2}, PivotX: 400, PivotY: 200}),
		"3d":     perspective.Mode(scrolling.PlaneProjection{Focal: 250, Depth: 150, CenterX: 400, CenterY: 200}),
		"dna":    dna.Mode(),
	}
	c := scrolling.Config{Text: "ONE SCROLL WITH SHARED EFFECTS     {font:alternate}ANOTHER FONT WITH ITS OWN METRICS     {font:default}CHOOSE ANY EFFECT ORDER     ", Fonts: faces, Controls: scrolltext.Braces, Speed: 150, Repeat: true, Gap: 200, X: 800, Y: 190, Modes: modes, Shape: *mode}
	if *controls {
		c.Shape = "normal"
		c.Text = "NORMAL     {shape:bounce}BOUNCE     {font:alternate}{shape:sine}SINE     {shape:zoom}ZOOM     {font:default}{shape:3d}THREE D     {shape:dna}DNA SCROLL     "
	} else if *mode == "sequence" {
		c.Shape = "normal"
		c.Sequence, err = scrolling.NewModeSequence([]scrolling.Cue{{At: 0, Mode: "normal"}, {At: 5, Mode: "bounce"}, {At: 10, Mode: "sine"}, {At: 15, Mode: "zoom"}, {At: 20, Mode: "3d"}, {At: 25, Mode: "dna"}}, 30)
		if err != nil {
			log.Fatal(err)
		}
	} else if _, ok := modes[*mode]; !ok {
		log.Fatal(fmt.Errorf("unknown mode %q", *mode))
	}
	scroll, err := scrolling.New(c)
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowSize(1000, 600)
	ebiten.SetWindowTitle("DCK scroll modes and mixed fonts")
	if err = ebiten.RunGame(&demo{scroll: scroll, dna: dna, faces: *first + " + " + *second}); err != nil {
		log.Fatal(err)
	}
}
