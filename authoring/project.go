// Package authoring loads and compiles a versioned, serializable subset of DCK.
// It is an authoring foundation rather than a graphical editor. Images and fonts
// are resolved by the host; the core neither opens files nor uses global assets.
package authoring

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/olivierh59500/democonstructionkit/modulation"
)

const Version = 1

// Units makes saved-project conventions explicit. Version 1 requires these
// exact units; it never silently converts legacy per-frame speeds to seconds.
type Units struct {
	Distance string `json:"distance"`
	Time     string `json:"time"`
	Angle    string `json:"angle"`
}

func DefaultUnits() Units { return Units{Distance: "pixels", Time: "seconds", Angle: "radians"} }

type Canvas struct {
	Width      int      `json:"width"`
	Height     int      `json:"height"`
	TPS        int      `json:"tps"`
	Background [4]uint8 `json:"background"`
}

// Assets maps stable identifiers to "image" or "font". Physical locations and
// font atlas metrics belong to Resolver and can vary between desktop and mobile.
type Project struct {
	Version int                        `json:"version"`
	Units   Units                      `json:"units"`
	Canvas  Canvas                     `json:"canvas"`
	BPM     float64                    `json:"bpm,omitempty"`
	Assets  map[string]string          `json:"assets"`
	Signals map[string]modulation.Spec `json:"signals,omitempty"`
	Layers  []Layer                    `json:"layers"`
}

type Window struct {
	Start    float64 `json:"start,omitempty"`
	Duration float64 `json:"duration,omitempty"`
	FadeIn   float64 `json:"fadeIn,omitempty"`
	FadeOut  float64 `json:"fadeOut,omitempty"`
}

