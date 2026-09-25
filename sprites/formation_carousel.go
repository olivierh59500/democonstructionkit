package sprites

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// FormationCarouselConfig cycles compiled sprite trajectories with an entry
// and exit slide. Radii are fractions of half the destination size. Formulas
// own the position logic, while Atlas and ImageOrder select borrowed artwork.
type FormationCarouselConfig struct {
	Atlas                   *Atlas
	Count                   int
	ImageOrder              []int
	Modes                   []motion.FormulaFormationConfig
	Hold, Slide, SlideIndex float64
	RadiusX, RadiusY        float64
	AnchorX, AnchorY        float64
	PixelSnap               bool
	Filter                  ebiten.Filter
	Blend                   ebiten.Blend
}

// FormationCarousel draws an arbitrary image formation from absolute time.
// It retains no GPU surface or mutable clock, so repeated Draw calls at the
// same time reproduce the same poses without allocating per sprite.
type FormationCarousel struct {
	config FormationCarouselConfig
	modes  []*motion.FormulaFormation
}

func NewFormationCarousel(config FormationCarouselConfig) (*FormationCarousel, error) {
	if config.Atlas == nil || config.Atlas.Image == nil || config.Count < 1 || config.Count > 1_000_000 || len(config.Modes) == 0 || len(config.Modes) > 64 ||
		len(config.ImageOrder) != 0 && len(config.ImageOrder) != config.Count ||
		config.Hold <= 0 || config.Slide <= 0 || config.SlideIndex < 0 {
		return nil, fmt.Errorf("sprites: invalid formation carousel dimensions")
	}
	for _, value := range [...]float64{config.Hold, config.Slide, config.SlideIndex, config.RadiusX, config.RadiusY, config.AnchorX, config.AnchorY} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("sprites: nonfinite formation carousel setting")
		}
	}
	cycle := config.Hold + config.Slide*2
	if math.IsInf(cycle, 0) || math.IsInf(cycle*float64(len(config.Modes)), 0) {
		return nil, fmt.Errorf("sprites: formation carousel duration overflows")
	}
	for _, index := range config.ImageOrder {
		if index < 0 {
			return nil, fmt.Errorf("sprites: negative image order index")
		}
	}
	carousel := &FormationCarousel{config: config, modes: make([]*motion.FormulaFormation, len(config.Modes))}
	for i, recipe := range config.Modes {
		mode, err := motion.NewFormulaFormation(recipe)
		if err != nil {
			return nil, fmt.Errorf("sprites: formation mode %d: %w", i, err)
		}
		carousel.modes[i] = mode
	}
	carousel.config.ImageOrder = append([]int(nil), config.ImageOrder...)
	carousel.config.Modes = nil
	return carousel, nil
}

type carouselPoseContext struct {
	time, radiusX, radiusY, centerX, centerY, slideX float64
	mode                                             int
}

func carouselInterpolate(time, duration, min, max float64) float64 {
	return ((math.Mod(time, duration) * (max - min)) / duration) + min
}

func (carousel *FormationCarousel) context(time, width, height float64) carouselPoseContext {
	config := carousel.config
	cycle := config.Hold + config.Slide*2
	maximum := float64(len(carousel.modes)) * cycle
	if time >= maximum {
		time = math.Mod(time, maximum)
	}
	time += config.Slide
	timerCycle := math.Floor(carouselInterpolate(time, cycle, 0, cycle))
	slideX := 0.0
	if timerCycle <= config.Slide-config.SlideIndex {
		slideX = -width * carouselInterpolate(time, config.Slide, 0, 1)
	} else if timerCycle <= config.Slide+config.Slide-config.SlideIndex {
		slideX = width * carouselInterpolate(time, config.Slide, 1, 0)
	}
	mode := int(math.Floor((time-config.Slide)/cycle)) % len(carousel.modes)
	if mode < 0 {
		mode += len(carousel.modes)
	}
	centerX, centerY := width*.5, height*.5
	return carouselPoseContext{
		time: time, mode: mode, centerX: centerX, centerY: centerY,
		radiusX: centerX * config.RadiusX, radiusY: centerY * config.RadiusY,
		slideX: slideX,
	}
}

func (carousel *FormationCarousel) pose(context carouselPoseContext, index int) (motion.Point, int) {
	p := carousel.modes[context.mode].At(context.time, index, context.radiusX, context.radiusY, carousel.config.Count)
	p.X += context.centerX + context.slideX
	p.Y += context.centerY
	if carousel.config.PixelSnap {
		p.X, p.Y = math.Floor(p.X), math.Floor(p.Y)
	}
	p.X -= carousel.config.AnchorX
	p.Y -= carousel.config.AnchorY
	frame := index
	if len(carousel.config.ImageOrder) > 0 {
		frame = carousel.config.ImageOrder[index]
	}
	return p, frame
}

// PoseAt exposes one prepared top-left position for another sprite material.
func (carousel *FormationCarousel) PoseAt(time float64, index, width, height int) (motion.Point, int) {
	if index < 0 || index >= carousel.config.Count || width < 1 || height < 1 || math.IsNaN(time) || math.IsInf(time, 0) {
		return motion.Point{}, 0
	}
	return carousel.pose(carousel.context(time, float64(width), float64(height)), index)
}

func (carousel *FormationCarousel) Draw(dst *ebiten.Image, time float64) {
	if carousel == nil || dst == nil || math.IsNaN(time) || math.IsInf(time, 0) {
		return
	}
	context := carousel.context(time, float64(dst.Bounds().Dx()), float64(dst.Bounds().Dy()))
	for index := 0; index < carousel.config.Count; index++ {
		point, frame := carousel.pose(context, index)
		var options ebiten.DrawImageOptions
		options.Filter, options.Blend = carousel.config.Filter, carousel.config.Blend
		options.GeoM.Translate(point.X, point.Y)
		dst.DrawImage(carousel.config.Atlas.Tile(frame), &options)
	}
}
