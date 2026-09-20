// Composer demonstrates one scrolling program with mixed fonts and optional
// controls, independently layered raster bands, logos and scanline deformation.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/assets"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/effects"
	"github.com/olivierh59500/democonstructionkit/presets"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/democonstructionkit/scrolltext"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	root := flag.String("demos", "../../demos", "original asset directory")
	frames := flag.Int("frames", 0, "optional update limit for smoke tests")
	rate := flag.Float64("rate", 1, "animation playback rate")
	flag.Parse()
	store := assets.New(os.DirFS(*root))
	defer store.Close()
	faces := map[string]scrolling.Face{}
	for alias, id := range map[string]string{"soap": "bilizir-demo", "chrome": "viva_tcb"} {
		spec, _ := presets.FindFont(id)
		atlas, err := store.Texture(spec.Path)
		if err != nil {
			return err
		}
		metrics, err := spec.Build(atlas.Bounds())
		if err != nil {
			return err
		}
		faces[alias] = scrolling.Face{Atlas: atlas, Metrics: metrics}
	}
	// All controls are opt-in. Set Controls:nil to render braces as literal text.
	program, err := scrolling.New(scrolling.Config{
		Fonts: faces, Font: "soap", Controls: scrolltext.Braces, Speed: 140, X: 928, Y: 12, Gap: 180, Repeat: true,
		Text:   "ONE SCROLLING! {font:chrome}TWO FONTS {speed:280}FASTER {pause:1.5}{speed:140}{shape:sine}SINE {font:soap}{shape:twist}{effect:gold}TWIST AND COLOR {effect:none}{shape:flat}BACK TO NORMAL! ",
		Shapes: map[string]scrolling.Mapper{"flat": nil, "sine": nil, "twist": nil},
		Effects: map[string]scrolling.Mapper{"gold": func(s scrolling.Sample, op *ebiten.DrawImageOptions) bool {
			op.ColorScale.Scale(1, .7, .2, 1)
			return true
		}},
	})
	if err != nil {
		return err
	}
	// Deformation consumes the program's active shape control. The same pass can
	// be applied to a logo or sprite surface by changing its Source.
	textLayer, err := composite.NewPass(program, 1056, 80, func(dst, source *ebiten.Image, f kit.Frame) {
		shape := program.StateAt(f.Time).Shape
		composite.Strips{Thickness: 1, Map: func(row int, r image.Rectangle, f kit.Frame) composite.Strip {
			x := 0.0
			y := float64(row) + 380
			switch shape {
			case "sine":
				x = 30 * math.Sin(float64(row)*.1+f.Time*2)
			case "twist":
				x = 50 * math.Sin(float64(row)*.2-f.Time*3)
				y += 35 * math.Sin(f.Time*1.4)
			}
			op := ebiten.DrawImageOptions{}
			op.GeoM.Translate(x-128, y)
			return composite.Strip{Source: r, Options: op}
		}}.Draw(dst, source, f)
	})
	if err != nil {
		return err
	}
	logo, err := store.Texture("bilizir-demo/assets/logo.png")
	if err != nil {
		textLayer.Close()
		return err
	}
	logos := &composite.Sprites{Count: 4, Sample: func(i int, f kit.Frame) composite.Instance {
		op := ebiten.DrawImageOptions{}
		op.GeoM.Scale(.35, .35)
		op.GeoM.Translate(300+220*math.Sin(f.Time+float64(i)*.8), 120+70*math.Cos(f.Time*.7+float64(i)))
		op.ColorScale.ScaleAlpha(.7)
		return composite.Instance{Image: logo, Options: op}
	}}
	raster, err := store.Texture("bilizir-demo/assets/bars.png")
	if err != nil {
		textLayer.Close()
		return err
	}
	rasters := &composite.Sprites{Count: 40, Sample: func(i int, f kit.Frame) composite.Instance {
		y := i % 20
		r := image.Rect(0, y, raster.Bounds().Dx(), y+1)
		op := ebiten.DrawImageOptions{}
		op.GeoM.Scale(800/float64(r.Dx()), 2)
		op.GeoM.Translate(0, 190+float64(i)*2+60*math.Sin(f.Time+float64(i)*.08))
		return composite.Instance{Image: raster, Source: &r, Options: op}
	}}
	// Reorder these layers freely, or wrap any one in composite.Layer for a
	// separate native surface, transform, opacity and blend mode.
	game, err := kit.NewGame(kit.Group{&effects.Solid{Color: color.NRGBA{8, 10, 25, 255}}, rasters, logos, textLayer}, kit.Config{Width: 800, Height: 600, TPS: 60})
	if err != nil {
		textLayer.Close()
		return err
	}
	defer game.Close()
	if err = game.Clock.SetSpeed(*rate); err != nil {
		return err
	}
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("DCK controlled scrolling and composition")
	ebiten.SetRunnableOnUnfocused(*frames > 0)
	return ebiten.RunGame(&application{Game: game, limit: *frames})
}

type application struct {
	*kit.Game
	count, limit int
	paused       bool
}

func (a *application) Update() error {
	if (a.limit > 0 && a.count >= a.limit) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		a.paused = !a.paused
		a.Clock.Pause(a.paused)
	}
	a.count++
	return a.Game.Update()
}
