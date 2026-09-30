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
	faceColors := [...]color.NRGBA{
		{R: 255, G: 87, B: 66, A: 255},
		{R: 64, G: 210, B: 255, A: 255},
		{R: 255, G: 195, B: 56, A: 255},
		{R: 105, G: 225, B: 126, A: 255},
		{R: 188, G: 112, B: 255, A: 255},
		{R: 255, G: 115, B: 197, A: 255},
	}
	pixels := image.NewNRGBA(image.Rect(0, 0, len(faceColors), height))
	for y := 0; y < height; y++ {
		for x, base := range faceColors {
			// A subtle row gradient retains a distinct color on every face.
			shade := 180 + y*75/height
			paint := color.NRGBA{R: byte(int(base.R) * shade / 255), G: byte(int(base.G) * shade / 255), B: byte(int(base.B) * shade / 255), A: 255}
			pixels.SetNRGBA(x, y, paint)
		}
	}
	g.palette = ebiten.NewImageFromImage(pixels)
	cube := effects.WordMeshModel{Points: []motion.WrappedPoint{{X: -80, Y: -80, Z: -80}, {X: 80, Y: -80, Z: -80}, {X: 80, Y: 80, Z: -80}, {X: -80, Y: 80, Z: -80}, {X: -80, Y: -80, Z: 80}, {X: 80, Y: -80, Z: 80}, {X: 80, Y: 80, Z: 80}, {X: -80, Y: 80, Z: 80}}}
	for i, face := range effects.CubeFaces() {
		// CubeFaces uses outward world normals; WordMesh's screen convention
		// accepts positive winding with Y down. Reverse this recipe's faces so
		// the near surfaces hide the far surfaces instead of looking inverted.
		for a, b := 0, len(face)-1; a < b; a, b = a+1, b-1 {
			face[a], face[b] = face[b], face[a]
		}
		cube.Faces = append(cube.Faces, effects.WordFace{Contours: [][]int{face}, Material: i})
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
