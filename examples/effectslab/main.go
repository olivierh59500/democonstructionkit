// Command effectslab demonstrates reusable live layers without external assets.
package main

import (
	"flag"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/examples/effectslab/scene"
	"log"
)

func main() {
	c := scene.Config{StopWhenDone: true}
	flag.BoolVar(&c.Eco, "eco", false, "render at 320x180 with four-pixel reflection strips")
	flag.IntVar(&c.Frames, "frames", 0, "stop after this many 60 Hz update ticks; zero runs interactively")
	flag.StringVar(&c.Profile, "profile", "", "write bounded CPU and memory measurements as JSON; defaults to 600 frames")
	flag.StringVar(&c.Capture, "capture", "", "write one logical-frame PNG snapshot")
	flag.IntVar(&c.CaptureFrame, "capture-frame", 300, "update tick at which to capture the PNG")
	flag.Parse()
	if c.Capture == "" {
		c.CaptureFrame = 0
	}
	g, err := scene.NewGame(c)
	if err != nil {
		log.Fatal(err)
	}
	defer g.Close()
	ebiten.SetTPS(60)
	ebiten.SetWindowSize(960, 540)
	ebiten.SetWindowTitle("DCK Effects Lab")
	if err = ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
	if r := g.Report(); r != nil {
		fmt.Printf("%dx%d: %.1f FPS / %.1f TPS, Update %.1f us, Draw submission %.1f us\n", r.Resolution[0], r.Resolution[1], r.ActualFPS, r.ActualTPS, r.Update.MeanUS, r.Draw.MeanUS)
	}
}
