// Package mobile exposes the same scene to a small native Android host.
package mobile

import (
	"github.com/hajimehoshi/ebiten/v2"
	enginemobile "github.com/hajimehoshi/ebiten/v2/mobile"
	"github.com/olivierh59500/democonstructionkit/examples/effectslab/scene"
	"sync"
)

var settings struct {
	sync.Mutex
	config scene.Config
}
var game *scene.Game

// Configure must be called by the host before it creates EbitenView. Profile
// paths are provided by Android's private files directory, never shared storage.
func Configure(eco bool, frames int64, profile string) {
	settings.Lock()
	defer settings.Unlock()
	settings.config = scene.Config{Eco: eco, Frames: int(frames), Profile: profile, ContinueAfterProfile: true}
}

// SetAuthoring selects the saved-project demonstration before creating EbitenView.
func SetAuthoring(enabled bool) {
	settings.Lock()
	defer settings.Unlock()
	settings.config.Authoring = enabled
}

type adapter struct{}

func (adapter) Update() error {
	if game == nil {
		settings.Lock()
		c := settings.config
		settings.Unlock()
		var err error
		game, err = scene.NewGame(c)
		if err != nil {
			return err
		}
	}
	return game.Update()
}
func (adapter) Draw(dst *ebiten.Image) {
	if game != nil {
		game.Draw(dst)
	}
}
func (adapter) Layout(w, h int) (int, int) {
	if game != nil {
		return game.Layout(w, h)
	}
	settings.Lock()
	eco := settings.config.Eco
	settings.Unlock()
	if eco {
		return 320, 180
	}
	return 640, 360
}
func init() { ebiten.SetTPS(60); enginemobile.SetGame(adapter{}) }
