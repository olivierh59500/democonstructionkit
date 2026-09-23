// Package sound selects decoders and adapts music to stereo PCM. Opening a
// stream does not open an audio device or a graphics window.
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

// Stream is a synchronized, seekable stereo little-endian PCM stream.
// Position counts bytes delivered to the consumer, not decoded-ahead samples.
// Seek replays from the beginning for sample accuracy, so long seeks are costly.
// Device playback position is available through sound/ebiten.Player instead.
type Stream struct {
	mu          sync.Mutex
	synth       synthesizer
	rate        int
	samples     [blockFrames * 2]float32
	encoded     [blockFrames * frameBytes]byte
	begin, end  int
	position    int64
	pendingErr  error
	format      PCMFormat
	gain        float64
	quantize16  bool
	blockLimit  int
	metadata    Metadata
	totalFrames int64 // -1 when only approximate duration metadata is available.
}

func newStream(s synthesizer, rate int) *Stream {
	return &Stream{synth: s, rate: rate, gain: 1, blockLimit: blockFrames, totalFrames: -1}
}
func validRate(rate int) error {
	if rate < 8000 || rate > 192000 {
		return fmt.Errorf("sound: sample rate must be between 8000 and 192000 Hz")
	}
	return nil
}
func (s *Stream) SampleRate() int   { return s.rate }
func (s *Stream) Format() PCMFormat { return s.format }
func (s *Stream) bytesPerFrame() int64 {
	if s.format == PCM16 {
		return 4
	}
	return frameBytes
}
func (s *Stream) Metadata() Metadata { s.mu.Lock(); defer s.mu.Unlock(); return s.metadata }
func (s *Stream) lengthLocked() int64 {
	if s.totalFrames >= 0 {
		return s.totalFrames * s.bytesPerFrame()
	}
	d := s.metadata.Duration
	frames := int64(d/time.Second)*int64(s.rate) + int64(d%time.Second)*int64(s.rate)/int64(time.Second)
	return frames * s.bytesPerFrame()
}
func (s *Stream) Length() int64 { s.mu.Lock(); defer s.mu.Unlock(); return s.lengthLocked() }

// Volume is the stream gain before output-device gain. SetVolume clamps to
// [0,1] and re-encodes only unread complete frames, retaining partial-frame bytes.
func (s *Stream) Volume() float64 { s.mu.Lock(); defer s.mu.Unlock(); return s.gain }
func (s *Stream) SetVolume(value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if math.IsNaN(value) {
		value = 0
	}
	s.gain = max(0, min(1, value))
	frameSize := int(s.bytesPerFrame())
	firstFrame := (s.begin + frameSize - 1) / frameSize
	s.encodeSamples(firstFrame*2, s.end/(frameSize/2))
}
func (s *Stream) Position() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return time.Duration(float64(s.position) / float64(s.bytesPerFrame()*int64(s.rate)) * float64(time.Second))
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
			// Decode fixed-size blocks to reuse staging storage and amortize
			// synthesizer calls across arbitrarily small consumer reads.
			frames := s.blockLimit
			count, err := s.synth.render(s.samples[:frames*2])
			if count < 0 || count > frames*2 || count%2 != 0 {
				return n, fmt.Errorf("sound: invalid synthesizer sample count %d", count)
			}
			if count == 0 && err == nil {
				err = io.ErrNoProgress
			}
			s.begin, s.end, s.pendingErr = 0, count*int(s.bytesPerFrame()/2), err
			s.encodeSamples(0, count)
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

func (s *Stream) encodeSamples(first, end int) {
	for i := first; i < end; i++ {
		v := float64(s.samples[i]) * s.gain
		if s.format == PCM16 || s.quantize16 {
			integer := int16(max(-32768, min(32767, v*32768)))
			if s.format == PCM16 {
				binary.LittleEndian.PutUint16(s.encoded[i*2:], uint16(integer))
				continue
			}
			v = float64(integer) / 32768
		}
		binary.LittleEndian.PutUint32(s.encoded[i*4:], math.Float32bits(float32(v)))
	}
}

// Seek uses bytes in the selected output PCM format. SeekEnd requires known
// duration. Invalid offsets leave the stream unchanged.
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
	case io.SeekEnd:
		if s.metadata.Duration <= 0 {
			return s.position, fmt.Errorf("sound: duration unavailable for seek end")
		}
		length := s.lengthLocked()
		if offset > 0 && offset > math.MaxInt64-length || offset < 0 && offset < -length {
			return s.position, fmt.Errorf("sound: seek offset out of range")
		}
		target = length + offset
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