// Layer order is drawing order. Exactly one kind-specific configuration is
// required. Blend combines the complete layer, after its internal image blends.
type Layer struct {
	ID         string       `json:"id"`
	Kind       string       `json:"kind"`
	Window     Window       `json:"window,omitempty"`
	LocalTime  bool         `json:"localTime,omitempty"`
	Blend      string       `json:"blend,omitempty"`
	Scroll     *Scroll      `json:"scroll,omitempty"`
	Sprites    *SpriteGroup `json:"sprites,omitempty"`
	Background *Background  `json:"background,omitempty"`
	JellyCube  *JellyCube   `json:"jellyCube,omitempty"`
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
type Rect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Wave follows DCK's spatial-wave convention: radians per pixel and per second.
// Modulation oscillators instead use cycles per selected time unit.
type Wave struct {
	Amplitude float64 `json:"amplitude"`
	Spatial   float64 `json:"spatial,omitempty"`
	Speed     float64 `json:"speed,omitempty"`
	Phase     float64 `json:"phase,omitempty"`
}
type Zoom struct {
	Base  Point `json:"base"`
	Pivot Point `json:"pivot,omitempty"`
	Wave  Wave  `json:"wave"`
}
type Perspective struct {
	Focal        float64 `json:"focal"`
	Near         float64 `json:"near"`
	Depth        float64 `json:"depth"`
	Center       Point   `json:"center"`
	DepthWave    Wave    `json:"depthWave"`
	VerticalWave Wave    `json:"verticalWave"`
}
type Path struct {
	Points        []Point `json:"points"`
	Closed        bool    `json:"closed,omitempty"`
	SplineSamples int     `json:"splineSamples,omitempty"`
	Offset        float64 `json:"offset,omitempty"`
	NormalOffset  float64 `json:"normalOffset,omitempty"`
	Rotation      float64 `json:"rotation,omitempty"`
	Orient        bool    `json:"orient,omitempty"`
	Clip          bool    `json:"clip,omitempty"`
}
type Mode struct {
	Kind        string       `json:"kind"`
	Wave        *Wave        `json:"wave,omitempty"`
	Zoom        *Zoom        `json:"zoom,omitempty"`
	Perspective *Perspective `json:"perspective,omitempty"`
	Path        *Path        `json:"path,omitempty"`
}
type Cue struct {
	At   float64 `json:"at"`
	Mode string  `json:"mode"`
}
type Page struct {
	Width      float64 `json:"width"`
	LineHeight float64 `json:"lineHeight"`
	Align      string  `json:"align,omitempty"`
}

// Fonts maps text-control names to manifest font IDs. Controls is "braces" or
// empty for literal text. Modes can be selected by {shape:name} or timed cues.
type Scroll struct {
	Text           string            `json:"text"`
	Fonts          map[string]string `json:"fonts"`
	Font           string            `json:"font"`
	Controls       string            `json:"controls,omitempty"`
	Origin         Point             `json:"origin,omitempty"`
	Speed          float64           `json:"speed"`
	Gap            float64           `json:"gap,omitempty"`
	Advance        float64           `json:"advance,omitempty"`
	Vertical       bool              `json:"vertical,omitempty"`
	Repeat         bool              `json:"repeat,omitempty"`
	RepeatBounds   *Rect             `json:"repeatBounds,omitempty"`
	Page           *Page             `json:"page,omitempty"`
	Modes          map[string]Mode   `json:"modes,omitempty"`
	Mode           string            `json:"mode,omitempty"`
	Sequence       []Cue             `json:"sequence,omitempty"`
	SequencePeriod float64           `json:"sequencePeriod,omitempty"`
}

type Orbit struct {
	Center          Point   `json:"center"`
	Radius          Point   `json:"radius"`
	XRate           float64 `json:"xRate"`
	YRate           float64 `json:"yRate"`
	ModulationRate  float64 `json:"modulationRate"`
	ModulationPhase float64 `json:"modulationPhase"`
	ModulationDepth float64 `json:"modulationDepth"`
}
type Weave struct {
	Center                  Point   `json:"center"`
	HorizontalAmplitude     float64 `json:"horizontalAmplitude"`
	VerticalAmplitude       float64 `json:"verticalAmplitude"`
	VerticalSecondAmplitude float64 `json:"verticalSecondAmplitude"`
	HorizontalPeriod        float64 `json:"horizontalPeriod"`
	EnvelopePeriod          float64 `json:"envelopePeriod"`
	VerticalPeriod          float64 `json:"verticalPeriod"`
	VerticalSecondPeriod    float64 `json:"verticalSecondPeriod"`
	Spacing                 float64 `json:"spacing"`
}

// Bindings maps x, y, scaleX, scaleY, angle or opacity to a Project.Signals ID.
// Offsets/angle are additive; scale and opacity multiply the base sprite pose.
type Bindings map[string]string
type SpriteGroup struct {
	Images        []string   `json:"images"`
	Count         int        `json:"count"`
	FPS           float64    `json:"fps,omitempty"`
	FrameStride   int        `json:"frameStride,omitempty"`
	FrameOffset   int        `json:"frameOffset,omitempty"`
	Points        []Point    `json:"points,omitempty"`
	Closed        bool       `json:"closed,omitempty"`
	SplineSamples int        `json:"splineSamples,omitempty"`
	Orbit         *Orbit     `json:"orbit,omitempty"`
	Weave         *Weave     `json:"weave,omitempty"`
	Origin        Point      `json:"origin,omitempty"`
	Velocity      Point      `json:"velocity,omitempty"`
	Spacing       Point      `json:"spacing,omitempty"`
	Speed         float64    `json:"speed,omitempty"`
	Phase         float64    `json:"phase,omitempty"`
	PhaseSpacing  float64    `json:"phaseSpacing,omitempty"`
	Delay         float64    `json:"delay,omitempty"`
	Orient        bool       `json:"orient,omitempty"`
	Scale         Point      `json:"scale,omitempty"`
	Angle         float64    `json:"angle,omitempty"`
	Anchor        Point      `json:"anchor,omitempty"`
	Filter        string     `json:"filter,omitempty"`
	Blend         string     `json:"blend,omitempty"`
	Reverse       bool       `json:"reverse,omitempty"`
	Signals       Bindings   `json:"signals,omitempty"`
	PerInstance   []Bindings `json:"perInstance,omitempty"`
}

type Background struct {
	Image          string `json:"image"`
	Source         *Rect  `json:"source,omitempty"`
	Period         Point  `json:"period,omitempty"`
	Scale          Point  `json:"scale,omitempty"`
	Parallax       Point  `json:"parallax,omitempty"`
	Origin         Point  `json:"origin,omitempty"`
	Camera         Point  `json:"camera,omitempty"`
	CameraVelocity Point  `json:"cameraVelocity,omitempty"`
	Velocity       Point  `json:"velocity,omitempty"`
	Filter         string `json:"filter,omitempty"`
	Blend          string `json:"blend,omitempty"`
}

// Decode rejects unknown fields, duplicate object keys, trailing JSON values and
// unsupported configurations before the host allocates any graphics resources.
func Decode(r io.Reader) (*Project, error) {
	b, err := io.ReadAll(io.LimitReader(r, 16<<20+1))
	if err != nil {
		return nil, err
	}
	if len(b) > 16<<20 {
		return nil, fmt.Errorf("authoring: document exceeds 16 MiB")
	}
	return decodeBytes(b)
}

// Encode writes readable canonical field names after validating the project.
func Encode(w io.Writer, p Project) error {
	if err := p.Validate(); err != nil {
		return err
	}
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	return e.Encode(p)
}
