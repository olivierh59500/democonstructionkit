// Package output routes PCM either to Ebitengine's device or to an offline mix.
// Device playback is the default. A recording session must begin before the
// game creates its context; no system audio or unrelated applications are read.
package output

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

var active atomic.Pointer[Session]

// Now follows the recording clock during export and the wall clock otherwise.
func Now() time.Time {
	if s := active.Load(); s != nil {
		return time.Unix(946684800, s.elapsed.Load())
	}
	return time.Now()
}

type Context struct {
	device *audio.Context
	mix    *Session
	rate   int
}

func NewContext(sampleRate int) *Context {
	if s := active.Load(); s != nil {
		return &Context{mix: s, rate: sampleRate}
	}
	return &Context{device: audio.NewContext(sampleRate), rate: sampleRate}
}

func CurrentContext() *Context {
	if s := active.Load(); s != nil {
		return &Context{mix: s, rate: s.rate}
	}
	if c := audio.CurrentContext(); c != nil {
		return &Context{device: c, rate: c.SampleRate()}
	}
	return nil
}

func (c *Context) SampleRate() int                             { return c.rate }
func (c *Context) IsReady() bool                               { return c.mix != nil || c.device.IsReady() }
func (c *Context) NewPlayer(src io.Reader) (*Player, error)    { return c.newPlayer(src, false) }
func (c *Context) NewPlayerF32(src io.Reader) (*Player, error) { return c.newPlayer(src, true) }

func (c *Context) newPlayer(src io.Reader, floating bool) (*Player, error) {
	if src == nil {
		return nil, fmt.Errorf("output: nil PCM source")
	}
	p := &Player{source: src, mix: c.mix, rate: c.rate, floating: floating, volume: 1}
	if c.mix != nil {
		if c.rate != c.mix.rate {
			return nil, fmt.Errorf("output: context rate %d differs from recording rate %d", c.rate, c.mix.rate)
		}
		c.mix.mu.Lock()
		defer c.mix.mu.Unlock()
		if c.mix.closed {
			return nil, io.ErrClosedPipe
		}
		c.mix.players = append(c.mix.players, p)
		return p, nil
	}
	var err error
	if floating {
		p.device, err = c.device.NewPlayerF32(src)
	} else {
		p.device, err = c.device.NewPlayer(src)
	}
	return p, err
}

// Player mirrors the device controls. Like audio.Player, Close does not close
// the caller-owned source. Offline position counts consumed PCM sample frames.
type Player struct {
	device                    *audio.Player
	mix                       *Session
	source                    io.Reader
	rate                      int
	floating, playing, closed bool
	volume                    float64
	frames                    int64
	buffer                    []byte
}

func (p *Player) Play() {
	if p.device != nil {
		p.device.Play()
		return
	}
	p.mix.mu.Lock()
	defer p.mix.mu.Unlock()
	if !p.closed {
		p.playing = true
	}
}
func (p *Player) Pause() {
	if p.device != nil {
		p.device.Pause()
		return
	}
	p.mix.mu.Lock()
	defer p.mix.mu.Unlock()
	p.playing = false
}
func (p *Player) IsPlaying() bool {
	if p.device != nil {
		return p.device.IsPlaying()
	}
	p.mix.mu.Lock()
	defer p.mix.mu.Unlock()
	return p.playing && !p.closed
}
func (p *Player) Position() time.Duration {
	if p.device != nil {
		return p.device.Position()
	}
	p.mix.mu.Lock()
	defer p.mix.mu.Unlock()
	return time.Duration(p.frames * int64(time.Second) / int64(p.rate))
}
func (p *Player) Volume() float64 {
	if p.device != nil {
		return p.device.Volume()
	}
	p.mix.mu.Lock()
	defer p.mix.mu.Unlock()
	return p.volume
}
func (p *Player) SetVolume(v float64) {
	if p.device != nil {
		p.device.SetVolume(v)
		return
	}
	if math.IsNaN(v) || v < 0 || v > 1 {
		panic("output: volume outside [0, 1]")
	}
	p.mix.mu.Lock()
	defer p.mix.mu.Unlock()
	p.volume = v
}
func (p *Player) SetBufferSize(d time.Duration) {
	if p.device != nil {
		p.device.SetBufferSize(d)
	}
}
func (p *Player) Rewind() error { return p.SetPosition(0) }
func (p *Player) SetPosition(d time.Duration) error {
	if p.device != nil {
		return p.device.SetPosition(d)
	}
	p.mix.mu.Lock()
	defer p.mix.mu.Unlock()
	if p.closed {
		return io.ErrClosedPipe
	}
	if d < 0 {
		return fmt.Errorf("output: negative position")
	}
	seeker, ok := p.source.(io.Seeker)
	if !ok {
		return fmt.Errorf("output: source is not seekable")
	}
	frames := int64(d/time.Second)*int64(p.rate) + int64(d%time.Second)*int64(p.rate)/int64(time.Second)
	size := int64(4)
	if p.floating {
		size = 8
	}
	if _, err := seeker.Seek(frames*size, io.SeekStart); err != nil {
		return err
	}
	p.frames = frames
	return nil
}
func (p *Player) Close() error {
	if p.device != nil {
		return p.device.Close()
	}
	p.mix.mu.Lock()
	defer p.mix.mu.Unlock()
	p.closed, p.playing = true, false
	for i, other := range p.mix.players {
		if other == p {
			p.mix.players = append(p.mix.players[:i], p.mix.players[i+1:]...)
			break
		}
	}
	p.source, p.buffer = nil, nil
	return nil
}

