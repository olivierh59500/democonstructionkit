//go:build dck_authoring_postcheck

package scene

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/authoring"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
)

type postCompositionCheck struct {
	full, noLens, noTitlePass *Scene
	err                       error
	checked                   bool
}

func (c *postCompositionCheck) Layout(int, int) (int, int) { return 640, 360 }
func (c *postCompositionCheck) Update() error              { return nil }
func (c *postCompositionCheck) Draw(*ebiten.Image) {
	if c.checked {
		return
	}
	c.checked = true
	render := func(scene *Scene, seconds float64) ([]byte, error) {
		if err := scene.Update(kit.Frame{Time: seconds, Tick: uint64(seconds * 60)}); err != nil {
			return nil, err
		}
		image := ebiten.NewImageWithOptions(image.Rect(0, 0, 640, 360), &ebiten.NewImageOptions{Unmanaged: true})
		defer image.Deallocate()
		scene.Draw(image)
		pixels := make([]byte, 640*360*4)
		image.ReadPixels(pixels)
		return pixels, nil
	}
	before, err := render(c.full, .5)
	if err != nil {
		c.err = err
		return
	}
	beforeWithoutLens, err := render(c.noLens, .5)
	if err != nil {
		c.err = err
		return
	}
	if !bytes.Equal(before, beforeWithoutLens) {
		c.err = fmt.Errorf("inactive scene lens changed pixels before its window")
		return
	}
	after, err := render(c.full, 1.5)
	if err != nil {
		c.err = err
		return
	}
	afterWithoutLens, err := render(c.noLens, 1.5)
	if err != nil {
		c.err = err
		return
	}
	withoutTitlePass, err := render(c.noTitlePass, 1.5)
	if err != nil {
		c.err = err
		return
	}
	lensPixels, outsidePixels, titlePixels := 0, 0, 0
	for y := 0; y < 360; y++ {
		for x := 0; x < 640; x++ {
			offset := (y*640 + x) * 4
			if !bytes.Equal(after[offset:offset+4], afterWithoutLens[offset:offset+4]) {
				if x >= 255 && x < 405 && y >= 150 && y < 295 {
					lensPixels++
				} else {
					outsidePixels++
				}
			}
			if y >= 40 && y < 190 && !bytes.Equal(afterWithoutLens[offset:offset+4], withoutTitlePass[offset:offset+4]) {
				titlePixels++
			}
		}
	}
	if lensPixels == 0 || outsidePixels != 0 || titlePixels == 0 {
		c.err = fmt.Errorf("unexpected composition coverage: lens=%d outside=%d title=%d", lensPixels, outsidePixels, titlePixels)
		return
	}
	repeated, err := render(c.full, 10.5)
	if err != nil {
		c.err = err
		return
	}
	repeatedWithoutLens, err := render(c.noLens, 10.5)
	if err != nil {
		c.err = err
		return
	}
	if bytes.Equal(repeated, repeatedWithoutLens) {
		c.err = fmt.Errorf("the magnifier did not return in its second window")
		return
	}
	half := .5
	project := authoring.Project{Version: 1, Units: authoring.DefaultUnits(),
		Canvas: authoring.Canvas{Width: 16, Height: 16, TPS: 60},
		Assets: map[string]string{"tile": "image"},
		Layers: []authoring.Layer{{ID: "tile", Kind: "background",
			Background: &authoring.Background{Image: "tile"}}},
	}
	tile := ebiten.NewImage(16, 16)
	defer tile.Deallocate()
	tile.Fill(color.NRGBA{R: 255, A: 128})
	assets := authoring.Assets{Images: map[string]*ebiten.Image{"tile": tile}}
	baseline, err := authoring.Compile(project, assets, authoring.Options{})
	if err != nil {
		c.err = err
		return
	}
	defer baseline.Close()
	project.Layers[0].Passes = []authoring.PostEffect{{Kind: "crt", CRT: &authoring.CRTPost{
		NormalizeSource: true, OutsideTransparent: true, Opacity: &half,
	}}}
	processed, err := authoring.Compile(project, assets, authoring.Options{})
	if err != nil {
		c.err = err
		return
	}
	defer processed.Close()
	readCenter := func(effect *authoring.Compiled) (byte, error) {
		if err := effect.Update(kit.Frame{}); err != nil {
			return 0, err
		}
		out := ebiten.NewImage(16, 16)
		defer out.Deallocate()
		effect.Draw(out)
		pixels := make([]byte, 16*16*4)
		out.ReadPixels(pixels)
		return pixels[(8*16+8)*4+3], nil
	}
	baseAlpha, err := readCenter(baseline)
	if err != nil {
		c.err = err
		return
	}
	processedAlpha, err := readCenter(processed)
	if err != nil {
		c.err = err
		return
	}
	if delta := int(processedAlpha) - int(baseAlpha); delta < -1 || delta > 1 {
		c.err = fmt.Errorf("faded CRT changed partial-alpha coverage from %d to %d", baseAlpha, processedAlpha)
	}
}

func TestMain(m *testing.M) {
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	file, err := os.Open("../projects/composed-effects.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	project, err := authoring.Decode(file)
	_ = file.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	noLens := *project
	noLens.Passes = nil
	noTitlePass := noLens
	noTitlePass.Layers = append([]authoring.Layer(nil), project.Layers...)
	noTitlePass.Layers[1].Passes = nil
	directory, err := os.MkdirTemp("", "dck-authoring-post-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var check *postCompositionCheck
	err = capture.Run(capture.Config{Directory: directory, Frames: []int{0}, Width: 640, Height: 360}, func() (ebiten.Game, error) {
		full, err := New(project)
		if err != nil {
			return nil, err
		}
		withoutLens, err := New(&noLens)
		if err != nil {
			_ = full.Close()
			return nil, err
		}
		withoutTitle, err := New(&noTitlePass)
		if err != nil {
			_ = full.Close()
			_ = withoutLens.Close()
			return nil, err
		}
		check = &postCompositionCheck{full: full, noLens: withoutLens, noTitlePass: withoutTitle}
		return check, nil
	})
	if check != nil {
		_ = check.full.Close()
		_ = check.noLens.Close()
		_ = check.noTitlePass.Close()
		if err == nil {
			err = check.err
		}
		if err == nil && !check.checked {
			err = fmt.Errorf("post composition was not drawn")
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
