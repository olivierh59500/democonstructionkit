package main

import (
	"fmt"
	"strings"
)

// captureOptions changes only the archived fixture, never the live production.
type captureOptions struct {
	Target     string
	Code       string
	InspectTCB bool
}

func selectProbe(demo, screen, scenario string) (probe, captureOptions, error) {
	if scenario != "idle" && scenario != "controls" {
		return probe{}, captureOptions{}, fmt.Errorf("unknown scenario %q", scenario)
	}
	if screen == "" {
		if scenario != "idle" {
			return probe{}, captureOptions{}, fmt.Errorf("input scenarios require -screen")
		}
		p, ok := probes[demo]
		if !ok {
			return probe{}, captureOptions{}, fmt.Errorf("no fidelity probe for %s; collection captures require -screen", demo)
		}
		return p, captureOptions{}, nil
	}
	if demo != "go-cuddlymenu" && demo != "go-uniondemo" {
		return probe{}, captureOptions{}, fmt.Errorf("-screen is available for Cuddly and Union only")
	}
	if screen == "menu" || strings.HasPrefix(screen, "loader:") {
		if scenario != "idle" {
			return probe{}, captureOptions{}, fmt.Errorf("menu and loader captures support the idle scenario only")
		}
		return presentationProbe(demo, screen)
	}
	if demo == "go-cuddlymenu" {
		return cuddlyProbe(screen, scenario)
	}
	return unionProbe(screen, scenario)
}

func cuddlyProbe(screen, scenario string) (probe, captureOptions, error) {
	dimensions := map[string][2]int{
		"big-sprite": {768, 540}, "colorshock": {768, 540}, "ehh": {768, 540},
		"megascroller": {768, 540}, "spreadpoint": {832, 552}, "digi": {768, 540},
		"led": {768, 540}, "3d-doc": {768, 540}, "fullscreen": {768, 536},
		"starwars": {768, 540}, "knucklebuster": {768, 540}, "dna": {832, 552},
		"megaball": {768, 540}, "intro": {768, 540}, "reset": {768, 540},
	}
	size, ok := dimensions[screen]
	if !ok {
		return probe{}, captureOptions{}, fmt.Errorf("unknown Cuddly screen %q", screen)
	}
	if scenario == "controls" && screen != "megaball" {
		return probe{}, captureOptions{}, fmt.Errorf("the Cuddly controls fixture is defined for megaball only")
	}
	input := ""
	if scenario == "controls" {
		input = `switch dckFidelityTick % 2400 {case 1: in.Up=true; case 601: in.Left=true; case 1201: in.Right=true; case 1801: in.Down=true}`
	}
	code := fmt.Sprintf(`
type dckCollectionGame struct{*Scene}
func(g *dckCollectionGame)Update()error{var in Input;%s;g.Scene.Input(in);return g.Scene.Update()}
`, input)
	return probe{size[0], size[1], fmt.Sprintf(`s,err:=New(%q);if err!=nil{return nil,err};return &dckCollectionGame{s},nil`, screen), ""},
		captureOptions{Target: "dck/screens", Code: code}, nil
}

