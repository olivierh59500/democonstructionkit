package motion

import (
	"fmt"
	"math"
)

// WalkLayer has a per-input displacement and independent wrap rules in each
// direction. Positive input moves toward Lower; negative input moves toward
// Upper. The single matching rule applies, even when its restart is exactly
// on the opposite boundary.
type WalkLayer struct {
	Start, Speed float64
	Lower, Upper *WrapLimit
}

type WalkParallaxConfig struct{ Layers []WalkLayer }

// WalkParallax keeps background, banner or logo positions synchronized with
// one movement input while preserving each layer's speed and wrap period.
type WalkParallax struct {
	config    []WalkLayer
	positions []float64
}

func NewWalkParallax(config WalkParallaxConfig) (*WalkParallax, error) {
	if len(config.Layers) == 0 || len(config.Layers) > 1<<20 {
		return nil, fmt.Errorf("motion: invalid walk parallax layer count")
	}
	layers := make([]WalkLayer, len(config.Layers))
	for i, layer := range config.Layers {
		if math.IsNaN(layer.Start) || math.IsInf(layer.Start, 0) || math.IsNaN(layer.Speed) || math.IsInf(layer.Speed, 0) || layer.Speed < 0 {
			return nil, fmt.Errorf("motion: invalid walk parallax position or speed")
		}
		for _, limit := range [...]*WrapLimit{layer.Lower, layer.Upper} {
			if limit != nil && (math.IsNaN(limit.Boundary) || math.IsInf(limit.Boundary, 0) || math.IsNaN(limit.Restart) || math.IsInf(limit.Restart, 0)) {
				return nil, fmt.Errorf("motion: nonfinite walk parallax limit")
			}
		}
		layers[i] = layer
		if layer.Lower != nil {
			copy := *layer.Lower
			layers[i].Lower = &copy
		}
		if layer.Upper != nil {
			copy := *layer.Upper
			layers[i].Upper = &copy
		}
	}
	walk := &WalkParallax{config: layers, positions: make([]float64, len(layers))}
	walk.Reset()
	return walk, nil
}

func (walk *WalkParallax) Len() int             { return len(walk.positions) }
func (walk *WalkParallax) At(index int) float64 { return walk.positions[index] }

// Advance applies one caller-chosen signed movement without allocating.
func (walk *WalkParallax) Advance(direction int) {
	if direction == 0 {
		return
	}
	for i, layer := range walk.config {
		position := walk.positions[i] - float64(direction)*layer.Speed
		if direction > 0 && layer.Lower != nil && (position < layer.Lower.Boundary || layer.Lower.Inclusive && position <= layer.Lower.Boundary) {
			position = layer.Lower.Restart
		} else if direction < 0 && layer.Upper != nil && (position > layer.Upper.Boundary || layer.Upper.Inclusive && position >= layer.Upper.Boundary) {
			position = layer.Upper.Restart
		}
		walk.positions[i] = position
	}
}

// Set places a layer at an authored camera position, such as a selected door.
func (walk *WalkParallax) Set(index int, position float64) {
	walk.positions[index] = position
}

func (walk *WalkParallax) Reset() {
	for i, layer := range walk.config {
		walk.positions[i] = layer.Start
	}
}
