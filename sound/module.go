package sound

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/olivierh59500/go-zikmu"
)

// ModuleOptions controls go-zikmu playback. Duration is an optional explicit
// endpoint; Loop requires it because go-zikmu does not expose song-end markers.
// Without Duration the stream continues until closed, following the replay engine.
type ModuleOptions struct {
	SampleRate    int
	Duration      time.Duration
	Loop          bool
	Interpolation bool
}
type moduleSynth struct {
	player          zikmu.Player
	limit, position int64
	loop            bool
}

// NewModule supports the formats detected by go-zikmu (MOD, XM, S3M and IT).
func NewModule(data []byte, options ModuleOptions) (*Stream, error) {
	if options.SampleRate == 0 {
		options.SampleRate = 48000
	}
	if err := validRate(options.SampleRate); err != nil {
		return nil, err
	}
	if options.Duration < 0 || (options.Loop && options.Duration == 0) {
		return nil, fmt.Errorf("sound: module looping requires a positive explicit duration")
	}
	module, err := zikmu.Load(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("sound: load module: %w", err)
	}
	config := zikmu.DefaultConfig()
	config.SampleRate = options.SampleRate
	config.Channels = 2
	config.BufferSamples = blockFrames * 2
	config.Interpolation = options.Interpolation
	player, err := zikmu.NewPlayer(module, config)
	if err != nil {
		return nil, err
	}
	// Split seconds and fractional seconds to avoid overflowing duration*rate.
	limit := int64(options.Duration/time.Second)*int64(options.SampleRate) + int64(options.Duration%time.Second)*int64(options.SampleRate)/int64(time.Second)
	if options.Duration > 0 && limit == 0 {
		return nil, fmt.Errorf("sound: duration shorter than one PCM frame")
	}
	return newStream(&moduleSynth{player: player, limit: limit, loop: options.Loop}, options.SampleRate), nil
}
func (m *moduleSynth) render(dst []float32) (int, error) {
	n := 0
	for n < len(dst) {
		if m.limit > 0 && m.position >= m.limit {
			if !m.loop {
				return n, io.EOF
			}
			if err := m.reset(); err != nil {
				return n, err
			}
		}
		frames := (len(dst) - n) / 2
		if m.limit > 0 {
			frames = int(min(int64(frames), m.limit-m.position))
		}
		count, err := m.player.Render(dst[n : n+frames*2])
		n += count
		m.position += int64(count / 2)
		if err != nil {
			return n, err
		}
		if count == 0 {
			return n, io.ErrNoProgress
		}
	}
	return n, nil
}
func (m *moduleSynth) reset() error {
	if err := m.player.Reset(); err != nil {
		return err
	}
	m.player.Play()
	m.position = 0
	return nil
}
func (m *moduleSynth) close() error { return m.player.Stop() }
