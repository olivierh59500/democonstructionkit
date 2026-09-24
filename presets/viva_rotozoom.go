package presets

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// VivaRotozoomConfig exposes the entrance, independent phase clocks and nested
// orbit of a repeated rotozoom texture. The default values reproduce Viva TCB;
// another image can use different dimensions, phases, speeds and amplitudes.
type VivaRotozoomConfig struct {
	Width, Height, TilePhaseX, TilePhaseY                       float64
	EntryPanStep, EntryPanDistance, EntryAngleDegrees           float64
	EntryAngleSteps                                             int
	OrbitStep, ZoomStep, RotationStep, OrbitStart, ZoomStart    float64
	OrbitXAmplitude, OrbitYAmplitude, OrbitXRate, OrbitYRate    float64
	OrbitLag, OrbitPhase, ZoomPhase, RotationPhase              float64
	ZoomBase, ZoomAmplitude, RotationBaseDegrees, RotationScale float64
	RotationRate, RotationLag                                   float64
	Filter                                                      ebiten.Filter
}

func DefaultVivaRotozoomConfig(width, height float64) VivaRotozoomConfig {
	return VivaRotozoomConfig{
		Width: width, Height: height, TilePhaseX: width * 8, TilePhaseY: height * 8,
		EntryPanStep: 4, EntryPanDistance: width * 2, EntryAngleDegrees: .3, EntryAngleSteps: 45,
		OrbitStep: .008, ZoomStep: .003, RotationStep: .005, OrbitStart: 2, ZoomStart: 1.5,
		OrbitXAmplitude: width / 4, OrbitYAmplitude: height / 2.7, OrbitXRate: 4, OrbitYRate: 2.3,
		OrbitLag: .1, ZoomBase: .5, ZoomAmplitude: 2.5, RotationBaseDegrees: 90, RotationScale: .3,
		RotationRate: 4, RotationLag: .01,
	}
}

// VivaRotozoom is an editable motion program for composite.RotozoomBackground.
// The first stage pans and turns the tile, then orbit, zoom and rotation clocks
// activate in order without resetting the current center or texture phase.
type VivaRotozoom struct {
	config                        VivaRotozoomConfig
	stage, entryAngles            int
	entryX, orbit, zoom, rotation float64
}

func NewVivaRotozoom(c VivaRotozoomConfig) (*VivaRotozoom, error) {
	if c.Width <= 0 || c.Height <= 0 || c.EntryPanStep <= 0 || c.EntryPanDistance <= 0 || c.EntryAngleSteps < 1 ||
		c.OrbitStep <= 0 || c.ZoomStep <= 0 || c.RotationStep <= 0 || c.OrbitStart <= 0 || c.ZoomStart <= 0 || c.ZoomBase <= 0 || c.ZoomAmplitude < 0 {
		return nil, fmt.Errorf("presets: invalid rotozoom dimensions or clocks")
	}
	for _, value := range [...]float64{c.Width, c.Height, c.TilePhaseX, c.TilePhaseY, c.EntryPanStep, c.EntryPanDistance, c.EntryAngleDegrees,
		c.OrbitStep, c.ZoomStep, c.RotationStep, c.OrbitStart, c.ZoomStart, c.OrbitXAmplitude, c.OrbitYAmplitude,
		c.OrbitXRate, c.OrbitYRate, c.OrbitLag, c.OrbitPhase, c.ZoomPhase, c.RotationPhase, c.ZoomBase,
		c.ZoomAmplitude, c.RotationBaseDegrees, c.RotationScale, c.RotationRate, c.RotationLag} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("presets: nonfinite rotozoom setting")
		}
	}
	return &VivaRotozoom{config: c}, nil
}

func (v *VivaRotozoom) Update(kit.Frame) error {
	if v == nil {
		return fmt.Errorf("presets: nil rotozoom program")
	}
	c := v.config
	if v.stage >= 1 {
		v.orbit += c.OrbitStep
	}
	if v.stage >= 2 {
		v.zoom += c.ZoomStep
	}
	if v.stage >= 3 {
		v.rotation += c.RotationStep
	}
	if v.orbit >= c.OrbitStart {
		v.stage = 2
	}
	if v.zoom >= c.ZoomStart {
		v.stage = 3
	}
	if v.stage == 0 {
		v.entryX -= c.EntryPanStep
		if v.entryX <= -c.EntryPanDistance {
			v.entryAngles++
		}
		if v.entryAngles >= c.EntryAngleSteps {
			v.stage = 1
		}
	}
	return nil
}

func (v *VivaRotozoom) Repetition() composite.Repetition {
	c := v.config
	p := composite.Repetition{PhaseX: c.TilePhaseX, PhaseY: c.TilePhaseY, Filter: c.Filter}
	if v.stage == 0 {
		p.CenterX = c.Width/2 + v.entryX
		p.CenterY = c.Height / 2
		p.Zoom = 1
		p.Rotation = -float64(v.entryAngles) * c.EntryAngleDegrees * math.Pi / 180
		return p
	}
	orbit := v.orbit + c.OrbitPhase
	curve := math.Cos(orbit - c.OrbitLag)
	p.CenterX = c.Width/2 + c.OrbitXAmplitude*math.Cos(orbit*c.OrbitXRate-curve)
	p.CenterY = c.Height/2 + c.OrbitYAmplitude*-math.Sin(orbit*c.OrbitYRate-curve)
	p.Zoom = c.ZoomBase + math.Abs(math.Sin(v.zoom+c.ZoomPhase)*c.ZoomAmplitude)
	rotation := v.rotation + c.RotationPhase
	p.Rotation = (c.RotationBaseDegrees * math.Cos(rotation*c.RotationRate-math.Cos(rotation-c.RotationLag))) * c.RotationScale * math.Pi / 180
	return p
}

var _ composite.RotozoomProgram = (*VivaRotozoom)(nil)
