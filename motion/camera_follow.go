package motion

import "fmt"

// CameraFollowConfig anchors an object inside a viewport while clamping the
// camera at both world edges. Anchor is the object's top-left screen position
// during the central follow section, independent of sprite dimensions.
type CameraFollowConfig struct {
	ViewportW, ViewportH int
	WorldW, WorldH       int
	AnchorX, AnchorY     int
}

type CameraFollowPose struct {
	CameraX, CameraY int
	ScreenX, ScreenY int
}

// CameraFollow samples one 2D camera/object pose without retaining time or
// allocating. The caller chooses rounding of its own world coordinates.
type CameraFollow struct{ config CameraFollowConfig }

func NewCameraFollow(config CameraFollowConfig) (*CameraFollow, error) {
	if config.ViewportW < 1 || config.ViewportH < 1 || config.WorldW < config.ViewportW || config.WorldH < config.ViewportH ||
		config.AnchorX < 0 || config.AnchorY < 0 || config.AnchorX > config.ViewportW || config.AnchorY > config.ViewportH {
		return nil, fmt.Errorf("motion: invalid camera follow bounds")
	}
	return &CameraFollow{config: config}, nil
}

func followAxis(position, viewport, world, anchor int) (camera, screen int) {
	if position <= anchor {
		return 0, position
	}
	if position > world-(viewport-anchor) {
		camera = world - viewport
		return camera, position - camera
	}
	return position - anchor, anchor
}

func (follow *CameraFollow) At(worldX, worldY int) CameraFollowPose {
	c := follow.config
	cameraX, screenX := followAxis(worldX, c.ViewportW, c.WorldW, c.AnchorX)
	cameraY, screenY := followAxis(worldY, c.ViewportH, c.WorldH, c.AnchorY)
	return CameraFollowPose{CameraX: cameraX, CameraY: cameraY, ScreenX: screenX, ScreenY: screenY}
}
