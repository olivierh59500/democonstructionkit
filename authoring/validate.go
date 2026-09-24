package authoring

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"reflect"

	"github.com/olivierh59500/democonstructionkit/modulation"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/democonstructionkit/scrolltext"
	"github.com/olivierh59500/democonstructionkit/timeline"
)

func decodeBytes(data []byte) (*Project, error) {
	d := json.NewDecoder(bytes.NewReader(data))
	if err := uniqueValue(d); err != nil {
		return nil, fmt.Errorf("authoring: %w", err)
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, fmt.Errorf("authoring: trailing JSON value")
	}
	if err := canonicalFields(data, reflect.TypeOf(Project{})); err != nil {
		return nil, fmt.Errorf("authoring: %w", err)
	}
	d = json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	var p Project
	if err := d.Decode(&p); err != nil {
		return nil, fmt.Errorf("authoring: %w", err)
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return &p, nil
}

// encoding/json intentionally accepts case-insensitive struct field aliases.
// Saved projects instead require the canonical schema names, while dictionary
// keys such as asset IDs remain case-sensitive and are never normalized.
func canonicalFields(raw json.RawMessage, t reflect.Type) error {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		switch t.Kind() {
		case reflect.Pointer, reflect.Map, reflect.Slice:
			return nil
		}
		return fmt.Errorf("null is not a valid %s", t)
	}
	if t.Kind() == reflect.Pointer {
		return canonicalFields(raw, t.Elem())
	}
	switch t.Kind() {
	case reflect.Struct:
		var object map[string]json.RawMessage
		if err := json.Unmarshal(raw, &object); err != nil {
			return err
		}
		fields := map[string]reflect.Type{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			name := f.Tag.Get("json")
			for j, ch := range name {
				if ch == ',' {
					name = name[:j]
					break
				}
			}
			if name == "-" {
				continue
			}
			if name == "" {
				name = f.Name
			}
			fields[name] = f.Type
		}
		for name, value := range object {
			field, ok := fields[name]
			if !ok {
				return fmt.Errorf("unknown or noncanonical field %q in %s", name, t)
			}
			if err := canonicalFields(value, field); err != nil {
				return fmt.Errorf("field %q: %w", name, err)
			}
		}
	case reflect.Map:
		var values map[string]json.RawMessage
		if err := json.Unmarshal(raw, &values); err != nil {
			return err
		}
		for name, value := range values {
			if err := canonicalFields(value, t.Elem()); err != nil {
				return fmt.Errorf("entry %q: %w", name, err)
			}
		}
	case reflect.Slice, reflect.Array:
		var values []json.RawMessage
		if err := json.Unmarshal(raw, &values); err != nil {
			return err
		}
		if t.Kind() == reflect.Array && len(values) != t.Len() {
			return fmt.Errorf("expected %d array entries", t.Len())
		}
		for i, value := range values {
			if err := canonicalFields(value, t.Elem()); err != nil {
				return fmt.Errorf("index %d: %w", i, err)
			}
		}
	}
	return nil
}

func uniqueValue(d *json.Decoder) error {
	t, err := d.Token()
	if err != nil {
		return err
	}
	delim, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		keys := map[string]bool{}
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return err
			}
			s, ok := key.(string)
			if !ok {
				return fmt.Errorf("invalid object key")
			}
			if keys[s] {
				return fmt.Errorf("duplicate field %q", s)
			}
			keys[s] = true
			if err = uniqueValue(d); err != nil {
				return err
			}
		}
	case '[':
		for d.More() {
			if err := uniqueValue(d); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected delimiter %q", delim)
	}
	_, err = d.Token()
	return err
}

func (p Project) Validate() error {
	if p.Version != Version {
		return fmt.Errorf("authoring: unsupported version %d", p.Version)
	}
	if p.Units != DefaultUnits() {
		return fmt.Errorf("authoring: version 1 requires pixels, seconds and radians")
	}
	if p.Canvas.Width < 1 || p.Canvas.Height < 1 || p.Canvas.Width > 8192 || p.Canvas.Height > 8192 || p.Canvas.TPS < 1 || p.Canvas.TPS > 240 {
		return fmt.Errorf("authoring: invalid canvas dimensions or tick rate")
	}
	if _, err := json.Marshal(p); err != nil {
		return fmt.Errorf("authoring: nonserializable project: %w", err)
	}
	if p.BPM < 0 || p.BPM > 1000 {
		return fmt.Errorf("authoring: invalid tempo")
	}
	if len(p.Layers) < 1 || len(p.Layers) > 256 {
		return fmt.Errorf("authoring: expected 1..256 ordered layers")
	}
	for id, kind := range p.Assets {
		if id == "" || (kind != "image" && kind != "font") {
			return fmt.Errorf("authoring: invalid asset %q kind %q", id, kind)
		}
	}
	for id, spec := range p.Signals {
		if id == "" {
			return fmt.Errorf("authoring: empty signal ID")
		}
		if spec.TimeBase == modulation.Beats && p.BPM <= 0 {
			return fmt.Errorf("authoring: signal %q needs a positive BPM", id)
		}
		if _, err := modulation.New(spec); err != nil {
			return fmt.Errorf("authoring: signal %q: %w", id, err)
		}
	}
	ids := map[string]bool{}
	for _, l := range p.Layers {
		if l.ID == "" || ids[l.ID] {
			return fmt.Errorf("authoring: missing or duplicate layer ID %q", l.ID)
		}
		ids[l.ID] = true
		if err := p.validateLayer(l); err != nil {
			return fmt.Errorf("authoring: layer %q: %w", l.ID, err)
		}
	}
	return nil
}

