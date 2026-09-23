package scrolling

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// BitmapSlotSample describes a stable recycled slot before its visible pose is
// mapped. Index identifies the slot, not the current character in the message.
type BitmapSlotSample struct {
	Index int
	Rune  rune
	Tick  uint64
	PenX  float64
}
type BitmapSlotPose struct{ X, Y, Angle, ScaleX, ScaleY, Opacity float64 }

// BitmapSlotsConfig combines fixed-step recycled text with arbitrary glyph poses.
// PreviousTangent retains the direction from the previously visited glyph,
// including the previous tick's last glyph. Angles are radians, speeds pixels
// per Update. Recycling happens after the visible pose is recorded.
type BitmapSlotsConfig struct {
	Font                                           BitmapGrid
	Text                                           string
	Count                                          int
	Start, Advance, Speed, RecycleBelow, Period, Y float64
	Waves                                          []RingWave
	PreviousTangent                                bool
	PreviousX, PreviousY                           float64
	Map                                            func(BitmapSlotSample, *BitmapSlotPose) bool
}
type bitmapSlot struct {
	x     float64
	glyph int
}
type bitmapSlotDraw struct {
	pose    BitmapSlotPose
	glyph   int
	visible bool
}
type BitmapSlots struct {
	config               BitmapSlotsConfig
	text                 *BitmapText
	runes                []rune
	slots                []bitmapSlot
	draws                []bitmapSlotDraw
	waves                []RingWave
	next                 int
	tick                 uint64
	previousX, previousY float64
}

func NewBitmapSlots(c BitmapSlotsConfig) (*BitmapSlots, error) {
	if c.Count < 1 || c.Count > 65536 || c.Text == "" {
		return nil, fmt.Errorf("scrolling: invalid recycled bitmap slots")
	}
	if c.Advance == 0 {
		c.Advance = c.Font.Width
	}
	if c.Period == 0 {
		c.Period = float64(c.Count) * c.Advance
	}
	for _, v := range []float64{c.Start, c.Advance, c.Speed, c.RecycleBelow, c.Period, c.Y, c.PreviousX, c.PreviousY} {
		if !finite(v) {
			return nil, fmt.Errorf("scrolling: invalid bitmap slot coordinate")
		}
	}
	if c.Advance <= 0 || c.Period <= 0 || c.Speed < 0 || c.Speed >= c.Period || !finite(c.Start+float64(c.Count-1)*c.Advance) {
		return nil, fmt.Errorf("scrolling: invalid bitmap slot transport")
	}
	for _, w := range c.Waves {
		for _, v := range []float64{w.Phase, w.Amplitude, w.LetterStep, w.TickStep} {
			if !finite(v) {
				return nil, fmt.Errorf("scrolling: invalid bitmap slot wave")
			}
		}
	}
	text, err := NewBitmapText(c.Font, c.Text, c.Advance)
	if err != nil {
		return nil, err
	}
	b := &BitmapSlots{config: c, text: text, runes: []rune(c.Text), slots: make([]bitmapSlot, c.Count), draws: make([]bitmapSlotDraw, c.Count), waves: append([]RingWave(nil), c.Waves...), previousX: c.PreviousX, previousY: c.PreviousY}
	b.config.Text = ""
	b.config.Waves = nil
	b.next = c.Count % text.Len()
	for i := range b.slots {
		b.slots[i] = bitmapSlot{x: c.Start + float64(i)*c.Advance, glyph: i % text.Len()}
		b.draws[i] = bitmapSlotDraw{pose: BitmapSlotPose{X: b.slots[i].x, Y: c.Y, ScaleX: 1, ScaleY: 1, Opacity: 1}, glyph: b.slots[i].glyph, visible: true}
	}
	return b, nil
}
func (b *BitmapSlots) Update(kit.Frame) error { b.Step(); return nil }
func (b *BitmapSlots) Step() {
	for i := range b.slots {
		slot := &b.slots[i]
		slot.x -= b.config.Speed
		y := b.config.Y
		for _, w := range b.waves {
			v := math.Sin(w.Phase+float64(i)*w.LetterStep) * w.Amplitude
			if w.Floor {
				v = math.Floor(v)
			}
			y += v
		}
		draw := &b.draws[i]
		draw.pose = BitmapSlotPose{X: slot.x, Y: y, ScaleX: 1, ScaleY: 1, Opacity: 1}
		pose := &draw.pose
		visible := true
		if b.config.Map != nil {
			visible = b.config.Map(BitmapSlotSample{Index: i, Rune: b.runes[slot.glyph], Tick: b.tick, PenX: slot.x}, pose)
		}
		if b.config.PreviousTangent {
			pose.Angle += math.Atan2(pose.Y-b.previousY, pose.X-b.previousX)
		}
		draw.glyph, draw.visible = slot.glyph, visible
		b.previousX, b.previousY = pose.X, pose.Y
		if slot.x < b.config.RecycleBelow {
			slot.x += b.config.Period
			slot.glyph = b.next
			b.next = (b.next + 1) % b.text.Len()
		}
	}
	for i := range b.waves {
		b.waves[i].Phase += b.waves[i].TickStep
	}
	b.tick++
}
func (b *BitmapSlots) Draw(dst *ebiten.Image) {
	if dst == nil {
		return
	}
	for _, draw := range b.draws {
		p := draw.pose
		if !draw.visible || !b.text.valid[draw.glyph] || !finite(p.X) || !finite(p.Y) || !finite(p.Angle) || !finite(p.ScaleX) || !finite(p.ScaleY) || !finite(p.Opacity) {
			continue
		}
		op := ebiten.DrawImageOptions{Filter: b.config.Font.Filter}
		op.GeoM.Scale(p.ScaleX, p.ScaleY)
		op.GeoM.Rotate(p.Angle)
		op.GeoM.Translate(p.X, p.Y)
		op.ColorScale.ScaleAlpha(float32(p.Opacity))
		composite.DrawRegion(dst, b.config.Font.Image, b.text.regions[draw.glyph], &op)
	}
}

// SetSpeed changes transport without resetting slot positions or wave phases.
func (b *BitmapSlots) SetSpeed(pixelsPerUpdate float64) error {
	if !finite(pixelsPerUpdate) || pixelsPerUpdate < 0 || pixelsPerUpdate >= b.config.Period {
		return fmt.Errorf("scrolling: invalid bitmap slot speed")
	}
	b.config.Speed = pixelsPerUpdate
	return nil
}
