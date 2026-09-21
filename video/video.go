// Package video exports the game canvas and its own PCM audio to a web-ready
// MP4. Simulation and audio use a shared clock; export speed never changes the
// animation speed. FFmpeg is only used for H.264/AAC encoding and muxing.
package video

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/sound/output"
)

type Config struct {
	Output               string
	Title                string
	Width, Height        int
	FPS, TPS, SampleRate int
	Duration             time.Duration // Zero records until the game returns ebiten.Termination.
	CRF                  int
	Preset               string
	PosterAt             time.Duration
}

// Flags adds the shared export options to a command's flag set.
func (c *Config) Flags(fs *flag.FlagSet) {
	if c.PosterAt == 0 {
		c.PosterAt = 15 * time.Second
	}
	fs.StringVar(&c.Output, "output", c.Output, "output MP4 path (must not already exist)")
	fs.DurationVar(&c.Duration, "duration", c.Duration, "recording duration; 0 records the complete production")
	fs.IntVar(&c.CRF, "crf", 17, "H.264 quality (0-51; lower is better)")
	fs.StringVar(&c.Preset, "preset", "fast", "FFmpeg H.264 encoding preset")
	fs.DurationVar(&c.PosterAt, "poster-at", c.PosterAt, "time of the PNG poster frame")
}

