package sound

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// PCM16Options describes a decoded, interleaved, little-endian stereo stream.
// This adapter lets recorded sound coexist with YM/modules on the same player.
type PCM16Options struct {
	SampleRate int
	Loop       bool
	// LoopStartFrame skips a one-time introduction on subsequent loops. The
	// position counts decoded stereo frames at SampleRate, not source bytes.
	// It requires Loop; zero repeats the entire stream. Seeking to the beginning
	// always replays the introduction. An empty loop region ends playback.
	LoopStartFrame int64
}
type pcm16 struct {
	source    io.ReadSeeker
	loop      bool
	loopStart int64
	buffer    [blockFrames * 4]byte
}

// NewPCM16 owns a seekable decoder on success; no device backend is imported.
// Closing the returned stream also closes the decoder when it implements Closer.
func NewPCM16(source io.ReadSeeker, c PCM16Options) (*Stream, error) {
	if source == nil {
		return nil, fmt.Errorf("sound: nil PCM decoder")
	}
	if c.SampleRate == 0 {
		c.SampleRate = 48000
	}
	if err := validRate(c.SampleRate); err != nil {
		return nil, err
	}
	if c.LoopStartFrame < 0 || c.LoopStartFrame > math.MaxInt64/4 {
		return nil, fmt.Errorf("sound: PCM loop start frame out of range")
	}
	if c.LoopStartFrame != 0 && !c.Loop {
		return nil, fmt.Errorf("sound: PCM loop start requires looping")
	}
	return newStream(&pcm16{source: source, loop: c.Loop, loopStart: c.LoopStartFrame * 4}, c.SampleRate), nil
}
func (p *pcm16) render(dst []float32) (int, error) {
	written := 0
	empty := false
	for written < len(dst) {
		want := min((len(dst)-written)*2, len(p.buffer))
		want -= want % 4
		if want == 0 {
			return written, fmt.Errorf("sound: PCM destination is not stereo aligned")
		}
		n, err := io.ReadFull(p.source, p.buffer[:want])
		if n%4 != 0 {
			return written, fmt.Errorf("sound: truncated PCM frame")
		}
		for i := 0; i < n; i += 2 {
			dst[written] = float32(int16(binary.LittleEndian.Uint16(p.buffer[i:i+2]))) / 32768
			written++
		}
		if err != nil {
			if err != io.EOF && err != io.ErrUnexpectedEOF {
				return written, err
			}
			if !p.loop {
				return written, io.EOF
			}
			if n == 0 && empty {
				return written, io.EOF
			}
			empty = n == 0
			if _, err = p.source.Seek(p.loopStart, io.SeekStart); err != nil {
				return written, err
			}
		}
	}
	return written, nil
}
func (p *pcm16) reset() error { _, err := p.source.Seek(0, io.SeekStart); return err }
func (p *pcm16) close() error {
	if c, ok := p.source.(io.Closer); ok {
		return c.Close()
	}
	return nil
}
