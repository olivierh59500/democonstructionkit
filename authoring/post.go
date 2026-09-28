package authoring

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/effects"
	"github.com/olivierh59500/democonstructionkit/render"
	"github.com/olivierh59500/democonstructionkit/timeline"
)

// withPosts keeps the source's update clock singular while wrapping its image
// in ordered DCK passes. A failed pass closes every resource built so far.
func (b *compiler) withPosts(source kit.Effect, specs []PostEffect) (kit.Effect, error) {
	if len(specs) == 0 {
		return source, nil
	}
	passes := make([]kit.ImagePass, 0, len(specs))
	cleanup := func() {
		for _, pass := range passes {
			if pass.Close != nil {
				_ = pass.Close()
			}
		}
		_ = kit.Close(source)
	}
	for i, spec := range specs {
		pass, err := b.post(spec)
		if err != nil {
			cleanup()
			return nil, fmt.Errorf("pass %d: %w", i, err)
		}
		passes = append(passes, pass)
	}
	pipeline, err := kit.NewPipeline(source, b.canvas.Width, b.canvas.Height, passes...)
	if err != nil {
		cleanup()
		return nil, err
	}
	return pipeline, nil
}

func postOpacity(value *float64) float64 {
	if value == nil {
		return 1
	}
	return *value
}

type postClock struct {
	window timeline.Window
	period float64
}

func (p postClock) At(seconds float64) (local, alpha float64, active bool) {
	if p.period == 0 || seconds < p.window.Start {
		return p.window.At(seconds)
	}
	cycle := math.Mod(seconds-p.window.Start, p.period)
	w := p.window
	w.Start = 0
	return w.At(cycle)
}

func postAlpha(w postClock, base float64, frame kit.Frame) (local, alpha float64) {
	local, envelope, active := w.At(frame.Time)
	if !active {
		return local, 0
	}
	return local, base * envelope
}

func postEnabled(w postClock, base float64) func(kit.Frame) bool {
	return func(frame kit.Frame) bool {
		_, alpha := postAlpha(w, base, frame)
		return alpha > 0
	}
}

func (b *compiler) post(spec PostEffect) (kit.ImagePass, error) {
	w := postClock{window: window(spec.Window), period: spec.Period}
	switch spec.Kind {
	case "crt":
		c := spec.CRT
		blendMode, _ := blend(c.Blend)
		config := effects.CRTOverlayConfig{
			Curvature: float32(c.Curvature), ScanlineFrequency: float32(c.ScanlineFrequency),
			ScanlineAmplitude: float32(c.ScanlineAmplitude), ChromaticShift: float32(c.ChromaticShift),
			Vignette: float32(c.Vignette), NormalizeSource: c.NormalizeSource,
			OutsideTransparent: c.OutsideTransparent,
			SourceOrigin:       image.Pt(c.SourceOrigin.X, c.SourceOrigin.Y), Blend: blendMode,
		}
		pass, err := effects.NewCRTOverlay(config)
		if err != nil {
			return kit.ImagePass{}, err
		}
		opacity := postOpacity(c.Opacity)
		var faded *ebiten.Image
		if opacity < 1 || spec.Window.FadeIn > 0 || spec.Window.FadeOut > 0 {
			faded = render.NewSurface(b.canvas.Width, b.canvas.Height)
		}
		return kit.ImagePass{
			Enabled: postEnabled(w, opacity),
			Apply: func(dst, src *ebiten.Image, frame kit.Frame) {
				_, alpha := postAlpha(w, opacity, frame)
				if alpha >= 1 {
					pass.DrawAt(dst, src, 0, 0)
					return
				}
				if alpha <= 0 {
					dst.DrawImage(src, nil)
					return
				}
				faded.Clear()
				pass.DrawAt(faded, src, 0, 0)
				var options ebiten.DrawImageOptions
				options.Blend = ebiten.BlendCopy
				options.ColorScale.ScaleAlpha(float32(1 - alpha))
				dst.DrawImage(src, &options)
				options = ebiten.DrawImageOptions{Blend: ebiten.BlendLighter}
				options.ColorScale.ScaleAlpha(float32(alpha))
				dst.DrawImage(faded, &options)
			},
			Close: func() error {
				if faded != nil {
					faded.Deallocate()
				}
				return pass.Close()
			},
		}, nil
	case "water_reflection":
		c := spec.WaterReflection
		filterMode, _ := filter(c.Filter)
		blendMode, _ := blend(c.Blend)
		config := composite.DefaultWaterReflectionConfig()
		config.X, config.Horizon, config.Fade = c.X, c.Horizon, c.Fade
		if c.ScaleY != 0 {
			config.ScaleY = c.ScaleY
		}
		if c.Alpha != nil {
			config.Alpha = float32(*c.Alpha)
		}
		config.Wave = composite.WaterWave{Amplitude: c.Wave.Amplitude,
			Wavelength: c.Wave.Wavelength, Speed: c.Wave.Speed, Phase: c.Wave.Phase}
		if config.Wave.Wavelength == 0 {
			config.Wave.Wavelength = 32
		}
		config.RowHeight, config.Filter, config.Blend = c.RowHeight, filterMode, blendMode
		if c.Source != nil {
			config.Source = rectangle(*c.Source)
		}
		if c.Tint != nil {
			config.Tint.Scale(float32(c.Tint[0])/255, float32(c.Tint[1])/255,
				float32(c.Tint[2])/255, float32(c.Tint[3])/255)
		}
		reflection, err := composite.NewWaterReflection(config)
		if err != nil {
			return kit.ImagePass{}, err
		}
		baseAlpha := config.Alpha
		currentAlpha := baseAlpha
		return kit.ImagePass{
			Enabled: postEnabled(w, float64(baseAlpha)),
			Apply: func(dst, src *ebiten.Image, frame kit.Frame) {
				dst.DrawImage(src, nil)
				local, alpha := postAlpha(w, float64(baseAlpha), frame)
				if alpha <= 0 {
					return
				}
				if value := float32(alpha); currentAlpha != value {
					config.Alpha = value
					if err := reflection.SetConfig(config); err != nil {
						return // Validated project data cannot trigger this branch.
					}
					currentAlpha = value
				}
				localFrame := frame
				localFrame.Time = local
				reflection.Draw(dst, src, localFrame)
			},
			Close: reflection.Close,
		}, nil
	case "magnifier":
		c := spec.Magnifier
		magnifier, err := composite.NewMagnifier()
		if err != nil {
			return kit.ImagePass{}, err
		}
		filterMode, _ := filter(c.Filter)
		opacity := postOpacity(c.Opacity)
		options := composite.MagnifierOptions{Radius: c.Radius, Zoom: c.Zoom,
			Falloff: c.Falloff, Feather: c.Feather, Filter: filterMode}
		if c.Crop != nil {
			crop := rectangle(*c.Crop)
			options.Crop = &crop
		}
		return kit.ImagePass{
			Enabled: postEnabled(w, opacity),
			Apply: func(dst, src *ebiten.Image, frame kit.Frame) {
				dst.DrawImage(src, nil)
				local, alpha := postAlpha(w, opacity, frame)
				options.CenterX = c.Center.X + c.Velocity.X*local
				options.CenterY = c.Center.Y + c.Velocity.Y*local
				options.Opacity = alpha
				magnifier.Draw(dst, src, options)
			},
			Close: magnifier.Close,
		}, nil
	}
	return kit.ImagePass{}, fmt.Errorf("unknown pass kind %q", spec.Kind)
}
