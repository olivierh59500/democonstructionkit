package sound

import (
	"fmt"
	"io"
	"time"

	"github.com/olivierh59500/go-zikmu"
)

// TrackerPosition identifies a tick in the soundtrack's order and pattern row.
// PlusFlags describes the signed distance to an order-separator transition.
type TrackerPosition struct {
	Order, Row uint16
	Frame      uint32
	PlusFlags  int16
}

type st3Synth struct {
	player                 *zikmu.ScreamTracker3
	data                   []byte
	options                zikmu.ScreamTracker3Options
	pcm                    [blockFrames * 2]int16
	limit, position, cycle int64
	loop                   bool
}

// openST3 is selected by Open for FC data or an explicit compatibility profile.
// All tracker decoding remains inside go-zikmu.
func openST3(data []byte, options Options, packed bool) (*Stream, error) {
	s := &st3Synth{data: append([]byte(nil), data...), options: zikmu.ScreamTracker3Options{
		SampleRate: options.SampleRate, Interpolation: options.Interpolation,
		StartOrder: options.StartOrder, PackedPatterns: packed,
	}}
	s.limit = int64(options.Duration/time.Second)*int64(options.SampleRate) + int64(options.Duration%time.Second)*int64(options.SampleRate)/int64(time.Second)
	s.loop = options.Loop
	if options.Duration > 0 && s.limit == 0 {
		return nil, fmt.Errorf("sound: duration shorter than one PCM frame")
	}
	if err := s.reset(); err != nil {
		return nil, err
	}
	stream := newStream(s, options.SampleRate)
	stream.metadata = Metadata{Format: FormatS3M, Title: s.player.Title(), Duration: options.Duration}
	return stream, nil
}
func (s *st3Synth) reset() error {
	if s.player != nil {
		s.player.Close()
	}
	var err error
	s.player, err = zikmu.NewScreamTracker3(s.data, s.options)
	s.position, s.cycle = 0, 0
	return err
}
func (s *st3Synth) render(dst []float32) (int, error) {
	n := 0
	for n < len(dst) {
		if s.limit > 0 && s.position == s.limit {
			if !s.loop {
				return n, io.EOF
			}
			cycle := s.cycle + 1
			if err := s.reset(); err != nil {
				return n, err
			}
			s.cycle = cycle
		}
		frames := (len(dst) - n) / 2
		if s.limit > 0 {
			frames = int(min(int64(frames), s.limit-s.position))
		}
		count := s.player.Fill(s.pcm[:frames*2])
		for i, value := range s.pcm[:count*2] {
			dst[n+i] = float32(value) / 32768
		}
		n += count * 2
		s.position += int64(count)
		if count != frames {
			return n, io.EOF
		}
	}
	return n, nil
}
func (s *st3Synth) close() error { return s.player.Close() }

// TrackerPositionAt maps the output player's audible position to musical
// markers, compensating for decoder and device buffering. Marker history is
// bounded; older requests clamp to the oldest retained marker. This capability
// is available on timing-compatible S3M/FC streams and returns false otherwise.
// An explicit-duration loop discards markers from previous cycles.
func (s *Stream) TrackerPositionAt(position time.Duration) (TrackerPosition, bool) {
	if s == nil {
		return TrackerPosition{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	synth, ok := s.synth.(*st3Synth)
	if !ok || synth.player == nil {
		return TrackerPosition{}, false
	}
	position = max(position, 0)
	frames := uint64(position/time.Second)*uint64(s.rate) + uint64(position%time.Second)*uint64(s.rate)/uint64(time.Second)
	if synth.limit > 0 && synth.loop {
		if frames/uint64(synth.limit) != uint64(synth.cycle) {
			return TrackerPosition{}, false
		}
		frames %= uint64(synth.limit)
	}
	snapshot := synth.player.PositionAt(frames)
	return TrackerPosition{Order: snapshot.Order, Row: snapshot.Row, Frame: snapshot.Frame, PlusFlags: snapshot.PlusFlags}, true
}