type Chapter struct {
	Title string  `json:"title"`
	Start float64 `json:"start_seconds"`
}
type Report struct {
	Title      string    `json:"title"`
	File       string    `json:"file"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	FPS        int       `json:"fps"`
	SampleRate int       `json:"sample_rate"`
	Frames     int64     `json:"frames"`
	Duration   float64   `json:"duration_seconds"`
	Finished   bool      `json:"production_finished"`
	Chapters   []Chapter `json:"chapters,omitempty"`
}

// Run constructs the unmodified game inside the engine, after activating the
// offline audio route. The factory may return a game with Cleanup or Close.
// Each simulated tick is drawn, including ticks omitted from a lower FPS video.
func Run(c Config, factory func() (ebiten.Game, error)) error {
	if c.FPS == 0 {
		c.FPS = 60
	}
	if c.TPS == 0 {
		c.TPS = c.FPS
	}
	if c.SampleRate == 0 {
		c.SampleRate = 48000
	}
	if c.Preset == "" {
		c.Preset = "fast"
	}
	if c.PosterAt == 0 {
		c.PosterAt = 15 * time.Second
	}
	if factory == nil || c.Output == "" || c.Width <= 0 || c.Height <= 0 || c.Width%2 != 0 || c.Height%2 != 0 || c.FPS <= 0 || c.TPS < c.FPS || c.Duration < 0 || c.PosterAt < 0 || c.CRF < 0 || c.CRF > 51 {
		return fmt.Errorf("video: invalid output, dimensions, rates, duration or quality")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("video: FFmpeg is required: %w", err)
	}
	if _, err := os.Stat(c.Output); !errors.Is(err, os.ErrNotExist) {
		if err != nil {
			return err
		}
		return fmt.Errorf("video: output already exists: %s", c.Output)
	}
	if err := os.MkdirAll(filepath.Dir(c.Output), 0755); err != nil {
		return err
	}
	work, err := os.MkdirTemp(filepath.Dir(c.Output), ".video-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)
	pcm, err := os.Create(filepath.Join(work, "audio.f32"))
	if err != nil {
		return err
	}
	defer pcm.Close()
	mix, err := output.Begin(c.SampleRate)
	if err != nil {
		return err
	}
	defer mix.Close()

	encoded := filepath.Join(work, "video.mp4")
	cmd := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error", "-nostdin", "-f", "rawvideo", "-pixel_format", "rgba", "-video_size", fmt.Sprintf("%dx%d", c.Width, c.Height), "-framerate", fmt.Sprint(c.FPS), "-i", "pipe:0", "-an", "-c:v", "libx264", "-preset", c.Preset, "-crf", fmt.Sprint(c.CRF), "-threads", "4", "-pix_fmt", "yuv420p", "-y", encoded)
	cmd.Stderr = os.Stderr
	pipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		pipe.Close()
		return err
	}
	g := &recorder{config: c, factory: factory, mix: mix, pcm: pcm, pipe: pipe, lastProgress: time.Now()}
	ebiten.SetWindowSize(min(c.Width, 960), min(c.Height, 600))
	ebiten.SetWindowTitle("Video export: " + c.Title)
	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetVsyncEnabled(false)
	ebiten.SetScreenClearedEveryFrame(true)
	ebiten.SetTPS(ebiten.SyncWithFPS)
	runErr := ebiten.RunGame(g)
	closeErr := pipe.Close()
	encodeErr := cmd.Wait()
	if g.game != nil {
		switch game := g.game.(type) {
		case interface{ Close() error }:
			runErr = errors.Join(runErr, game.Close())
		case interface{ Close() }:
			game.Close()
		case interface{ Cleanup() }:
			game.Cleanup()
		}
	}
	if err = errors.Join(runErr, g.err, closeErr, encodeErr); err != nil {
		return err
	}
	if !g.done || g.frames == 0 {
		return fmt.Errorf("video: recording interrupted before completion")
	}
	if int64(c.PosterAt)*int64(c.FPS)/int64(time.Second) >= g.frames {
		g.poster = image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))
		copy(g.poster.Pix, g.pixels)
	}
	if err = pcm.Close(); err != nil {
		return err
	}
	duration := float64(g.frames) / float64(c.FPS)
	metadata := filepath.Join(work, "chapters.txt")
	if err = writeChapters(metadata, g.chapters, duration); err != nil {
		return err
	}
	final := filepath.Join(work, "final.mp4")
	cmd = exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error", "-nostdin", "-i", encoded, "-f", "f32le", "-ar", fmt.Sprint(c.SampleRate), "-ac", "2", "-i", pcm.Name(), "-f", "ffmetadata", "-i", metadata, "-map", "0:v:0", "-map", "1:a:0", "-map_metadata", "2", "-map_chapters", "2", "-metadata", "title="+c.Title, "-c:v", "copy", "-c:a", "aac", "-b:a", "192k", "-t", fmt.Sprintf("%.9f", duration), "-movie_timescale", "1000", "-movflags", "+faststart", "-y", final)
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("video: mux audio: %w", err)
	}
	// Link instead of Rename prevents an existing output from being overwritten.
	if err = os.Link(final, c.Output); err != nil {
		return err
	}
	report := Report{Title: c.Title, File: filepath.Base(c.Output), Width: c.Width, Height: c.Height, FPS: c.FPS, SampleRate: c.SampleRate, Frames: g.frames, Duration: duration, Finished: g.finished, Chapters: g.chapters}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	base := strings.TrimSuffix(c.Output, filepath.Ext(c.Output))
	if err = os.WriteFile(base+".json", append(data, '\n'), 0644); err != nil {
		return err
	}
	if g.poster != nil {
		f, err := os.Create(base + ".png")
		if err != nil {
			return err
		}
		err = png.Encode(f, g.poster)
		err = errors.Join(err, f.Close())
		if err != nil {
			return err
		}
	}
	fmt.Fprintf(os.Stderr, "%s: exported %.2fs, %d frames, %dx%d, %d FPS\n", c.Output, duration, g.frames, c.Width, c.Height, c.FPS)
	return nil
}

type recorder struct {
	config          Config
	factory         func() (ebiten.Game, error)
	game            ebiten.Game
	mix             *output.Session
	pcm, pipe       io.Writer
	surface, canvas *ebiten.Image
	pixels          []byte
	poster          *image.RGBA
	tick, frames    int64
	done, finished  bool
	err             error
	chapters        []Chapter
	lastProgress    time.Time
}

func (r *recorder) Layout(int, int) (int, int) { return r.config.Width, r.config.Height }
func (r *recorder) Update() error {
	if r.done {
		return ebiten.Termination
	}
	if r.game == nil {
		var err error
		r.game, err = r.factory()
		if err != nil {
			return err
		}
		// A game factory may configure its interactive TPS.
		ebiten.SetTPS(ebiten.SyncWithFPS)
		r.canvas = ebiten.NewImage(r.config.Width, r.config.Height)
		r.pixels = make([]byte, r.config.Width*r.config.Height*4)
	}
	return nil
}

func (r *recorder) Draw(screen *ebiten.Image) {
	if r.game == nil || r.done {
		return
	}
	// ReadPixels flushes each exported frame, bounding GPU command memory even
	// when a screen uses feedback surfaces or creates many draw commands.
	for batch := 0; batch < 4 && !r.done; batch++ {
		if r.config.Duration > 0 && r.frames >= int64(r.config.Duration)*int64(r.config.FPS)/int64(time.Second) {
			r.done = true
			break
		}
		target := (r.frames + 1) * int64(r.config.TPS) / int64(r.config.FPS)
		for r.tick < target {
			if err := r.game.Update(); err != nil {
				r.done = true
				if errors.Is(err, ebiten.Termination) {
					r.finished = true
				} else {
					r.err = err
				}
				break
			}
			w, h := r.game.Layout(r.config.Width, r.config.Height)
			if r.surface == nil || r.surface.Bounds().Dx() != w || r.surface.Bounds().Dy() != h {
				if r.surface != nil {
					r.surface.Deallocate()
				}
				r.surface = ebiten.NewImage(w, h)
			}
			r.surface.Clear()
			r.game.Draw(r.surface)
			if chapters, ok := r.game.(interface{ RecordingChapter() string }); ok {
				name := chapters.RecordingChapter()
				if name != "" && (len(r.chapters) == 0 || r.chapters[len(r.chapters)-1].Title != name) {
					r.chapters = append(r.chapters, Chapter{Title: name, Start: float64(r.tick) / float64(r.config.TPS)})
					fmt.Fprintf(os.Stderr, "%.2fs: %s\n", float64(r.tick)/float64(r.config.TPS), name)
				}
			}
			r.tick++
			if err := r.mix.Advance(r.tick, r.config.TPS, r.pcm); err != nil {
				r.err = err
				r.done = true
				break
			}
		}
		if r.done {
			break
		}
		r.canvas.Fill(color.Black)
		scale := math.Min(float64(r.config.Width)/float64(r.surface.Bounds().Dx()), float64(r.config.Height)/float64(r.surface.Bounds().Dy()))
		op := &ebiten.DrawImageOptions{}
		if scale != math.Trunc(scale) {
			op.Filter = ebiten.FilterLinear
		}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate((float64(r.config.Width)-float64(r.surface.Bounds().Dx())*scale)/2, (float64(r.config.Height)-float64(r.surface.Bounds().Dy())*scale)/2)
		r.canvas.DrawImage(r.surface, op)
		r.canvas.ReadPixels(r.pixels)
		if n, err := r.pipe.Write(r.pixels); err != nil {
			r.err = err
			r.done = true
			break
		} else if n != len(r.pixels) {
			r.err = io.ErrShortWrite
			r.done = true
			break
		}
		if r.poster == nil || r.frames == int64(r.config.PosterAt)*int64(r.config.FPS)/int64(time.Second) {
			r.poster = image.NewRGBA(r.canvas.Bounds())
			copy(r.poster.Pix, r.pixels)
		}
		r.frames++
		if time.Since(r.lastProgress) >= 15*time.Second {
			fmt.Fprintf(os.Stderr, "%s: %.1fs rendered\n", r.config.Title, float64(r.frames)/float64(r.config.FPS))
			r.lastProgress = time.Now()
		}
	}
	screen.DrawImage(r.canvas, nil)
}

func writeChapters(name string, chapters []Chapter, duration float64) error {
	var text strings.Builder
	text.WriteString(";FFMETADATA1\n")
	escape := strings.NewReplacer("\\", "\\\\", "=", "\\=", ";", "\\;", "#", "\\#", "\n", " ")
	for i, c := range chapters {
		end := duration
		if i+1 < len(chapters) {
			end = chapters[i+1].Start
		}
		if c.Start >= duration {
			break
		}
		fmt.Fprintf(&text, "[CHAPTER]\nTIMEBASE=1/1000\nSTART=%d\nEND=%d\ntitle=%s\n", int64(c.Start*1000), int64(min(end, duration)*1000), escape.Replace(c.Title))
	}
	return os.WriteFile(name, []byte(text.String()), 0644)
}
