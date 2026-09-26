package composite

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/render"
)

// SurfaceImagePass applies a borrowed image after the layer's effect sources.
// Scale occurs before rotation and placement. Empty Source uses the full image.
type SurfaceImagePass struct {
	Image                   *ebiten.Image
	Source                  image.Rectangle
	X, Y                    float64
	ScaleX, ScaleY          float64
	Angle, AnchorX, AnchorY float64
	Filter                  ebiten.Filter
	Blend                   ebiten.Blend
}

// SurfaceOutput draws one copy of the completed layer on the destination.
type SurfaceOutput struct {
	X, Y           float64
	ScaleX, ScaleY float64
	Filter         ebiten.Filter
	Blend          ebiten.Blend
}

// SurfaceLayerConfig composes live effects and ordered image passes into one
// bounded surface, then places any number of output copies. OwnSources closes
// the input effects with the layer; borrowed artwork remains caller-owned.
// A layer may contain only image passes when no live background is needed.
type SurfaceLayerConfig struct {
	Width, Height int
	Sources       []kit.Effect
	Passes        []SurfaceImagePass
	Outputs       []SurfaceOutput
	Background    color.Color // Nil clears to transparent before drawing.
	Retain        bool        // Keep previous surface pixels for feedback.
	OwnSources    bool
}

type surfacePass struct {
	config SurfaceImagePass
	view   *ebiten.Image
}

// SurfaceLayer owns exactly one persistent working image. Sources are updated
// in configured order; Draw never advances their animation clocks.
type SurfaceLayer struct {
	config SurfaceLayerConfig
	passes []surfacePass
	canvas *ebiten.Image
}

func NewSurfaceLayer(c SurfaceLayerConfig) (*SurfaceLayer, error) {
	if c.Width < 1 || c.Height < 1 || c.Width > 8192 || c.Height > 8192 ||
		len(c.Sources) == 0 && len(c.Passes) == 0 || len(c.Sources) > 256 || len(c.Passes) > 256 ||
		len(c.Outputs) == 0 || len(c.Outputs) > 256 {
		return nil, fmt.Errorf("composite: invalid surface layer dimensions or content")
	}
	for _, source := range c.Sources {
		if source == nil {
			return nil, fmt.Errorf("composite: nil surface layer source")
		}
	}
	passes := make([]surfacePass, len(c.Passes))
	for i, pass := range c.Passes {
		if pass.Image == nil {
			return nil, fmt.Errorf("composite: nil surface layer material")
		}
		for _, value := range [...]float64{pass.X, pass.Y, pass.ScaleX, pass.ScaleY,
			pass.Angle, pass.AnchorX, pass.AnchorY} {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return nil, fmt.Errorf("composite: nonfinite surface layer pass")
			}
		}
		if pass.ScaleX == 0 {
			pass.ScaleX = 1
		}
		if pass.ScaleY == 0 {
			pass.ScaleY = 1
		}
		view := pass.Image
		if !pass.Source.Empty() {
			if !pass.Source.In(pass.Image.Bounds()) {
				return nil, fmt.Errorf("composite: surface layer crop exceeds image")
			}
			view = pass.Image.SubImage(pass.Source).(*ebiten.Image)
		}
		passes[i] = surfacePass{config: pass, view: view}
	}
	outputs := make([]SurfaceOutput, len(c.Outputs))
	for i, output := range c.Outputs {
		for _, value := range [...]float64{output.X, output.Y, output.ScaleX, output.ScaleY} {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return nil, fmt.Errorf("composite: nonfinite surface layer output")
			}
		}
		if output.ScaleX == 0 {
			output.ScaleX = 1
		}
		if output.ScaleY == 0 {
			output.ScaleY = 1
		}
		outputs[i] = output
	}
	c.Sources = append([]kit.Effect(nil), c.Sources...)
	c.Passes = nil
	c.Outputs = outputs
	return &SurfaceLayer{config: c, passes: passes,
		canvas: render.NewSurface(c.Width, c.Height)}, nil
}

// SetPassPosition moves one image pass without rebuilding the layer or its
// working surface. Motion clocks can call this once per logical update.
func (l *SurfaceLayer) SetPassPosition(index int, x, y float64) error {
	if l == nil || l.canvas == nil || index < 0 || index >= len(l.passes) ||
		math.IsNaN(x) || math.IsInf(x, 0) || math.IsNaN(y) || math.IsInf(y, 0) {
		return fmt.Errorf("composite: invalid surface layer pass position")
	}
	l.passes[index].config.X = x
	l.passes[index].config.Y = y
	return nil
}

// SetPassTransform changes one pass's placement, scale and rotation without
// rebuilding its image view or the layer surface. Zero scale intentionally
// hides the pass for cue-driven transitions.
func (l *SurfaceLayer) SetPassTransform(index int, x, y, scaleX, scaleY, angle float64) error {
	if l == nil || l.canvas == nil || index < 0 || index >= len(l.passes) {
		return fmt.Errorf("composite: invalid surface layer pass")
	}
	for _, value := range [...]float64{x, y, scaleX, scaleY, angle} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("composite: nonfinite surface layer transform")
		}
	}
	c := &l.passes[index].config
	c.X, c.Y, c.ScaleX, c.ScaleY, c.Angle = x, y, scaleX, scaleY, angle
	return nil
}

// SetPassFilter edits source sampling for a single image pass at a visual cue.
func (l *SurfaceLayer) SetPassFilter(index int, filter ebiten.Filter) error {
	if l == nil || l.canvas == nil || index < 0 || index >= len(l.passes) {
		return fmt.Errorf("composite: invalid surface layer pass filter")
	}
	l.passes[index].config.Filter = filter
	return nil
}

func (l *SurfaceLayer) Update(frame kit.Frame) error {
	for _, source := range l.config.Sources {
		if err := source.Update(frame); err != nil {
			return err
		}
	}
	return nil
}

func (l *SurfaceLayer) Draw(dst *ebiten.Image) {
	if l == nil || dst == nil || l.canvas == nil {
		return
	}
	if l.config.Background != nil {
		l.canvas.Fill(l.config.Background)
	} else if !l.config.Retain {
		l.canvas.Clear()
	}
	for _, source := range l.config.Sources {
		source.Draw(l.canvas)
	}
	for _, pass := range l.passes {
		c := pass.config
		var op ebiten.DrawImageOptions
		op.Filter, op.Blend = c.Filter, c.Blend
		op.GeoM.Translate(-c.AnchorX, -c.AnchorY)
		op.GeoM.Scale(c.ScaleX, c.ScaleY)
		op.GeoM.Rotate(c.Angle)
		op.GeoM.Translate(c.X, c.Y)
		l.canvas.DrawImage(pass.view, &op)
	}
	for _, output := range l.config.Outputs {
		var op ebiten.DrawImageOptions
		op.Filter, op.Blend = output.Filter, output.Blend
		op.GeoM.Scale(output.ScaleX, output.ScaleY)
		op.GeoM.Translate(output.X, output.Y)
		dst.DrawImage(l.canvas, &op)
	}
}

// Canvas exposes the prepared image until the next Draw or Close.
func (l *SurfaceLayer) Canvas() *ebiten.Image { return l.canvas }

func (l *SurfaceLayer) Close() error {
	if l == nil || l.canvas == nil {
		return nil
	}
	l.canvas.Deallocate()
	l.canvas = nil
	if !l.config.OwnSources {
		return nil
	}
	var closeErr error
	for _, source := range l.config.Sources {
		closeErr = errors.Join(closeErr, kit.Close(source))
	}
	return closeErr
}