// Session mixes stereo float32 little-endian PCM at exact simulation ticks.
// Advance writes silence when no player is active, preserving delayed cues.
type Session struct {
	mu      sync.Mutex
	rate    int
	players []*Player
	frames  int64
	elapsed atomic.Int64
	closed  bool
	samples []float32
	encoded []byte
}

func Begin(sampleRate int) (*Session, error) {
	if sampleRate < 8000 || sampleRate > 192000 {
		return nil, fmt.Errorf("output: invalid sample rate")
	}
	s := &Session{rate: sampleRate}
	if !active.CompareAndSwap(nil, s) {
		return nil, fmt.Errorf("output: a recording session is already active")
	}
	return s, nil
}

func (s *Session) Advance(tick int64, ticksPerSecond int, dst io.Writer) error {
	if tick < 0 || ticksPerSecond <= 0 {
		return fmt.Errorf("output: invalid clock")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return io.ErrClosedPipe
	}
	target := tick * int64(s.rate) / int64(ticksPerSecond)
	if target < s.frames {
		return fmt.Errorf("output: recording clock moved backwards")
	}
	count := int(target - s.frames)
	if cap(s.samples) < count*2 {
		s.samples = make([]float32, count*2)
	}
	s.samples = s.samples[:count*2]
	clear(s.samples)
	for _, p := range s.players {
		if !p.playing || p.closed {
			continue
		}
		size := 4
		if p.floating {
			size = 8
		}
		if cap(p.buffer) < count*size {
			p.buffer = make([]byte, count*size)
		}
		p.buffer = p.buffer[:count*size]
		n, err := io.ReadFull(p.source, p.buffer)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return fmt.Errorf("output: read PCM: %w", err)
		}
		if n%size != 0 {
			return fmt.Errorf("output: incomplete stereo sample frame")
		}
		if err != nil {
			p.playing = false
		}
		p.frames += int64(n / size)
		for i := 0; i < n/(size/2); i++ {
			var v float32
			if p.floating {
				v = math.Float32frombits(binary.LittleEndian.Uint32(p.buffer[i*4:]))
			} else {
				v = float32(int16(binary.LittleEndian.Uint16(p.buffer[i*2:]))) / 32768
			}
			s.samples[i] += v * float32(p.volume)
		}
	}
	if cap(s.encoded) < count*8 {
		s.encoded = make([]byte, count*8)
	}
	s.encoded = s.encoded[:count*8]
	for i, v := range s.samples {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return fmt.Errorf("output: non-finite PCM sample")
		}
		binary.LittleEndian.PutUint32(s.encoded[i*4:], math.Float32bits(max(-1, min(1, v))))
	}
	if n, err := dst.Write(s.encoded); err != nil {
		return err
	} else if n != len(s.encoded) {
		return io.ErrShortWrite
	}
	s.frames = target
	s.elapsed.Store(tick * int64(time.Second) / int64(ticksPerSecond))
	return nil
}

func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	for _, p := range s.players {
		p.playing = false
		p.closed = true
	}
	s.players = nil
	active.CompareAndSwap(s, nil)
	return nil
}
