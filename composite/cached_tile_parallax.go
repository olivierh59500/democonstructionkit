package composite

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// CachedTileParallaxConfig describes a bounded repeated tile surface. The
// overscan covers camera offsets without redrawing tiles every frame. Camera
// divisors and wrap periods are independent of the tile's image dimensions.
type CachedTileParallaxConfig struct {
	Tile                         *ebiten.Image
	Width, Height                int
	PeriodX, PeriodY             int
	OverscanX, OverscanY         int
	DivisorX, DivisorY           float64
	WrapX, WrapY                 float64
	ClampNegative, PixelQuantize bool
	Unmanaged                    bool
	TileFilter, Filter           ebiten.Filter
	TileBlend, Blend             ebiten.Blend
}

// CachedTileParallax owns one bounded GPU surface and draws it once per frame.
// The source tile remains caller-owned. Camera sampling is read-only.
type CachedTileParallax struct {
	config  CachedTileParallaxConfig
	surface *ebiten.Image
}

func NewCachedTileParallax(config CachedTileParallaxConfig) (*CachedTileParallax, error) {
	if config.Tile == nil || config.Width < 1 || config.Height < 1 || config.Width > 8192 || config.Height > 8192 ||
		config.OverscanX < 0 || config.OverscanY < 0 || config.OverscanX > 8192-config.Width || config.OverscanY > 8192-config.Height {
		return nil, fmt.Errorf("composite: invalid cached tile surface")
	}
	if config.PeriodX == 0 {
		config.PeriodX = config.Tile.Bounds().Dx()
	}
	if config.PeriodY == 0 {
		config.PeriodY = config.Tile.Bounds().Dy()
	}
	if config.WrapX == 0 {
		config.WrapX = float64(config.PeriodX)
	}
	if config.WrapY == 0 {
		config.WrapY = float64(config.PeriodY)
	}
	if config.PeriodX < 1 || config.PeriodY < 1 || config.PeriodX > 8192 || config.PeriodY > 8192 || config.DivisorX <= 0 || config.DivisorY <= 0 || config.WrapX <= 0 || config.WrapY <= 0 {
		return nil, fmt.Errorf("composite: invalid cached tile periods or camera divisors")
	}
	for _, value := range [...]float64{config.DivisorX, config.DivisorY, config.WrapX, config.WrapY} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("composite: nonfinite cached tile setting")
		}
	}
	w, h := config.Width+config.OverscanX, config.Height+config.OverscanY
	countX, countY := (w+config.PeriodX-1)/config.PeriodX, (h+config.PeriodY-1)/config.PeriodY
	if countX > 16384/countY {
		return nil, fmt.Errorf("composite: cached tile copy budget exceeded")
	}
	var surface *ebiten.Image
	if config.Unmanaged {
		surface = ebiten.NewImageWithOptions(image.Rect(0, 0, w, h), &ebiten.NewImageOptions{Unmanaged: true})
	} else {
		surface = ebiten.NewImage(w, h)
	}
	for y := 0; y < h; y += config.PeriodY {
		for x := 0; x < w; x += config.PeriodX {
			var options ebiten.DrawImageOptions
			options.Filter, options.Blend = config.TileFilter, config.TileBlend
			options.GeoM.Translate(float64(x), float64(y))
			surface.DrawImage(config.Tile, &options)
		}
	}
	return &CachedTileParallax{config: config, surface: surface}, nil
}

// Offset returns the final surface placement for a camera sample. Quantized
// mode truncates toward zero before wrapping, matching integer camera clocks.
func (parallax *CachedTileParallax) Offset(cameraX, cameraY float64) (x, y float64) {
	if parallax.config.ClampNegative {
		cameraX, cameraY = max(0, cameraX), max(0, cameraY)
	}
	x, y = cameraX/parallax.config.DivisorX, cameraY/parallax.config.DivisorY
	if parallax.config.PixelQuantize {
		x, y = math.Trunc(x), math.Trunc(y)
	}
	return -math.Mod(x, parallax.config.WrapX), -math.Mod(y, parallax.config.WrapY)
}

func (parallax *CachedTileParallax) DrawAt(dst *ebiten.Image, cameraX, cameraY float64) {
	if parallax == nil || parallax.surface == nil || dst == nil || math.IsNaN(cameraX) || math.IsNaN(cameraY) || math.IsInf(cameraX, 0) || math.IsInf(cameraY, 0) {
		return
	}
	x, y := parallax.Offset(cameraX, cameraY)
	var options ebiten.DrawImageOptions
	options.Filter, options.Blend = parallax.config.Filter, parallax.config.Blend
	options.GeoM.Translate(x, y)
	dst.DrawImage(parallax.surface, &options)
}

// Bounds reports the bounded cache size for resource inspection.
func (parallax *CachedTileParallax) Bounds() image.Rectangle {
	if parallax == nil || parallax.surface == nil {
		return image.Rectangle{}
	}
	return parallax.surface.Bounds()
}

func (parallax *CachedTileParallax) Close() error {
	if parallax != nil && parallax.surface != nil {
		parallax.surface.Deallocate()
		parallax.surface = nil
	}
	return nil
}
