package composite

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// TourSource updates continuously and renders into either a direct viewport or
// one retained transition canvas. A production keeps ownership of each source.
type TourSource interface {
	Update() error
	Draw(*ebiten.Image)
}

// SceneTourConfig binds a pure CameraTour to up to four running scenes. Tile
// origins are world-space positions for the DrawImage fallback. A supplied
// shader is borrowed; ShaderSource is compiled once and owned by SceneTour.
// Canvases may be omitted for owned tile-sized images, or supplied individually
// as borrowed images by an existing host or capture fixture.
type SceneTourConfig struct {
	Camera                        *motion.CameraTour
	Sources                       []TourSource
	Names                         []string
	TileOrigins                   []image.Point
	TileWidth, TileHeight         int
	ViewportWidth, ViewportHeight int
	Canvases                      []*ebiten.Image
	Shader                        *ebiten.Shader
	ShaderSource                  []byte
	OnShaderError                 func(error)
}

// SceneTour keeps scenes synchronized, culls transition canvases by the camera
// mask and draws fixed views directly. It creates no full-viewport copy.
type SceneTour struct {
	config     SceneTourConfig
	canvases   []*ebiten.Image
	owned      []bool
	shader     *ebiten.Shader
	ownsShader bool
	center     [2]float32
	uniforms   map[string]any
	dirty      bool
}

func NewSceneTour(c SceneTourConfig) (*SceneTour, error) {
	n := len(c.Sources)
	if c.Camera == nil || n < 1 || n > 4 || c.TileWidth < 1 || c.TileHeight < 1 ||
		c.ViewportWidth < 1 || c.ViewportHeight < 1 || len(c.TileOrigins) != n ||
		len(c.Names) != 0 && len(c.Names) != n ||
		len(c.Canvases) != 0 && len(c.Canvases) != n {
		return nil, fmt.Errorf("composite: invalid scene tour dimensions or sources")
	}
	for i, source := range c.Sources {
		if source == nil {
			return nil, fmt.Errorf("composite: nil scene tour source %d", i)
		}
	}
	t := &SceneTour{config: c, canvases: make([]*ebiten.Image, n),
		owned: make([]bool, n), dirty: true}
	t.config.Sources = append([]TourSource(nil), c.Sources...)
	t.config.Names = append([]string(nil), c.Names...)
	t.config.TileOrigins = append([]image.Point(nil), c.TileOrigins...)
	for i := range t.canvases {
		if len(c.Canvases) != 0 && c.Canvases[i] != nil {
			if c.Canvases[i].Bounds().Dx() != c.TileWidth || c.Canvases[i].Bounds().Dy() != c.TileHeight {
				t.Close()
				return nil, fmt.Errorf("composite: wrong scene tour canvas size %d", i)
			}
			t.canvases[i] = c.Canvases[i]
			continue
		}
		t.canvases[i] = ebiten.NewImage(c.TileWidth, c.TileHeight)
		t.owned[i] = true
	}
	t.shader = c.Shader
	if t.shader == nil && len(c.ShaderSource) != 0 {
		var err error
		t.shader, err = ebiten.NewShader(c.ShaderSource)
		if err != nil {
			if c.OnShaderError != nil {
				c.OnShaderError(err)
			}
		} else {
			t.ownsShader = true
		}
	}
	t.uniforms = map[string]any{"CameraCenter": t.center[:], "CameraZoom": float32(1)}
	return t, nil
}

func (t *SceneTour) Update() error {
	t.dirty = true
	for i, source := range t.config.Sources {
		if err := source.Update(); err != nil {
			name := fmt.Sprintf("scene %d", i)
			if len(t.config.Names) != 0 && t.config.Names[i] != "" {
				name = t.config.Names[i]
			}
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	t.config.Camera.Step()
	return nil
}

func (t *SceneTour) Draw(dst *ebiten.Image) {
	if t == nil || dst == nil || !t.dirty {
		return
	}
	t.dirty = false
	pose := t.config.Camera.State()
	if pose.Direct >= 0 && pose.Direct < len(t.config.Sources) {
		t.config.Sources[pose.Direct].Draw(dst)
		return
	}
	for i, source := range t.config.Sources {
		if pose.VisibleMask&(uint64(1)<<i) != 0 {
			source.Draw(t.canvases[i])
		}
	}
	if t.shader != nil {
		t.center[0], t.center[1] = float32(pose.CenterX), float32(pose.CenterY)
		t.uniforms["CameraZoom"] = float32(pose.Zoom)
		options := ebiten.DrawRectShaderOptions{Uniforms: t.uniforms, Blend: ebiten.BlendCopy}
		for i, canvas := range t.canvases {
			if pose.VisibleMask&(uint64(1)<<i) != 0 {
				options.Images[i] = canvas
			}
		}
		dst.DrawRectShader(t.config.ViewportWidth, t.config.ViewportHeight, t.shader, &options)
		return
	}
	dst.Fill(color.Black)
	for i, canvas := range t.canvases {
		if pose.VisibleMask&(uint64(1)<<i) == 0 {
			continue
		}
		origin := t.config.TileOrigins[i]
		if math.IsNaN(pose.Zoom) || pose.Zoom <= 0 {
			continue
		}
		var options ebiten.DrawImageOptions
		options.GeoM.Translate(float64(origin.X)-pose.CenterX, float64(origin.Y)-pose.CenterY)
		options.GeoM.Scale(pose.Zoom, pose.Zoom)
		options.GeoM.Translate(float64(t.config.ViewportWidth)/2, float64(t.config.ViewportHeight)/2)
		dst.DrawImage(canvas, &options)
	}
}

// Camera exposes live segment state for cues or an authoring interface.
func (t *SceneTour) Camera() *motion.CameraTour { return t.config.Camera }

// Canvases returns borrowed, retained tile images until Close.
func (t *SceneTour) Canvases() []*ebiten.Image { return t.canvases }
func (t *SceneTour) Shader() *ebiten.Shader    { return t.shader }

// Close releases only canvases and shader compiled by this component.
func (t *SceneTour) Close() {
	if t == nil {
		return
	}
	for i, canvas := range t.canvases {
		if t.owned[i] && canvas != nil {
			canvas.Deallocate()
			t.canvases[i] = nil
		}
	}
	if t.ownsShader && t.shader != nil {
		t.shader.Deallocate()
		t.shader = nil
	}
}
