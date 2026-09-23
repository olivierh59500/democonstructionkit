package sound

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/olivierh59500/go-zikmu"
	"github.com/olivierh59500/ym-player/pkg/lzh"
	"github.com/olivierh59500/ym-player/pkg/stsound"
)

type PCMFormat uint8

const (
	Float32 PCMFormat = iota
	PCM16
)

// Format identifies the source data, independently of the output PCM format.
type Format string

const (
	FormatYM  Format = "ym"
	FormatMOD Format = "mod"
	FormatXM  Format = "xm"
	FormatS3M Format = "s3m"
	FormatIT  Format = "it"
	FormatWAV Format = "wav"
	FormatMP3 Format = "mp3"
	FormatOgg Format = "ogg"
)

type ModuleProfile uint8

const (
	ModuleDefault ModuleProfile = iota
	ModuleProfileScreamTracker3
)

// Metadata contains decoder-independent information. Duration is zero when the
// replay's control flow has no known finite length. The original filename is not
// used as a title when the file supplies one.
type Metadata struct {
	Format                 Format
	Title, Author, Comment string
	Duration               time.Duration
}

// Options concerns playback, never decoder selection. Open recognizes the file
// content, with the name as a fallback. Zero values select 48000 Hz, float32,
// unit gain and 1024-frame decode blocks. Gain zero means the default; use
// SetVolume(0) to mute. PCM16 truncates after gain, and Quantize16 preserves that
// quantization when a migrated application consumes float32 PCM.
type Options struct {
	SampleRate     int
	PCMFormat      PCMFormat
	Gain           float64
	Quantize16     bool
	BlockFrames    int
	Loop           bool
	LoopStartFrame int64
	Duration       time.Duration
	Interpolation  bool
	Lowpass        *bool
	ModuleProfile  ModuleProfile
	StartOrder     int
	Track          int // Zero-based song in a supported multi-track container.
}

// Open selects the music decoder and returns one common PCM stream. No audio
// device is opened. A recognized signature takes precedence over a misleading
// extension. Packed YM and gzip-wrapped recordings are handled internally.
// The .fc extension identifies the packed S3M pattern convention used by the
// compatibility replay; ordinary .s3m files use the general module decoder.
func Open(name string, data []byte, c Options) (*Stream, error) {
	if c.SampleRate == 0 {
		c.SampleRate = 48000
	}
	if c.Gain == 0 {
		c.Gain = 1
	}
	if c.BlockFrames == 0 {
		c.BlockFrames = blockFrames
	}
	if err := validRate(c.SampleRate); err != nil {
		return nil, err
	}
	if c.PCMFormat > PCM16 || math.IsNaN(c.Gain) || math.IsInf(c.Gain, 0) || c.Gain < 0 || c.Gain > 1 || c.BlockFrames < 1 || c.BlockFrames > blockFrames || c.Duration < 0 || c.LoopStartFrame < 0 || c.ModuleProfile > ModuleProfileScreamTracker3 || c.StartOrder < 0 || c.Track < 0 {
		return nil, fmt.Errorf("sound: invalid playback options")
	}
	if len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b {
		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("sound: gzip: %w", err)
		}
		const maxExpanded = 256 << 20
		expanded, err := io.ReadAll(io.LimitReader(reader, maxExpanded+1))
		closeErr := reader.Close()
		if err != nil {
			return nil, fmt.Errorf("sound: gzip: %w", err)
		}
		if closeErr != nil {
			return nil, closeErr
		}
		if len(expanded) > maxExpanded {
			return nil, fmt.Errorf("sound: expanded recording exceeds limit")
		}
		data = expanded
		if strings.EqualFold(filepath.Ext(name), ".gz") {
			name = strings.TrimSuffix(name, filepath.Ext(name))
		}
	}
	format, err := Detect(name, data)
	if err != nil {
		return nil, err
	}
	if format == FormatS3M && strings.EqualFold(filepath.Ext(name), ".fc") && !(len(data) >= 48 && string(data[44:48]) == "SCRM") {
		var err error
		data, err = packedTrack(data, c.Track)
		if err != nil {
			return nil, err
		}
	} else if c.Track != 0 {
		return nil, fmt.Errorf("sound: track selection requires a supported music container")
	}
	var stream *Stream
	switch format {
	case FormatYM:
		if c.LoopStartFrame != 0 || c.StartOrder != 0 || c.ModuleProfile != ModuleDefault {
			return nil, fmt.Errorf("sound: module/recording options do not apply to YM")
		}
		stream, err = NewYM(data, YMOptions{SampleRate: c.SampleRate, Loop: c.Loop, Lowpass: c.Lowpass})
	case FormatMOD, FormatXM, FormatS3M, FormatIT:
		if c.LoopStartFrame != 0 {
			return nil, fmt.Errorf("sound: use tracker orders rather than PCM loop start for modules")
		}
		packed := format == FormatS3M && strings.EqualFold(filepath.Ext(name), ".fc")
		if packed || c.ModuleProfile == ModuleProfileScreamTracker3 {
			if format != FormatS3M {
				return nil, fmt.Errorf("sound: Scream Tracker profile requires S3M data")
			}
			stream, err = openST3(data, c, packed)
		} else {
			if c.StartOrder != 0 {
				return nil, fmt.Errorf("sound: start order requires the compatibility module profile")
			}
			stream, err = newAutoModule(data, ModuleOptions{SampleRate: c.SampleRate, Duration: c.Duration, Loop: c.Loop, Interpolation: c.Interpolation})
		}
	case FormatWAV, FormatMP3, FormatOgg:
		if c.StartOrder != 0 || c.ModuleProfile != ModuleDefault {
			return nil, fmt.Errorf("sound: tracker options do not apply to recordings")
		}
		var decoded interface {
			io.ReadSeeker
			Length() int64
		}
		switch format {
		case FormatWAV:
			decoded, err = wav.DecodeWithSampleRate(c.SampleRate, bytes.NewReader(data))
		case FormatMP3:
			decoded, err = mp3.DecodeWithSampleRate(c.SampleRate, bytes.NewReader(data))
		case FormatOgg:
			decoded, err = vorbis.DecodeWithSampleRate(c.SampleRate, bytes.NewReader(data))
		}
		if err == nil {
			stream, err = NewPCM16(decoded, PCM16Options{SampleRate: c.SampleRate, Loop: c.Loop, LoopStartFrame: c.LoopStartFrame})
			if err == nil {
				stream.metadata = Metadata{Format: format, Duration: time.Duration(decoded.Length()/4) * time.Second / time.Duration(c.SampleRate)}
				if decoded.Length() >= 0 {
					stream.totalFrames = decoded.Length() / 4
				}
			}
		}
	default:
		return nil, fmt.Errorf("sound: no decoder for format %q", format)
	}
	if err != nil {
		return nil, fmt.Errorf("sound: open %q: %w", name, err)
	}
	stream.format, stream.gain, stream.quantize16, stream.blockLimit = c.PCMFormat, c.Gain, c.Quantize16, c.BlockFrames
	if stream.metadata.Format == "" {
		stream.metadata.Format = format
	}
	return stream, nil
}

