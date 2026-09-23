// Package scene demonstrates multiple complete cube effects in one composition.
package scene

import (
	"bytes"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/olivierh59500/democonstructionkit/authoring"
)

type Scene struct {
	*authoring.Compiled
	Project       authoring.Project
	Width, Height int
}

func value(v float64) *float64 { return &v }

// Project builds data only: no model vertices, mode controller or cube drawing
// logic belongs in this example. Each layer compiles to its own DCK component.
func Project(width, height int) authoring.Project {
	sx, sy := float64(width)/640, float64(height)/360
	p := authoring.Project{Version: authoring.Version, Units: authoring.DefaultUnits(), Canvas: authoring.Canvas{Width: width, Height: height, TPS: 60, Background: [4]uint8{8, 12, 24, 255}}}
	palettes := [][6][4]uint8{
		{{224, 160, 192, 255}, {224, 160, 192, 255}, {224, 96, 192, 255}, {224, 96, 192, 255}, {224, 224, 224, 255}, {224, 224, 224, 255}},
		{{70, 190, 255, 255}, {70, 190, 255, 255}, {40, 105, 180, 255}, {40, 105, 180, 255}, {185, 240, 255, 255}, {185, 240, 255, 255}},
		{{255, 170, 65, 255}, {255, 170, 65, 255}, {200, 95, 30, 255}, {200, 95, 30, 255}, {255, 230, 150, 255}, {255, 230, 150, 255}},
	}
	for i, x := range []float64{110, 320, 530} {
		cube := &authoring.JellyCube{Center: &authoring.Point{X: x * sx, Y: 185 * sy}, HalfEdge: value(52 * sx), Colors: &palettes[i], Transition: value(.75), Phase: value(float64(i) * 3)}
		switch i {
		case 1:
			cube.Speed = value(.8)
			cube.Steps = []authoring.JellyCubeStep{{Mode: "bounce", Duration: 3}, {Mode: "swing", Duration: 3}, {Mode: "normal", Duration: 4}}
			cube.Deformation = &authoring.JellyDeformation{Wobble: value(.3), Ripple: value(1.5)}
		case 2:
			cube.Speed = value(1.2)
			cube.Steps = []authoring.JellyCubeStep{{Mode: "pulsate", Duration: 3}, {Mode: "tumble", Duration: 4}, {Mode: "normal", Duration: 3}}
			cube.Deformation = &authoring.JellyDeformation{Twist: value(1.8), Translation: value(.4)}
		}
		p.Layers = append(p.Layers, authoring.Layer{ID: fmt.Sprintf("cube-%d", i+1), Kind: "jelly_cube", JellyCube: cube})
	}
	return p
}

func NewSize(width, height int) (*Scene, error) { return New(Project(width, height)) }
func New(p authoring.Project) (*Scene, error) {
	var json bytes.Buffer
	if err := authoring.Encode(&json, p); err != nil {
		return nil, err
	}
	loaded, err := authoring.Decode(&json)
	if err != nil {
		return nil, err
	}
	compiled, err := authoring.Compile(*loaded, authoring.Assets{}, authoring.Options{})
	if err != nil {
		return nil, err
	}
	return &Scene{Compiled: compiled, Project: *loaded, Width: p.Canvas.Width, Height: p.Canvas.Height}, nil
}
func (s *Scene) Draw(dst *ebiten.Image) {
	s.Compiled.Draw(dst)
	label := "THREE CUBES / ONE DCK EFFECT"
	ebitenutil.DebugPrintAt(dst, label, max(0, (s.Width-len(label)*6)/2), 12)
	if s.Width >= 640 {
		ebitenutil.DebugPrintAt(dst, "Different palettes, timing, mode order and deformation", 150, s.Height-25)
	}
}

// SurfaceBytes counts the layer fade target and three owned white texels.
// The engine's debug-font atlas and backend storage are excluded.
func (s *Scene) SurfaceBytes() int64 { return int64(s.Width*s.Height*4 + len(s.Project.Layers)*4) }
