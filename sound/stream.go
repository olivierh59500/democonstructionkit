// Package sound adapts YM and tracker modules to stereo float32 PCM without
// importing a windowing or audio-device backend.
package sound

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"sync"
	"time"
)

const blockFrames = 1024
const frameBytes = 8

type synthesizer interface {
	render([]float32) (int, error)
	reset() error
	close() error
}

// Stream is a synchronized, seekable stereo float32 little-endian PCM stream.
// Position counts bytes delivered to the consumer, not decoded-ahead samples.
// Seek replays from the beginning for sample accuracy, so long seeks are costly.
// Device playback position is available through sound/ebiten.Player instead.
type Stream struct {
	mu         sync.Mutex
	synth      synthesizer
	rate       int
	samples    [blockFrames * 2]float32
	encoded    [blockFrames * frameBytes]byte
	begin, end int
	position   int64
	pendingErr error
}

func newStream(s synthesizer, rate int) *Stream { return &Stream{synth: s, rate: rate} }
func validRate(rate int) error {
	if rate < 8000 || rate > 192000 {
		return fmt.Errorf("sound: sample rate must be between 8000 and 192000 Hz")
	}
	return nil
}
func (s *Stream) SampleRate() int { return s.rate }
func (s *Stream) Position() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return time.Duration(float64(s.position) / float64(frameBytes*s.rate) * float64(time.Second))
}

// Read accepts arbitrary buffer lengths, including partial PCM samples.
func (s *Stream) Read(dst []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readLocked(dst)
}
func (s *Stream) readLocked(dst []byte) (int, error) {
	if len(dst) == 0 {
		return 0, nil
	}
	if s.synth == nil {
		return 0, io.ErrClosedPipe
	}
	n := 0
	for n < len(dst) {
		if s.begin == s.end {
			if s.pendingErr != nil {
				return n, s.pendingErr
			}
			// Keep decoder call boundaries independent of the consumer's read sizes.
			// Some YM register updates depend on Compute's block boundaries.
			frames := blockFrames
			count, err := s.synth.render(s.samples[:frames*2])
			if count < 0 || count > frames*2 || count%2 != 0 {
				return n, fmt.Errorf("sound: invalid synthesizer sample count %d", count)
			}
			if count == 0 && err == nil {
				err = io.ErrNoProgress
			}
			s.begin, s.end, s.pendingErr = 0, count*4, err
			for i, v := range s.samples[:count] {
				binary.LittleEndian.PutUint32(s.encoded[i*4:], math.Float32bits(v))
			}
			if count == 0 {
				return n, err
			}
		}
		copied := copy(dst[n:], s.encoded[s.begin:s.end])
		s.begin += copied
		n += copied
		s.position += int64(copied)
	}
	return n, nil
}

// Seek uses PCM byte offsets. SeekEnd is unsupported because duration may depend
// on tracker control flow. Invalid offsets leave the stream unchanged.
func (s *Stream) Seek(offset int64, whence int) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.synth == nil {
		return 0, io.ErrClosedPipe
	}
	target := offset
	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		if (offset > 0 && s.position > math.MaxInt64-offset) || (offset < 0 && offset < -s.position) {
			return s.position, fmt.Errorf("sound: seek offset out of range")
		}
		target = s.position + offset
	default:
		return s.position, fmt.Errorf("sound: unsupported seek origin")
	}
	if target < 0 {
		return s.position, fmt.Errorf("sound: negative seek offset")
	}
	if target == s.position {
		return target, nil
	}
	if target < s.position {
		if err := s.synth.reset(); err != nil {
			return s.position, err
		}
		s.position = 0
		s.begin = 0
		s.end = 0
		s.pendingErr = nil
	}
	// Separate discard storage prevents aliasing the encoded staging buffer.
	var discard [blockFrames * frameBytes]byte
	for s.position < target {
		size := min(int64(len(discard)), target-s.position)
		_, err := s.readLocked(discard[:size])
		if err != nil {
			return s.position, err
		}
	}
	return s.position, nil
}

// Close is idempotent and serialized with audio callbacks.
func (s *Stream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.synth == nil {
		return nil
	}
	err := s.synth.close()
	s.synth = nil
	s.begin = 0
	s.end = 0
	return err
}