func unionProbe(screen, scenario string) (probe, captureOptions, error) {
	valid := map[string]bool{"intro": true, "beatdis": true, "delta": true, "tnt3": true,
		"wow": true, "hidden": true, "starballs": true, "replicants": true, "tnt2": true,
		"level16": true, "multiplane": true, "diskcopier": true}
	if !valid[screen] {
		return probe{}, captureOptions{}, fmt.Errorf("unknown Union screen %q", screen)
	}
	imports := `"github.com/olivierh59500/go-uniondemo/assets"`
	input := ""
	if scenario == "controls" {
		switch screen {
		case "tnt3":
			input = `if (dckFidelityTick-1)%600==0{in.Number=1+int((dckFidelityTick-1)/600)%5}`
		case "starballs":
			input = `if (dckFidelityTick-1)%300==0{in.Number=1+int((dckFidelityTick-1)/300)%10}`
		case "tnt2":
			input = `in.Right=dckFidelityTick<1200;in.Left=dckFidelityTick>=1200 && dckFidelityTick<2400;if (dckFidelityTick-1)%600==0{in.Number=1+int((dckFidelityTick-1)/600)%4};in.Action=dckFidelityTick==2401;in.Up=dckFidelityTick>=120 && dckFidelityTick<180;in.Down=dckFidelityTick>=240 && dckFidelityTick<300`
		case "replicants":
			input = `in.Right=dckFidelityTick>=60 && dckFidelityTick<90;in.Left=dckFidelityTick>=120 && dckFidelityTick<150`
		case "diskcopier":
			input = `in.Action=dckFidelityTick==1;if dckFidelityTick==10800{in.Number=10}`
		case "hidden":
			imports += `;"math"`
			input = `t:=float64(dckFidelityTick)/60;in.PointerX+=220*math.Sin(t*2);in.PointerY+=140*math.Cos(t*3);in.PointerDown=true`
		default:
			return probe{}, captureOptions{}, fmt.Errorf("no controls fixture for Union screen %q", screen)
		}
	}
	musicFields, musicFactory, musicStep := "", "", ""
	if screen == "delta" {
		imports += `;"io";"github.com/olivierh59500/democonstructionkit/sound"`
		musicFields = `music *sound.Stream;samples []byte;`
		musicFactory = `data,err:=assets.Files.ReadFile(s.Music);if err!=nil{s.Close();return nil,err};music,err:=sound.Open(s.Music,data,sound.Options{SampleRate:48000,Loop:true,BlockFrames:800});if err!=nil{s.Close();return nil,err};g.music=music;g.samples=make([]byte,800*8);`
		musicStep = `if _,err:=io.ReadFull(g.music,g.samples);err!=nil{return err};if registers,ok:=g.music.YMRegisters();ok{for voice:=range g.scene.VoiceVolumes{g.scene.VoiceVolumes[voice]=registers[8+voice]}};`
	}
	code := fmt.Sprintf(`
type dckCollectionGame struct{scene *Scene;%s}
func(g *dckCollectionGame)Update()error{%sin:=Input{PointerX:384,PointerY:268};%s;return g.scene.Update(in)}
func(g *dckCollectionGame)Draw(dst *ebiten.Image){g.scene.Draw(dst)}
func(g *dckCollectionGame)Layout(int,int)(int,int){return 768,536}
`, musicFields, musicStep, input)
	factory := fmt.Sprintf(`s,err:=New(%q,assets.Files);if err!=nil{return nil,err};g:=&dckCollectionGame{scene:s};%sreturn g,nil`, screen, musicFactory)
	return probe{768, 536, factory, imports}, captureOptions{Target: "internal/screens", Code: code}, nil
}

func presentationProbe(demo, screen string) (probe, captureOptions, error) {
	if screen == "menu" {
		if demo == "go-cuddlymenu" {
			return probes[demo], captureOptions{}, nil
		}
		return probe{768, 536, `s,err:=New(assets.Files);if err!=nil{return nil,err};return &dckCollectionGame{s},nil`, `"github.com/olivierh59500/go-uniondemo/assets"`}, captureOptions{
			Target: "internal/menu",
			Code: `
type dckCollectionGame struct{*Game}
func(g *dckCollectionGame)Update()error{g.Game.Update(Input{});return nil}
func(g *dckCollectionGame)Layout(int,int)(int,int){return 768,536}
`,
		}, nil
	}
	door := strings.TrimPrefix(screen, "loader:")
	if demo == "go-cuddlymenu" {
		valid := map[string]bool{"BIG_SPRITE": true, "COLORSHOCK_II": true, "NO_NAME_1": true,
			"MEGA_SCROLLER": true, "SPREADPOINT": true, "DIGI_DEMO": true, "LED_SCROLLER": true,
			"DOC": true, "FULLSCREEN": true, "STARWARS_DEMO": true, "KNUCKLE_BUSTER": true,
			"DNA_DEMO": true, "NO_NAME_2": true}
		if !valid[door] {
			return probe{}, captureOptions{}, fmt.Errorf("unknown Cuddly loader door %q", door)
		}
		return probe{768, 536, fmt.Sprintf(`return New(%q)`, door), ""}, captureOptions{Target: "dck/loader"}, nil
	}
	if _, _, err := unionProbe(door, "idle"); err != nil || door == "intro" {
		return probe{}, captureOptions{}, fmt.Errorf("unknown Union loader door %q", door)
	}
	return probe{768, 536, fmt.Sprintf(`s,err:=New(%q,assets.Files,60);if err!=nil{return nil,err};return &dckCollectionGame{s},nil`, door), `"github.com/olivierh59500/go-uniondemo/assets"`}, captureOptions{
		Target: "internal/loader",
		Code: `
type dckCollectionGame struct{*Screen}
func(g *dckCollectionGame)Update()error{g.Screen.Update();return nil}
func(g *dckCollectionGame)Layout(int,int)(int,int){return 768,536}
`,
	}, nil
}
