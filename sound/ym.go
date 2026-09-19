package sound

import (
	"fmt"
	"io"

	"github.com/olivierh59500/ym-player/pkg/stsound"
)

type YMOptions struct {
	SampleRate int
	Loop       bool
}
type ymSynth struct {
	player  *stsound.StSound
	data    []byte
	options YMOptions
	mono    [blockFrames]int16
}

// NewYM loads YM data with ym-player. Zero SampleRate selects 48000 Hz.
// The decoder reports EOF at compute-block granularity; its final block can
// contain silence. Mono samples are duplicated into the two stereo channels.
func NewYM(data []byte, options YMOptions) (*Stream, error) {
	if options.SampleRate == 0 {
		options.SampleRate = 48000
	}
	if err := validRate(options.SampleRate); err != nil {
		return nil, err
	}
	y := &ymSynth{data: append([]byte(nil), data...), options: options}
	if err := y.reset(); err != nil {
		return nil, err
	}
	return newStream(y, options.SampleRate), nil
}
func (y *ymSynth) reset() error {
	if y.player != nil {
		y.player.Destroy()
	}
	y.player = stsound.CreateWithRate(y.options.SampleRate)
	if err := y.player.LoadMemory(y.data); err != nil {
		y.player.Destroy()
		y.player = nil
		return fmt.Errorf("sound: load YM: %w", err)
	}
	y.player.SetLoopMode(y.options.Loop)
	y.player.Play()
	return nil
}
func (y *ymSynth) render(dst []float32) (int, error) {
	frames := len(dst) / 2
	if !y.player.Compute(y.mono[:frames], frames) {
		return 0, io.EOF
	}
	for i, v := range y.mono[:frames] {
		sample := float32(v) / 32768
		dst[2*i] = sample
		dst[2*i+1] = sample
	}
	return frames * 2, nil
}
func (y *ymSynth) close() error {
	if y.player != nil {
		y.player.Destroy()
		y.player = nil
	}
	return nil
}