func (p Project) asset(id, kind string) error {
	if id == "" || p.Assets[id] != kind {
		return fmt.Errorf("unknown %s asset %q", kind, id)
	}
	return nil
}
func validRect(r *Rect) error {
	if r != nil && (r.Width <= 0 || r.Height <= 0 || r.Width > 8192 || r.Height > 8192) {
		return fmt.Errorf("invalid image rectangle")
	}
	if r != nil {
		limit := int(^uint(0) >> 1)
		if r.X > limit-r.Width || r.Y > limit-r.Height {
			return fmt.Errorf("image rectangle endpoint overflows")
		}
	}
	return nil
}
func (p Project) validateLayer(l Layer) error {
	if l.Window.Start < 0 {
		return fmt.Errorf("negative start time")
	}
	if err := window(l.Window).Validate(); err != nil {
		return err
	}
	if _, err := blend(l.Blend); err != nil {
		return err
	}
	n := 0
	if l.Scroll != nil {
		n++
	}
	if l.Sprites != nil {
		n++
	}
	if l.Background != nil {
		n++
	}
	if l.Rotozoom != nil {
		n++
	}
	if l.JellyCube != nil {
		n++
	}
	if n != 1 {
		return fmt.Errorf("choose exactly one effect configuration")
	}
	switch l.Kind {
	case "jelly_cube":
		if l.JellyCube == nil {
			return fmt.Errorf("jellyCube config required")
		}
		_, err := jellyConfig(*l.JellyCube)
		return err
	case "scroll":
		if l.Scroll == nil {
			return fmt.Errorf("scroll config required")
		}
		return p.validateScroll(*l.Scroll)
	case "sprites":
		if l.Sprites == nil {
			return fmt.Errorf("sprites config required")
		}
		return p.validateSprites(*l.Sprites)
	case "background":
		if l.Background == nil {
			return fmt.Errorf("background config required")
		}
		c := l.Background
		if err := p.asset(c.Image, "image"); err != nil {
			return err
		}
		if err := validRect(c.Source); err != nil {
			return err
		}
		if c.Period.X < 0 || c.Period.Y < 0 || c.Scale.X < 0 || c.Scale.Y < 0 {
			return fmt.Errorf("negative repeat period or scale")
		}
		if c.CopiesX < 0 || c.CopiesY < 0 || c.CopiesX > 1<<20 || c.CopiesY > 1<<20 {
			return fmt.Errorf("invalid background copy count")
		}
		if _, err := filter(c.Filter); err != nil {
			return err
		}
		_, err := blend(c.Blend)
		return err
	case "rotozoom":
		if l.Rotozoom == nil {
			return fmt.Errorf("rotozoom config required")
		}
		c := l.Rotozoom
		if err := p.asset(c.Image, "image"); err != nil {
			return err
		}
		if c.Zoom <= 0 {
			return fmt.Errorf("rotozoom zoom must be positive")
		}
		_, err := filter(c.Filter)
		return err
	default:
		return fmt.Errorf("unknown effect kind %q", l.Kind)
	}
}

