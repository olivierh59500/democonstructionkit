// Command wordobjects composes three native-precision mesh instances.
package main

import (
	"flag"
	"image"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/effects"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/motion"
)

const width, height = 720, 400

type game struct {
	meshes  [3]*effects.WordMesh
	palette *ebiten.Image
	tick    int
}

func newGame() (_ *game, err error) {
	g := &game{}
	defer func() {
		if err != nil {
			g.Close()
		}
	}()
	sines := make([]int16, 2048)
	for i := range sines {
		sines[i] = int16(math.Round(32767 * math.Sin(float64(i)*2*math.Pi/2048)))
	}
	matrix, err := motion.NewWordEulerMatrix(motion.WordEulerMatrixConfig{Sines: sines, Period: 4096, Quantum: 2, Quarter: 1024, ProductShift: 6, ThirdAxisShift: 7, MiddleAssociation: motion.WordMiddleXFirst})
	if err != nil {
		return nil, err
	}
	pixels := image.NewNRGBA(image.Rect(0, 0, 3, height))
	for y := 0; y < height; y++ {
		for x := 0; x < 3; x++ {
			v := byte(70 + y*180/height)
			paint := color.NRGBA{A: 255}
			switch x {
			case 0:
				paint.R, paint.G = v, v/3
			case 1:
				paint.G, paint.B = v, v/2
			case 2:
				paint.R, paint.B = v/2, v
			}
			pixels.SetNRGBA(x, y, paint)
		}
	}
	g.palette = ebiten.NewImageFromImage(pixels)
	cube := effects.WordMeshModel{Points: []motion.WrappedPoint{{X: -80, Y: -80, Z: -80}, {X: 80, Y: -80, Z: -80}, {X: 80, Y: 80, Z: -80}, {X: -80, Y: 80, Z: -80}, {X: -80, Y: -80, Z: 80}, {X: 80, Y: -80, Z: 80}, {X: 80, Y: 80, Z: 80}, {X: -80, Y: 80, Z: 80}}}
	for i, face := range effects.CubeFaces() {
		cube.Faces = append(cube.Faces, effects.WordFace{Contours: [][]int{face}, Material: i % 3})
	}
	for i := range g.meshes {
		index := i
		g.meshes[i], err = effects.NewWordMesh(effects.WordMeshConfig{Models: []effects.WordMeshModel{cube}, Matrix: matrix, Projection: geometry.WordProjectionConfig{DepthShift: 9, DepthBias: 420, Center: [2]int16{int16(120 + i*240), 220}}, Textures: []*ebiten.Image{g.palette}, CullBackFaces: true, BatchTriangles: 64,
			Animate: func(f kit.Frame) effects.WordMeshPose {
				return effects.WordMeshPose{Angles: [3]int16{int16(f.Tick * uint64(5+index*2)), int16(f.Tick * uint64(7+index)), int16(f.Tick * uint64(3+index))}, Depth: int16(90 * math.Sin(f.Time*.9+float64(index))), Visible: true}
			},
			UV: func(s effects.WordMaterialSample) geometry.Vec2 {
				return geometry.Vec2{X: float64(s.Material) + .5, Y: s.Screen.Y}
			},
		})
		if err != nil {
			return nil, err
		}
	}
	return g, nil
}
func (g *game) Update() error {
	g.tick++
	f := kit.Frame{Tick: uint64(g.tick), Time: float64(g.tick) / 50, Delta: 1.0 / 50}
	for _, m := range g.meshes {
		if m.Err() != nil {
			return m.Err()
		}
		if err := m.Update(f); err != nil {
			return err
		}
	}
	return nil
}
func (g *game) Draw(dst *ebiten.Image) {
	dst.Fill(color.NRGBA{R: 6, G: 12, B: 24, A: 255})
	for _, m := range g.meshes {
		m.Draw(dst)
	}
	ebitenutil.DebugPrintAt(dst, "SHARED WORD MESH - THREE INDEPENDENT CLOCKS / DEPTHS", 16, 18)
	ebitenutil.DebugPrintAt(dst, "COPIED GEOMETRY + CACHED PROJECTION / CULLING + BORROWED ROW PALETTE", 16, 376)
}
func (*game) Layout(int, int) (int, int) { return width, height }
func (g *game) Close() {
	for _, m := range g.meshes {
		if m != nil {
			m.Close()
		}
	}
	if g.palette != nil {
		g.palette.Deallocate()
	}
}
func main() {
	directory := flag.String("capture", "", "write native PNG frames")
	first := flag.Int("frame", 200, "first update tick")
	count := flag.Int("frames", 1, "consecutive frames (1..1500)")
	flag.Parse()
	if *directory != "" {
		if *first < 0 || *count < 1 || *count > 1500 || *first > int(^uint(0)>>1)-*count {
			log.Fatal("invalid capture range")
		}
		frames := make([]int, *count)
		for i := range frames {
			frames[i] = *first + i
		}
		var g *game
		err := capture.Run(capture.Config{Directory: *directory, Frames: frames, Width: width, Height: height}, func() (ebiten.Game, error) { var err error; g, err = newGame(); return g, err })
		if g != nil {
			g.Close()
		}
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	g, err := newGame()
	if err != nil {
		log.Fatal(err)
	}
	defer g.Close()
	ebiten.SetTPS(50)
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle("DCK Word Objects")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
