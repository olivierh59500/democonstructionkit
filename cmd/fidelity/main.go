// Command fidelity compares demo captures with a chosen Git revision.
package main

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"image"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/olivierh59500/democonstructionkit/fidelity"
)

type probe struct {
	Width, Height    int
	Factory, Imports string
}

var probes = map[string]probe{
	"go-multiscreen":              {800, 600, `g:=&MegaDemoGame{demo1:NewPhenomenaDemo(),demo2:NewTCBDemo(),demo3:NewCocoDemo(),demo4:NewVivaDemo(),cameraState:StateDemo1,needsRedraw:true};for i:=range g.demoCanvases{g.demoCanvases[i]=ebiten.NewImage(demoWidth,demoHeight)};var err error;g.compositeShader,err=ebiten.NewShader([]byte(compositeShaderSource));if err!=nil{return nil,err};g.compositeUniforms=map[string]any{"CameraCenter":g.compositeCenter[:],"CameraZoom":float32(1)};return g,nil`, ""},
	"go-secondreality":            {640, 400, `g:=&dckIndexedFixture{renderer:NewRenderer(),vram:make([]byte,640*400)};g.prepare();return g,nil`, ""},
	"bilizir-demo":                {800, 600, `g:=NewGame();if mode,ok:=any(g).(interface{SetLogoDeformation(bool)});ok{mode.SetLogoDeformation(false)};if err:=g.loadAssets();err!=nil{return nil,err};g.initScrollText();g.initialized=true;return g,nil`, ""},
	"viva_tcb":                    {768, 540, `g:=NewGame();if err:=g.Init();err!=nil{return nil,err};g.audioReady=true;return g,nil`, ""},
	"grodan-kvack-kvack-demo":     {640, 400, `g:=NewGame();g.audioInitialized=true;return g,nil`, ""},
	"dma-3d":                      {640, 480, `g,err:=NewGame();if err!=nil{return nil,err};g.audioInitAttempted=true;return g,nil`, ""},
	"tcb-replicants-demo":         {640, 400, `g:=NewGame();g.rng=rand.New(rand.NewSource(42));if err:=g.Init();err!=nil{return nil,err};g.audioReady=true;return g,nil`, `"math/rand"`},
	"go-dom-intro":                {768, 540, `g:=NewGame();g.audioReady=true;return g,nil`, ""},
	"nonameno-demo":               {640, 480, `g:=NewGame();g.audioReady=true;return g,nil`, ""},
	"3d_doc":                      {768, 540, `g:=NewGame();if err:=g.Init();err!=nil{return nil,err};g.audioReady=true;return g,nil`, ""},
	"teamg1-demo":                 {768, 540, `g:=NewGame();g.audioReady=true;return g,nil`, ""},
	"go-cocoisthebest":            {800, 600, `g:=NewGame();g.audioReady=true;return g,nil`, ""},
	"dma-is-back":                 {768, 540, `g:=NewGame();if mode,ok:=any(g).(interface{SetSmoothTransitions(bool)});ok{mode.SetSmoothTransitions(false)};g.audioReady=true;return g,nil`, ""},
	"go-cuddlymenu":               {768, 536, `g:=NewGame();g.audioReady=true;return g,nil`, ""},
	"go-fr010":                    {640, 480, `g,err:=NewGame();if err!=nil{return nil,err};g.audioReady=true;return g,nil`, ""},
	"go-vectorballs":              {640, 480, `g:=&Game{shapeManager:NewShapeManager(),zoomFactor:.35,fov:1450,centerX:320,centerY:193,position:Vector3{Z:850},transformed:make([]Point3D,0,64),dirty:true};g.loadImages();g.playgroundCanvas=ebiten.NewImage(640,386);g.reflectionSource=g.playgroundCanvas.SubImage(image.Rect(0,288,640,368)).(*ebiten.Image);g.whiteImage=ebiten.NewImage(1,1);g.whiteImage.Fill(color.White);g.initActions();g.currentAction=-1;g.nextAction();return g,nil`, `"image";"image/color"`},
	"megatwist":                   {832, 552, `g:=NewGame();g.audioReady=true;ebiten.SetVsyncEnabled(false);return g,nil`, ""},
	"phenomena-dna-scroll-intro":  {640, 480, `g:=NewGame();if err:=g.Init();err!=nil{return nil,err};g.audioReady=true;return g,nil`, ""},
	"tcb-multi-plane-3d-scroller": {768, 536, `g:=NewGame();g.audioReady=true;return g,nil`, ""},
}

