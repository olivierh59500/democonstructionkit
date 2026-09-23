package sound

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/olivierh59500/go-zikmu"
)

// ModuleOptions controls the older explicit module constructor. Its Loop option
// retains the historical explicit-duration contract. Prefer Open, which selects
// the module decoder and supports natural song endings and automatic looping.
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
	natural         bool
}

// NewModule supports the formats detected by go-zikmu (MOD, XM, S3M and IT).
func NewModule(data []byte, options ModuleOptions) (*Stream, error) {
	return newModule(data, options, false)
}

// Open uses natural song endings where available; the older constructor retains
// its explicit-duration looping contract for existing applications.
func newAutoModule(data []byte, options ModuleOptions) (*Stream, error) {
	return newModule(data, options, true)
}
func newModule(data []byte, options ModuleOptions, natural bool) (*Stream, error) {
	if options.SampleRate == 0 {
		options.SampleRate = 48000
	}
	if err := validRate(options.SampleRate); err != nil {
		return nil, err
	}
	if options.Duration < 0 || (!natural && options.Loop && options.Duration == 0) {
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
		_ = player.Stop()
		return nil, fmt.Errorf("sound: duration shorter than one PCM frame")
	}
	s := newStream(&moduleSynth{player: player, limit: limit, loop: options.Loop, natural: natural}, options.SampleRate)
	s.metadata = Metadata{Format: Format(module.Metadata.Format), Title: module.Metadata.Title, Duration: options.Duration}
	if limit > 0 {
		s.totalFrames = limit
	}
	return s, nil
}
func (m *moduleSynth) render(dst []float32) (int, error) {
	n := 0
	restartedEmpty := false
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
		var count int
		var err error
		if m.natural {
			finite, ok := m.player.(interface{ RenderUntilEnd([]float32) (int, error) })
			if !ok {
				return n, fmt.Errorf("sound: module engine lacks natural-end playback")
			}
			count, err = finite.RenderUntilEnd(dst[n : n+frames*2])
		} else {
			count, err = m.player.Render(dst[n : n+frames*2])
		}
		n += count
		m.position += int64(count / 2)
		if err == io.EOF && m.loop {
			if count == 0 && restartedEmpty {
				return n, io.EOF
			}
			restartedEmpty = count == 0
			if err := m.reset(); err != nil {
				return n, err
			}
			continue
		}
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
