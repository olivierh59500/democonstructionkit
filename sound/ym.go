package sound

import (
	"fmt"
	"io"
	"time"

	"github.com/olivierh59500/ym-player/pkg/stsound"
)

type YMOptions struct {
	SampleRate int
	Loop       bool
	Lowpass    *bool
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
	if options.Lowpass != nil {
		value := *options.Lowpass
		options.Lowpass = &value
	}
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
	s := newStream(y, options.SampleRate)
	info := y.player.GetInfo()
	s.metadata = Metadata{Format: FormatYM, Title: info.SongName, Author: info.SongAuthor, Comment: info.SongComment, Duration: time.Duration(info.MusicTimeInMs) * time.Millisecond}
	return s, nil
}

// YMRegisters returns a synchronized snapshot of the YM decoder's registers
// 0 through 13. The snapshot follows decoded audio, which may be ahead of audible
// playback because the stream and output device buffer samples. It does not
// advance playback. The result is false for non-YM or closed streams.
func (s *Stream) YMRegisters() ([14]uint8, bool) {
	var registers [14]uint8
	if s == nil {
		return registers, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	y, ok := s.synth.(*ymSynth)
	if !ok || y == nil || y.player == nil {
		return registers, false
	}
	for i := range registers {
		registers[i] = uint8(y.player.GetRegister(i))
	}
	return registers, true
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
	if y.options.Lowpass != nil {
		y.player.SetLowpassFilter(*y.options.Lowpass)
	}
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