type report struct {
	Demo, Reference, Candidate string
	Scope                      string
	Frames                     []frameResult
}
type frameResult struct {
	Frame int
	fidelity.Result
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	demos := flag.String("demos", "../../demos", "source demo repositories")
	demo := flag.String("demo", "bilizir-demo", "production to compare")
	kitRoot := flag.String("kit", ".", "construction kit checkout")
	out := flag.String("out", "captures/fidelity", "comparison directory")
	referenceOnly := flag.Bool("reference-only", false, "capture the pinned original only")
	reference := flag.String("reference", "", "optional Git revision to compare instead of the pinned original")
	frameList := flag.String("frames", "0,1,60,240,600,1200,2400,4800", "comma-separated capture ticks")
	flag.Parse()
	p, ok := probes[*demo]
	if !ok {
		return fmt.Errorf("no fidelity probe for %s", *demo)
	}
	root, err := filepath.Abs(*kitRoot)
	if err != nil {
		return err
	}
	source, err := filepath.Abs(filepath.Join(*demos, *demo))
	if err != nil {
		return err
	}
	output, err := filepath.Abs(filepath.Join(*out, *demo))
	if err != nil {
		return err
	}
	revision := *reference
	if revision == "" {
		auditData, err := os.ReadFile(filepath.Join(root, "docs/source-audit.json"))
		if err != nil {
			return err
		}
		var audit []struct{ Name, Revision string }
		if err = json.Unmarshal(auditData, &audit); err != nil {
			return err
		}
		for _, a := range audit {
			if a.Name == *demo {
				revision = a.Revision
			}
		}
	}
	if revision == "" {
		return fmt.Errorf("missing original revision")
	}
	revision, err = resolveLocalRevision(source, revision)
	if err != nil {
		return err
	}
	var frames []int
	for _, part := range strings.Split(*frameList, ",") {
		frame, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || frame < 0 || (len(frames) > 0 && frame <= frames[len(frames)-1]) {
			return fmt.Errorf("frames must be nonnegative and strictly increasing")
		}
		frames = append(frames, frame)
	}
	if err = captureRevision(source, revision, root, filepath.Join(output, "reference"), p, frames); err != nil {
		return err
	}
	if *referenceOnly {
		return nil
	}
	headBytes, err := command(source, "git", "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	head := strings.TrimSpace(string(headBytes))
	if err = captureRevision(source, head, root, filepath.Join(output, "candidate"), p, frames); err != nil {
		return err
	}
	r := report{Demo: *demo, Reference: revision, Candidate: head, Scope: "complete production frames; device audio disabled; deterministic clock"}
	if *demo == "bilizir-demo" {
		r.Scope += "; original logo mode (intentional logo deformation disabled)"
	}
	if *demo == "dma-is-back" {
		r.Scope += "; historical cube transitions (intentional continuity fix disabled)"
	}
	if *demo == "go-secondreality" {
		r.Scope = "indexed renderer fixture only; original scene choreography and ST3 synchronization are not exercised"
	}
	failed := false
	for _, frame := range frames {
		name := fmt.Sprintf("%06d.png", frame)
		ref, err := readImage(filepath.Join(output, "reference", name))
		if err != nil {
			return err
		}
		candidate, err := readImage(filepath.Join(output, "candidate", name))
		if err != nil {
			return err
		}
		result, diff, err := fidelity.Compare(ref, candidate)
		if err != nil {
			return err
		}
		r.Frames = append(r.Frames, frameResult{frame, result})
		if result.DifferentPixels > 0 {
			failed = true
			f, err := os.Create(filepath.Join(output, "diff-"+name))
			if err != nil {
				return err
			}
			err = png.Encode(f, diff)
			closeErr := f.Close()
			if err != nil {
				return err
			}
			if closeErr != nil {
				return closeErr
			}
		}
		fmt.Printf("%s frame %d: %d/%d pixels differ; max channel error %d\n", *demo, frame, result.DifferentPixels, result.Pixels, result.MaxChannelError)
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(output, "report.json"), append(data, '\n'), 0644); err != nil {
		return err
	}
	if failed {
		return fmt.Errorf("fidelity mismatch; inspect %s", output)
	}
	return nil
}

// resolveLocalRevision keeps local comparison records usable after an explicit
// history rewrite, without changing the preserved records themselves.
func resolveLocalRevision(source, revision string) (string, error) {
	resolve := func(value string) (string, error) {
		data, err := command(source, "git", "rev-parse", "--verify", value+"^{commit}")
		return strings.TrimSpace(string(data)), err
	}
	if resolved, err := resolve(revision); err == nil {
		return resolved, nil
	}
	path, err := command(source, "git", "rev-parse", "--git-path", "info/native-revision-map.txt")
	if err != nil {
		return "", err
	}
	name := strings.TrimSpace(string(path))
	if !filepath.IsAbs(name) {
		name = filepath.Join(source, name)
	}
	data, err := os.ReadFile(name)
	if os.IsNotExist(err) {
		return resolve(revision)
	}
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == revision {
			return resolve(fields[1])
		}
	}
	return resolve(revision)
}
func readImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}
func command(dir, name string, args ...string) ([]byte, error) {
	c := exec.Command(name, args...)
	c.Dir = dir
	c.Env = append(os.Environ(), "GOWORK=off", "GOPROXY=off", "GOCACHE=/private/tmp/dck-go-build")
	out, err := c.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s %v: %w\n%s", name, args, err, out)
	}
	return out, nil
}
func captureRevision(source, revision, root, output string, p probe, frames []int) error {
	tmp, err := os.MkdirTemp("", "dck-fidelity-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	archive, err := command(source, "git", "archive", "--format=tar", revision)
	if err != nil {
		return err
	}
	reader := tar.NewReader(bytes.NewReader(archive))
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if !filepath.IsLocal(header.Name) {
			return fmt.Errorf("invalid archive path")
		}
		path := filepath.Join(tmp, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(path, 0755)
		case tar.TypeReg:
			if err = os.MkdirAll(filepath.Dir(path), 0755); err == nil {
				var f *os.File
				f, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
				if err == nil {
					_, err = io.Copy(f, reader)
					closeErr := f.Close()
					if err == nil {
						err = closeErr
					}
				}
			}
		}
		if err != nil {
			return err
		}
	}
	packageDir := tmp
	target := "."
	// New revisions keep the DCK implementation alongside the original packages.
	if info, err := os.Stat(filepath.Join(tmp, "dck")); err == nil && info.IsDir() {
		packageDir = filepath.Join(tmp, "dck")
		target = "./dck"
	}
	if filepath.Base(source) == "go-cuddlymenu" {
		packageDir = filepath.Join(packageDir, "menu")
		target += "/menu"
	}
	if filepath.Base(source) == "go-secondreality" {
		packageDir = filepath.Join(packageDir, "internal/graphics")
		target += "/internal/graphics"
	}
	entries, err := os.ReadDir(packageDir)
	if err != nil {
		return err
	}
	pkg := ""
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".go") && !strings.HasSuffix(e.Name(), "_test.go") {
			path := filepath.Join(packageDir, e.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			// Freeze wall-clock animation and random seeding in both snapshots.
			text := strings.ReplaceAll(string(data), "time.Now()", "time.Unix(0, dckFidelityTick*int64(time.Second)/60)")
			text = strings.ReplaceAll(text, "time.Since(", "dckFidelitySince(")
			if err = os.WriteFile(path, []byte(text), 0644); err != nil {
				return err
			}
			f, err := parser.ParseFile(token.NewFileSet(), filepath.Join(packageDir, e.Name()), nil, parser.PackageClauseOnly)
			if err != nil {
				return err
			}
			pkg = f.Name.Name
		}
	}
	if pkg == "" {
		return fmt.Errorf("cannot find root package")
	}
	if filepath.Base(source) == "go-vectorballs" {
		data, err := os.ReadFile(filepath.Join(packageDir, "main.go"))
		if err != nil {
			return err
		}
		if bytes.Contains(data, []byte("func (g *Game) initReflection()")) {
			// New snapshots construct the shared pass; older snapshots retain
			// their original subimage field and must keep the original probe.
			p.Factory = strings.ReplaceAll(p.Factory, "g.reflectionSource=g.playgroundCanvas.SubImage(image.Rect(0,288,640,368)).(*ebiten.Image);", "g.initReflection();")
			p.Imports = `"image/color"`
		}
	}
	frameValues := make([]string, len(frames))
	for i, f := range frames {
		frameValues[i] = fmt.Sprint(f)
	}
	code := fmt.Sprintf(`package %s
import("os";"testing";"fmt";"time";"github.com/hajimehoshi/ebiten/v2";capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten";%s)
var dckFidelityTick int64
func dckFidelitySince(start time.Time)time.Duration{return time.Unix(0,dckFidelityTick*int64(time.Second)/60).Sub(start)}
type dckClockGame struct{ebiten.Game}
func(g dckClockGame)Update()error{dckFidelityTick++;return g.Game.Update()}
func TestMain(m *testing.M){err:=capture.Run(capture.Config{Directory:%q,Width:%d,Height:%d,Frames:[]int{%s}},func()(ebiten.Game,error){makeGame:=func()(ebiten.Game,error){%s};g,err:=makeGame();return dckClockGame{g},err});if err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(1)}}
`, pkg, p.Imports, output, p.Width, p.Height, strings.Join(frameValues, ","), p.Factory)
	if filepath.Base(source) == "go-secondreality" {
		code += `
type dckIndexedFixture struct{renderer *Renderer;vram []byte;palette [256][4]byte;frame int}
func(g *dckIndexedFixture)prepare(){m:=Mode{Width:320,Height:200};if g.frame>=60{m.Height=400};if g.frame>=240{m.Width=640;m.Height=350};if g.frame>=600{m.Height=400};for i:=range g.palette{g.palette[i]=[4]byte{byte(i+g.frame),byte(i^g.frame),byte(255-i),255}};for i:=range g.vram{g.vram[i]=byte(i*7+g.frame*3)};g.renderer.Capture(m,g.vram,&g.palette,0)}
func(g *dckIndexedFixture)Update()error{g.frame++;g.prepare();return nil}
func(g *dckIndexedFixture)Draw(dst *ebiten.Image){g.renderer.Draw(dst)}
func(g *dckIndexedFixture)Layout(int,int)(int,int){return 640,400}
`
	}
	if err = os.WriteFile(filepath.Join(packageDir, "dck_capture_test.go"), []byte(code), 0644); err != nil {
		return err
	}
	if _, err = command(tmp, "go", "mod", "edit", "-go=1.26.0", "-require=github.com/olivierh59500/democonstructionkit@v0.0.0", "-replace=github.com/olivierh59500/democonstructionkit="+root); err != nil {
		return err
	}
	// A production may use the sibling YM checkout for newly supported formats.
	// Rewrite only that existing local replacement inside the temporary archive.
	moduleData, readErr := os.ReadFile(filepath.Join(tmp, "go.mod"))
	if readErr != nil {
		return readErr
	}
	if strings.Contains(string(moduleData), "replace github.com/olivierh59500/ym-player =>") {
		if _, err = command(tmp, "go", "mod", "edit", "-replace=github.com/olivierh59500/ym-player="+filepath.Join(filepath.Dir(root), "ym-player")); err != nil {
			return err
		}
	}
	// Only resolve packages in this capture, preserving the original dependency pins.
	if _, err = command(tmp, "go", "test", "-mod=mod", "-count=1", "-timeout=180s", "-run=^$", target); err != nil {
		return err
	}
	fmt.Printf("captured %s at %s\n", filepath.Base(source), revision[:12])
	return nil
}
