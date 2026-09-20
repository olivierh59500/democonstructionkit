package composite

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image"
)

// Grid renders a camera window into any tile/sprite provider. Camera clamping and
// wrapping remain explicit application choices. Overscan preserves legacy edge
// behavior, including maps and fonts assembled from several adjacent tiles.
type Grid struct {
	Columns, Rows int
	Cell          image.Point
	Overscan      int
	Tile          func(x, y int) *ebiten.Image
}

func (g Grid) Draw(dst *ebiten.Image, camera, origin, view image.Point) {
	if g.Tile == nil || g.Cell.X <= 0 || g.Cell.Y <= 0 {
		return
	}
	sx, sy := camera.X/g.Cell.X, camera.Y/g.Cell.Y
	ox, oy := camera.X%g.Cell.X, camera.Y%g.Cell.Y
	for y := 0; y < view.Y/g.Cell.Y+g.Overscan; y++ {
		my := sy + y
		if my < 0 || my >= g.Rows {
			continue
		}
		for x := 0; x < view.X/g.Cell.X+g.Overscan; x++ {
			mx := sx + x
			if mx < 0 || mx >= g.Columns {
				continue
			}
			op := ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(origin.X+x*g.Cell.X-ox), float64(origin.Y+y*g.Cell.Y-oy))
			Instance{Image: g.Tile(mx, my), Options: op}.Draw(dst)
		}
	}
}
