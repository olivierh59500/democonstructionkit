// Package ebiten connects sound.Stream to an existing Ebitengine audio context.
package ebiten

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/olivierh59500/democonstructionkit/sound"
)

// Player owns its stream on success. Control methods run on the game goroutine;
// Stream serializes the concurrent audio callback. Close is idempotent.
type Player struct {
	source *sound.Stream
	device *audio.Player
	closed bool
}

func NewPlayer(context *audio.Context, stream *sound.Stream) (*Player, error) {
	if context == nil || stream == nil {
		return nil, fmt.Errorf("sound: nil context or stream")
	}
	if context.SampleRate() != stream.SampleRate() {
		return nil, fmt.Errorf("sound: context and stream sample rates differ")
	}
	p, err := context.NewPlayerF32(stream)
	if err != nil {
		return nil, err
	}
	return &Player{source: stream, device: p}, nil
}
func (p *Player) Play() {
	if !p.closed {
		p.device.Play()
	}
}
func (p *Player) Pause() {
	if !p.closed {
		p.device.Pause()
	}
}
func (p *Player) IsPlaying() bool { return !p.closed && p.device.IsPlaying() }

// Position reports audible progress, unlike the decoder's buffered position.
func (p *Player) Position() time.Duration {
	if p.closed {
		return 0
	}
	return p.device.Position()
}
func (p *Player) SetVolume(volume float64) error {
	if p.closed {
		return fmt.Errorf("sound: player closed")
	}
	if math.IsNaN(volume) || math.IsInf(volume, 0) {
		return fmt.Errorf("sound: invalid volume")
	}
	p.device.SetVolume(max(0, min(1, volume)))
	return nil
}
func (p *Player) Seek(position time.Duration) error {
	if p.closed {
		return fmt.Errorf("sound: player closed")
	}
	if position < 0 {
		return fmt.Errorf("sound: negative position")
	}
	// SetPosition calls the stream's Seek and resets Ebitengine's buffered samples.
	return p.device.SetPosition(position)
}
func (p *Player) Close() error {
	if p.closed {
		return nil
	}
	p.closed = true
	p.device.Pause()
	return errors.Join(p.device.Close(), p.source.Close())
}