func packedTrack(container []byte, track int) ([]byte, error) {
	if track < 0 || track > 1 || len(container) < 8 {
		return nil, fmt.Errorf("sound: invalid packed soundtrack track/header")
	}
	start := uint64(binary.LittleEndian.Uint32(container[track*4:]))
	end := uint64(len(container))
	if track == 0 {
		end = uint64(binary.LittleEndian.Uint32(container[4:]))
	}
	if start < 8 || end < start || end-start < 96 || end > uint64(len(container)) {
		return nil, fmt.Errorf("sound: invalid packed soundtrack offsets")
	}
	data := container[start:end]
	if string(data[44:48]) != "SCRM" {
		return nil, fmt.Errorf("sound: packed soundtrack is missing an S3M header")
	}
	return data, nil
}

// OpenFS reads an embedded or filesystem asset and delegates all format handling
// to Open. The stream owns its decoder; callers own the filesystem itself.
func OpenFS(files fs.FS, name string, options Options) (*Stream, error) {
	if files == nil {
		return nil, fmt.Errorf("sound: nil asset filesystem")
	}
	data, err := fs.ReadFile(files, name)
	if err != nil {
		return nil, err
	}
	return Open(name, data, options)
}

// Detect recognizes supported signatures without opening a playback device.
// A recognized extension provides an error-producing decoder for malformed or
// signature-free inputs; it never turns an invalid file into a silent stream.
func Detect(name string, data []byte) (Format, error) {
	if stsound.IsYMFile(data) || lzh.IsLZHCompressed(data) {
		return FormatYM, nil
	}
	if len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WAVE" {
		return FormatWAV, nil
	}
	if len(data) >= 4 && string(data[:4]) == "OggS" {
		return FormatOgg, nil
	}
	if len(data) > 0 {
		f, err := zikmu.Detect(bytes.NewReader(data), int64(len(data)))
		if err == nil && f != zikmu.FormatUnknown {
			return Format(f), nil
		}
	}
	if len(data) >= 3 && string(data[:3]) == "ID3" || len(data) >= 2 && data[0] == 0xff && data[1]&0xe0 == 0xe0 {
		return FormatMP3, nil
	}
	switch strings.ToLower(filepath.Ext(name)) {
	case ".ym", ".ymt", ".mix":
		return FormatYM, nil
	case ".mod":
		return FormatMOD, nil
	case ".xm":
		return FormatXM, nil
	case ".s3m", ".fc":
		return FormatS3M, nil
	case ".it":
		return FormatIT, nil
	case ".wav":
		return FormatWAV, nil
	case ".mp3":
		return FormatMP3, nil
	case ".ogg":
		return FormatOgg, nil
	}
	return "", fmt.Errorf("sound: unsupported music format %q", name)
}
