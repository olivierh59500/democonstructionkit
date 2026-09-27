package authoring

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

func formationProject() Project {
	p := testProject()
	p.Layers = []Layer{{ID: "letters", Kind: "sprites", Sprites: &SpriteGroup{
		Images: []string{"tile"}, Count: 3, FrameStride: 1, Speed: 1,
		Formation: &CuedFormation{
			Loop: 4,
			PoseKeys: []FormationPoseKey{
				{Time: 0, Origin: Point{X: 10, Y: 20}, Spacing: Point{X: 30}},
				{Time: 1, Origin: Point{X: 50, Y: 30}, Spacing: Point{X: -10}, Arc: Point{Y: -10}},
				{Time: 4, Origin: Point{X: 10, Y: 20}, Spacing: Point{X: 30}},
			},
			Cues: []FormationCue{{Start: .25, Duration: .5, LeadIndex: 0,
				X: []FormationHarmonic{{FirstAmplitude: 20, LastAmplitude: 20, Cycles: .5}}}},
		},
	}}}
	return p
}

func TestSerializableFormationCompilesToSpritePoses(t *testing.T) {
	p := formationProject()
	var encoded bytes.Buffer
	if err := Encode(&encoded, p); err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(&encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*decoded, p) {
		t.Fatalf("formation round trip changed project: %+v", decoded.Layers[0].Sprites.Formation)
	}
	assets := testAssets(t)
	compiled, err := Compile(*decoded, assets, Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = compiled.Close() })
	b := compiler{resolver: assets, images: map[string]*ebiten.Image{}}
	effect, err := b.sprites(*decoded.Layers[0].Sprites)
	if err != nil {
		t.Fatal(err)
	}
	group := effect.(*sprites.Group)
	for _, sample := range []struct {
		seconds float64
		want    [3][2]float64
	}{
		{0, [3][2]float64{{10, 20}, {40, 20}, {70, 20}}},
		{.5, [3][2]float64{{50, 25}, {60, 20}, {70, 25}}},
		{1, [3][2]float64{{50, 30}, {40, 20}, {30, 30}}},
		{4, [3][2]float64{{10, 20}, {40, 20}, {70, 20}}},
	} {
		if err := group.Update(kit.Frame{Time: sample.seconds}); err != nil {
			t.Fatal(err)
		}
		for i, pose := range group.Poses() {
			if pose.X != sample.want[i][0] || pose.Y != sample.want[i][1] {
				t.Fatalf("sprite %d at %.1fs = (%.1f, %.1f), want %v", i, sample.seconds, pose.X, pose.Y, sample.want[i])
			}
		}
	}
}

func TestSerializableFormationRejectsBrokenTrajectories(t *testing.T) {
	for name, mutate := range map[string]func(*SpriteGroup){
		"unknown easing": func(g *SpriteGroup) { g.Formation.PoseKeys[0].Ease = "bounce" },
		"duplicate key":  func(g *SpriteGroup) { g.Formation.PoseKeys[1].Time = 0 },
		"loop jump":      func(g *SpriteGroup) { g.Formation.PoseKeys[2].Spacing.X = 31 },
		"cue overflow":   func(g *SpriteGroup) { g.Formation.Cues[0].Start = 3.75 },
		"mixed keys":     func(g *SpriteGroup) { g.Formation.ArcKeys = []FormationPointKey{{Time: 0}} },
		"competing path": func(g *SpriteGroup) { g.Points = []Point{{}, {X: 10}} },
	} {
		t.Run(name, func(t *testing.T) {
			p := formationProject()
			mutate(p.Layers[0].Sprites)
			if err := p.Validate(); err == nil {
				t.Fatal("accepted invalid sprite formation")
			}
		})
	}
}

func TestSerializableFormationIndependentPropertyClocks(t *testing.T) {
	formation, err := compileFormation(CuedFormation{
		Spacing: Point{X: 20}, Loop: 2,
		OriginKeys: []FormationPointKey{
			{Time: 0, Value: Point{}, Ease: "smooth"},
			{Time: 1, Value: Point{X: 40}},
			{Time: 2, Value: Point{}},
		},
		ArcKeys: []FormationPointKey{
			{Time: 0, Value: Point{}},
			{Time: 1, Value: Point{Y: -20}},
			{Time: 2, Value: Point{}},
		},
	}, 3)
	if err != nil {
		t.Fatal(err)
	}
	if got := formation.At(.25, 1); got.X != 26.25 || got.Y != -5 {
		t.Fatalf("independent ease and arc at .25s = %+v", got)
	}
	if got := formation.At(2, 1); got.X != 20 || got.Y != 0 {
		t.Fatalf("independent property loop jumped: %+v", got)
	}
}