func (p Project) validateScroll(c Scroll) error {
	if c.Speed < 0 || c.Gap < 0 || c.Advance < 0 || len(c.Fonts) == 0 {
		return fmt.Errorf("invalid text speed, gap or font bank")
	}
	if _, ok := c.Fonts[c.Font]; !ok {
		return fmt.Errorf("unknown initial font %q", c.Font)
	}
	for name, id := range c.Fonts {
		if name == "" {
			return fmt.Errorf("empty font alias")
		}
		if err := p.asset(id, "font"); err != nil {
			return err
		}
	}
	if err := validRect(c.RepeatBounds); err != nil {
		return err
	}
	if c.Controls != "" && c.Controls != "braces" {
		return fmt.Errorf("unknown text control syntax %q", c.Controls)
	}
	for name, m := range c.Modes {
		if name == "" || name == "none" {
			return fmt.Errorf("reserved mode name %q", name)
		}
		if _, err := compileMode(m, c.Vertical || c.Page != nil); err != nil {
			return fmt.Errorf("mode %q: %w", name, err)
		}
	}
	knownMode := func(name string) bool {
		if name == "" || name == "none" {
			return true
		}
		_, ok := c.Modes[name]
		return ok
	}
	if !knownMode(c.Mode) {
		return fmt.Errorf("unknown initial mode %q", c.Mode)
	}
	if c.Page != nil {
		if c.Page.Width < 0 || c.Page.LineHeight < 0 {
			return fmt.Errorf("invalid page dimensions")
		}
		if _, err := alignment(c.Page.Align); err != nil {
			return err
		}
	}
	if len(c.Sequence) > 0 {
		cues := make([]scrolling.Cue, len(c.Sequence))
		for i, q := range c.Sequence {
			if !knownMode(q.Mode) {
				return fmt.Errorf("unknown cue mode %q", q.Mode)
			}
			cues[i] = scrolling.Cue{At: q.At, Mode: q.Mode}
		}
		if _, err := scrolling.NewModeSequence(cues, c.SequencePeriod); err != nil {
			return err
		}
	} else if c.SequencePeriod != 0 {
		return fmt.Errorf("sequence period without cues")
	}
	var decoder scrolltext.Decoder
	if c.Controls == "braces" {
		decoder = scrolltext.Braces
	}
	tokens, err := scrolltext.Parse(c.Text, decoder)
	if err != nil {
		return err
	}
	for _, t := range tokens {
		if math.IsNaN(t.Value) || math.IsInf(t.Value, 0) || math.IsNaN(t.Second) || math.IsInf(t.Second, 0) {
			return fmt.Errorf("nonfinite text control value")
		}
		switch t.Kind {
		case scrolltext.Font:
			if _, ok := c.Fonts[t.Text]; !ok {
				return fmt.Errorf("unknown text font %q", t.Text)
			}
		case scrolltext.Shape:
			if !knownMode(t.Text) {
				return fmt.Errorf("unknown text mode %q", t.Text)
			}
		case scrolltext.Effect:
			if t.Text != "none" {
				return fmt.Errorf("glyph effect controls are outside schema version 1")
			}
		case scrolltext.Speed, scrolltext.Pause:
			if t.Value < 0 {
				return fmt.Errorf("negative text speed or pause")
			}
		case scrolltext.Scale:
			if t.Value <= 0 || t.Second <= 0 {
				return fmt.Errorf("nonpositive text scale")
			}
		}
	}
	return nil
}

func (p Project) validateSprites(c SpriteGroup) error {
	if c.Count < 1 || c.Count > 10000 || len(c.Images) == 0 || c.FPS < 0 || c.Delay < 0 {
		return fmt.Errorf("invalid sprite count, frames, FPS or delay")
	}
	for _, id := range c.Images {
		if err := p.asset(id, "image"); err != nil {
			return err
		}
	}
	kinds := 0
	if len(c.Points) > 0 {
		kinds++
		if _, err := compilePath(Path{Points: c.Points, Closed: c.Closed, SplineSamples: c.SplineSamples}); err != nil {
			return err
		}
	}
	if c.Orbit != nil {
		kinds++
	}
	if c.Weave != nil {
		kinds++
	}
	if kinds > 1 {
		return fmt.Errorf("choose one sprite trajectory")
	}
	if len(c.Points) == 0 && (c.SplineSamples != 0 || c.Closed) {
		return fmt.Errorf("path options without points")
	}
	if len(c.PerInstance) > c.Count {
		return fmt.Errorf("too many per-instance bindings")
	}
	if _, err := filter(c.Filter); err != nil {
		return err
	}
	if _, err := blend(c.Blend); err != nil {
		return err
	}
	all := append([]Bindings{c.Signals}, c.PerInstance...)
	for _, bindings := range all {
		for property, id := range bindings {
			switch property {
			case "x", "y", "scaleX", "scaleY", "angle", "opacity":
			default:
				return fmt.Errorf("unknown sprite property %q", property)
			}
			if _, ok := p.Signals[id]; !ok {
				return fmt.Errorf("unknown signal %q", id)
			}
		}
	}
	return nil
}

func window(w Window) timeline.Window {
	return timeline.Window{Start: w.Start, Duration: w.Duration, FadeIn: w.FadeIn, FadeOut: w.FadeOut}
}
