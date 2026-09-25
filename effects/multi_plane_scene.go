package effects

import (
	"errors"
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// MultiPlaneSceneConfig combines a moving background, sampled logo rows,
// front/back center image and any configured DCK scrolling transport. Images
// are borrowed. Viewport is in destination pixels; StageSize is its native
// coordinate system. NativeStage composes logo and text before scaling, while
// the direct mode draws every layer into the destination's clipped viewport.
type MultiPlaneSceneConfig struct {
	Mountains, Logo  *ebiten.Image
	LogoSource       image.Rectangle
	CenterSource     image.Rectangle // Used when Center.Front is nil.
	Bands            composite.BandsConfig
	Rows             composite.ProfileImageConfig
	Center           sprites.AxisFlipConfig
	Scroll           scrolling.Config
	Viewport         image.Rectangle
	StageSize        image.Point
	CenterX, CenterY float64
	NativeStage      bool
	Filter           ebiten.Filter
}

// MultiPlaneScene owns its effect clocks and optional native-size surfaces.
// It leaves final canvas clearing, music and interaction to the host demo.
type MultiPlaneScene struct {
	config     MultiPlaneSceneConfig
	bands      *composite.Bands
	rows       *composite.ProfileImage
	center     *sprites.AxisFlip
	scroll     *scrolling.Scrolling
	background *ebiten.Image
	stage      *ebiten.Image
}

func NewMultiPlaneScene(config MultiPlaneSceneConfig) (*MultiPlaneScene, error) {
	if config.Mountains == nil || config.Logo == nil || config.Viewport.Empty() ||
		config.StageSize.X <= 0 || config.StageSize.Y <= 0 ||
		config.LogoSource.Intersect(config.Logo.Bounds()).Empty() ||
		math.IsNaN(config.CenterX) || math.IsInf(config.CenterX, 0) ||
		math.IsNaN(config.CenterY) || math.IsInf(config.CenterY, 0) {
		return nil, fmt.Errorf("effects: invalid multi-plane scene assets or placement")
	}
	if config.Center.Front == nil {
		if config.CenterSource.Empty() || !config.CenterSource.In(config.Logo.Bounds()) {
			return nil, fmt.Errorf("effects: missing multi-plane center image")
		}
		config.Center.Front = config.Logo.SubImage(config.CenterSource).(*ebiten.Image)
	}
	scene := &MultiPlaneScene{config: config}
	var err error
	scene.bands, err = composite.NewBands(config.Bands)
	if err != nil {
		return nil, err
	}
	scene.rows, err = composite.NewProfileImage(config.Logo.SubImage(config.LogoSource.Intersect(config.Logo.Bounds())).(*ebiten.Image), config.Rows)
	if err != nil {
		return nil, err
	}
	scene.center, err = sprites.NewAxisFlip(config.Center)
	if err != nil {
		scene.Close()
		return nil, err
	}
	scene.scroll, err = scrolling.New(config.Scroll)
	if err != nil {
		scene.Close()
		return nil, err
	}
	if config.NativeStage {
		opts := &ebiten.NewImageOptions{Unmanaged: true}
		scene.background = ebiten.NewImageWithOptions(image.Rect(0, 0, config.Viewport.Dx(), config.Viewport.Dy()), opts)
		scene.stage = ebiten.NewImageWithOptions(image.Rectangle{Max: config.StageSize}, opts)
	}
	return scene, nil
}

// Update advances every constituent once in the source's authored order.
// Draw remains read-only and may be called again without changing motion.
func (scene *MultiPlaneScene) Update(frame kit.Frame) error {
	scene.bands.Step()
	scene.rows.Advance()
	scene.center.Step()
	return scene.scroll.Update(frame)
}

func (scene *MultiPlaneScene) Draw(dst *ebiten.Image) {
	if scene == nil || dst == nil {
		return
	}
	c := scene.config
	if scene.stage != nil {
		scene.background.Clear()
		scene.stage.Clear()
		scene.bands.DrawAt(scene.background, c.Mountains, 0, 0)
		backgroundOptions := ebiten.DrawImageOptions{Filter: c.Filter, Blend: ebiten.BlendSourceOver}
		backgroundOptions.GeoM.Translate(float64(c.Viewport.Min.X), float64(c.Viewport.Min.Y))
		dst.DrawImage(scene.background, &backgroundOptions)
		scene.rows.Draw(scene.stage)
		scene.center.DrawAt(scene.stage, c.CenterX, c.CenterY)
		scene.scroll.Draw(scene.stage)
		stageOptions := ebiten.DrawImageOptions{Filter: c.Filter, Blend: ebiten.BlendSourceOver}
		stageOptions.GeoM.Scale(float64(c.Viewport.Dx())/float64(c.StageSize.X), float64(c.Viewport.Dy())/float64(c.StageSize.Y))
		stageOptions.GeoM.Translate(float64(c.Viewport.Min.X), float64(c.Viewport.Min.Y))
		dst.DrawImage(scene.stage, &stageOptions)
		return
	}
	clip := c.Viewport.Intersect(dst.Bounds())
	if clip.Empty() {
		return
	}
	view := dst.SubImage(clip).(*ebiten.Image)
	scene.bands.DrawAt(view, c.Mountains, float64(c.Viewport.Min.X), float64(c.Viewport.Min.Y))
	scene.rows.Draw(dst)
	parent := ebiten.GeoM{}
	parent.Scale(float64(c.Viewport.Dx())/float64(c.StageSize.X), float64(c.Viewport.Dy())/float64(c.StageSize.Y))
	parent.Translate(float64(c.Viewport.Min.X), float64(c.Viewport.Min.Y))
	scene.center.DrawAtWith(dst, c.CenterX, c.CenterY, parent)
	scene.scroll.Draw(view)
}

// Accessors expose live parameters for an inspector or timeline cue.
func (scene *MultiPlaneScene) Bands() *composite.Bands         { return scene.bands }
func (scene *MultiPlaneScene) Rows() *composite.ProfileImage   { return scene.rows }
func (scene *MultiPlaneScene) Center() *sprites.AxisFlip       { return scene.center }
func (scene *MultiPlaneScene) Scrolling() *scrolling.Scrolling { return scene.scroll }

func (scene *MultiPlaneScene) Close() error {
	if scene == nil {
		return nil
	}
	var closeErr error
	if scene.scroll != nil {
		closeErr = errors.Join(closeErr, scene.scroll.Close())
		scene.scroll = nil
	}
	if scene.rows != nil {
		closeErr = errors.Join(closeErr, scene.rows.Close())
		scene.rows = nil
	}
	if scene.background != nil {
		scene.background.Deallocate()
		scene.background = nil
	}
	if scene.stage != nil {
		scene.stage.Deallocate()
		scene.stage = nil
	}
	return closeErr
}
