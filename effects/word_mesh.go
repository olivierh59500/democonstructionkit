package effects

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/render"
)

// WordFace separates polygon topology, a material value and a borrowed texture
// index. Multiple contours retain holes; a repeated closing index is removed.
type WordFace struct {
	Contours          [][]int
	Material, Texture int
}
type WordMeshModel struct {
	Points []motion.WrappedPoint
	Faces  []WordFace
}
type WordMeshPose struct {
	Model       int
	Angles      [3]int16
	Translation [2]int16
	Depth       int16
	Visible     bool
	Phase       int
}
type WordMeshInstance struct {
	Offset geometry.Vec2
	Hidden bool
}
type WordMaterialSample struct {
	Face, Material, Phase int
	Point, Screen, Center geometry.Vec2
}
type WordMeshConfig struct {
	Models        []WordMeshModel
	Matrix        *motion.WordEulerMatrix
	Projection    geometry.WordProjectionConfig
	Textures      []*ebiten.Image
	CullBackFaces bool
	// FaceClip is an inclusive projected-coordinate rectangle before placement.
	FaceClip        *[4]float64 // minX,minY,maxX,maxY
	ClipVertexLimit int
	DestinationClip image.Rectangle
	Offset          geometry.Vec2
	PerFaceBatch    bool
	FillRule        ebiten.FillRule
	Blend           ebiten.Blend
	Filter          ebiten.Filter
	Address         ebiten.Address
	BatchTriangles  int
	MaxInstances    int
	InstanceCount   int
	Instance        func(index int, frame kit.Frame) WordMeshInstance
	UV              func(WordMaterialSample) geometry.Vec2
	Color           func(kit.Frame) color.NRGBA
	Animate         func(kit.Frame) WordMeshPose
}
type wordPreparedFace struct {
	contours          [][]geometry.Vec2
	counts            []int
	center            geometry.Vec2
	visible           bool
	material, texture int
}
type wordOwnedModel struct {
	model    *geometry.WordModel
	faces    []WordFace
	prepared []wordPreparedFace
}

// WordMesh owns model banks, cached projection/culling/clipping and batched
// material rendering. Authored clocks can SetPose, or Animate can supply a pose.
// Textures stay borrowed; an omitted texture bank creates one lazy white pixel.
type WordMesh struct {
	config        WordMeshConfig
	models        []wordOwnedModel
	pose          WordMeshPose
	frame         kit.Frame
	valid, closed bool
	batch         *render.Batch
	clipA, clipB  []geometry.Vec2
	instances     []WordMeshInstance
	count         int
	white         *ebiten.Image
	vertices      []ebiten.Vertex
	err           error
}

func wordMeshFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && math.Abs(v) <= math.MaxFloat32
}

func NewWordMesh(c WordMeshConfig) (*WordMesh, error) {
	if len(c.Models) < 1 || len(c.Models) > 256 || c.Matrix == nil || !wordMeshFinite(c.Offset.X) || !wordMeshFinite(c.Offset.Y) || c.FillRule < ebiten.FillRuleFillAll || c.FillRule > ebiten.FillRuleEvenOdd {
		return nil, fmt.Errorf("effects: invalid word mesh models, matrix or placement")
	}
	if c.BatchTriangles == 0 {
		c.BatchTriangles = 4096
	}
	if c.MaxInstances == 0 {
		c.MaxInstances = 1024
	}
	if c.InstanceCount == 0 {
		c.InstanceCount = 1
	}
	if c.ClipVertexLimit == 0 {
		c.ClipVertexLimit = 256
	}
	if c.BatchTriangles < 2 || c.BatchTriangles > 20000 || c.MaxInstances < 1 || c.MaxInstances > 65536 || c.InstanceCount < 1 || c.InstanceCount > c.MaxInstances || c.ClipVertexLimit < 3 || c.ClipVertexLimit > 65536 {
		return nil, fmt.Errorf("effects: invalid word mesh resource budget")
	}
	for _, t := range c.Textures {
		if t == nil {
			return nil, fmt.Errorf("effects: nil word material")
		}
	}
	if c.FaceClip != nil {
		r := *c.FaceClip
		for _, v := range r {
			if !wordMeshFinite(v) {
				return nil, fmt.Errorf("effects: invalid word clip")
			}
		}
		if r[0] > r[2] || r[1] > r[3] {
			return nil, fmt.Errorf("effects: inverted word clip")
		}
		c.FaceClip = &r
	}
	m := &WordMesh{config: c, models: make([]wordOwnedModel, len(c.Models)), batch: render.NewBatch(c.BatchTriangles), instances: make([]WordMeshInstance, c.MaxInstances), count: c.InstanceCount}
	pointCount, cacheCount, indexCount, vertexCount := 0, 0, 0, 0
	for index, model := range c.Models {
		pointCount += len(model.Points)
		if pointCount > 1<<20 || len(model.Faces) > 65536 {
			return nil, fmt.Errorf("effects: word mesh exceeds model budget")
		}
		bank, err := geometry.NewWordModel(geometry.WordModelConfig{Points: model.Points, Projection: c.Projection})
		if err != nil {
			return nil, err
		}
		owned := wordOwnedModel{model: bank, faces: make([]WordFace, len(model.Faces)), prepared: make([]wordPreparedFace, len(model.Faces))}
		for fi, face := range model.Faces {
			if face.Material < 0 || face.Texture < 0 || len(c.Textures) == 0 && face.Texture != 0 || len(c.Textures) > 0 && face.Texture >= len(c.Textures) || len(face.Contours) < 1 {
				return nil, fmt.Errorf("effects: invalid word mesh face material")
			}
			copyFace := WordFace{Material: face.Material, Texture: face.Texture, Contours: make([][]int, len(face.Contours))}
			prepared := wordPreparedFace{material: face.Material, texture: face.Texture, contours: make([][]geometry.Vec2, len(face.Contours)), counts: make([]int, len(face.Contours))}
			for ci, indices := range face.Contours {
				indexCount += len(indices)
				if indexCount > 1<<20 {
					return nil, fmt.Errorf("effects: too many word face indices")
				}
				if len(indices) > 1 && indices[0] == indices[len(indices)-1] {
					indices = indices[:len(indices)-1]
				}
				if len(indices) < 3 {
					return nil, fmt.Errorf("effects: word contour needs three vertices")
				}
				for _, pi := range indices {
					if pi < 0 || pi >= len(model.Points) {
						return nil, fmt.Errorf("effects: word face index outside model")
					}
				}
				copyFace.Contours[ci] = append([]int(nil), indices...)
				capacity := len(indices)
				if c.FaceClip != nil {
					capacity = c.ClipVertexLimit
					if len(indices) > capacity {
						return nil, fmt.Errorf("effects: word contour exceeds clipping budget")
					}
				}
				cacheCount += capacity
				if cacheCount > 1<<20 {
					return nil, fmt.Errorf("effects: word geometry cache exceeds budget")
				}
				prepared.contours[ci] = make([]geometry.Vec2, capacity)
				vertexCount = max(vertexCount, capacity)
			}
			owned.faces[fi], owned.prepared[fi] = copyFace, prepared
		}
		m.models[index] = owned
	}
	m.config.Models = nil
	m.vertices = make([]ebiten.Vertex, vertexCount)
	m.config.Textures = append([]*ebiten.Image(nil), c.Textures...)
	if c.FaceClip != nil {
		m.clipA = make([]geometry.Vec2, c.ClipVertexLimit)
		m.clipB = make([]geometry.Vec2, c.ClipVertexLimit)
	}
	m.batch.Options.FillRule = c.FillRule
	m.batch.Options.Blend = c.Blend
	m.batch.Options.Filter = c.Filter
	m.batch.Options.Address = c.Address
	return m, nil
}

func (m *WordMesh) SetInstanceCount(count int) error {
	if m == nil || m.closed || count < 0 || count > len(m.instances) {
		return fmt.Errorf("effects: invalid word instance count")
	}
	m.count = count
	return nil
}
func (m *WordMesh) SetPose(p WordMeshPose) error {
	if m == nil || m.closed || p.Model < 0 || p.Model >= len(m.models) {
		return fmt.Errorf("effects: invalid word mesh pose")
	}
	if !p.Visible {
		m.pose = p
		m.valid = false
		return nil
	}
	if m.valid && p.Model == m.pose.Model && p.Angles == m.pose.Angles && p.Translation == m.pose.Translation && p.Depth == m.pose.Depth {
		m.pose = p
		return nil
	}
	m.valid = false
	matrix, ok := m.config.Matrix.Sample(p.Angles)
	if !ok {
		return fmt.Errorf("effects: invalid word matrix phase")
	}
	model := &m.models[p.Model]
	if err := model.model.Project(geometry.WordPose{Matrix: matrix, Translation: p.Translation, Depth: p.Depth}); err != nil {
		return err
	}
	points := model.model.Points()
	for fi, face := range model.faces {
		prepared := &model.prepared[fi]
		prepared.visible = true
		for _, indices := range face.Contours {
			for _, pi := range indices {
				if !points[pi].Visible {
					prepared.visible = false
				}
			}
		}
		indices := face.Contours[0]
		if !prepared.visible || m.config.CullBackFaces && geometry.WordWinding(points[indices[0]], points[indices[1]], points[indices[2]]) < 0 {
			prepared.visible = false
			continue
		}
		minX, minY, maxX, maxY := 32767, 32767, -32768, -32768
		for ci, indices := range face.Contours {
			buf := prepared.contours[ci]
			for i, pi := range indices {
				v := points[pi]
				buf[i] = geometry.Vec2{X: float64(v.X), Y: float64(v.Y)}
				minX = min(minX, int(v.X))
				maxX = max(maxX, int(v.X))
				minY = min(minY, int(v.Y))
				maxY = max(maxY, int(v.Y))
			}
			count := len(indices)
			if m.config.FaceClip != nil {
				var err error
				count, err = m.clip(buf[:count], buf)
				if err != nil {
					return err
				}
			}
			prepared.counts[ci] = count
		}
		prepared.center = geometry.Vec2{X: float64(minX + (maxX-minX)/2), Y: float64(minY + (maxY-minY)/2)}
	}
	m.pose = p
	m.valid = true
	return nil
}

func (m *WordMesh) clip(input, output []geometry.Vec2) (int, error) {
	count := copy(m.clipA, input)
	a, b := m.clipA, m.clipB
	r := *m.config.FaceClip
	for axis := 0; axis < 4 && count > 0; axis++ {
		next := 0
		inside := func(p geometry.Vec2) bool {
			switch axis {
			case 0:
				return p.X >= r[0]
			case 1:
				return p.X <= r[2]
			case 2:
				return p.Y >= r[1]
			default:
				return p.Y <= r[3]
			}
		}
		intersection := func(p, q geometry.Vec2) geometry.Vec2 {
			if axis < 2 {
				x := r[0]
				if axis == 1 {
					x = r[2]
				}
				t := (x - p.X) / (q.X - p.X)
				return geometry.Vec2{X: x, Y: p.Y + t*(q.Y-p.Y)}
			}
			y := r[1]
			if axis == 3 {
				y = r[3]
			}
			t := (y - p.Y) / (q.Y - p.Y)
			return geometry.Vec2{X: p.X + t*(q.X-p.X), Y: y}
		}
		appendPoint := func(p geometry.Vec2) error {
			if next >= len(b) {
				return fmt.Errorf("effects: word clipping exceeds vertex budget")
			}
			b[next] = p
			next++
			return nil
		}
		previous := a[count-1]
		for _, current := range a[:count] {
			if inside(previous) != inside(current) {
				if err := appendPoint(intersection(previous, current)); err != nil {
					return 0, err
				}
			}
			if inside(current) {
				if err := appendPoint(current); err != nil {
					return 0, err
				}
			}
			previous = current
		}
		a, b, count = b, a, next
	}
	copy(output, a[:count])
	return count, nil
}

func (m *WordMesh) Update(f kit.Frame) error {
	if m == nil || m.closed || !wordMeshFinite(f.Time) {
		return fmt.Errorf("effects: invalid word mesh update")
	}
	if m.err != nil {
		return m.err
	}
	m.frame = f
	if m.config.Animate != nil {
		return m.SetPose(m.config.Animate(f))
	}
	return nil
}
func (m *WordMesh) Err() error {
	if m == nil {
		return nil
	}
	return m.err
}
func (m *WordMesh) ProjectedPoints() []geometry.WordScreenPoint {
	if m == nil || m.pose.Model < 0 || m.pose.Model >= len(m.models) {
		return nil
	}
	return m.models[m.pose.Model].model.Points()
}
func (m *WordMesh) Draw(dst *ebiten.Image) {
	if m == nil || m.closed || !m.valid || dst == nil || m.count == 0 {
		return
	}
	m.err = nil
	if !m.config.DestinationClip.Empty() {
		r := dst.Bounds().Intersect(m.config.DestinationClip)
		if r.Empty() {
			return
		}
		dst = dst.SubImage(r).(*ebiten.Image)
	}
	for i := 0; i < m.count; i++ {
		instance := WordMeshInstance{}
		if m.config.Instance != nil {
			instance = m.config.Instance(i, m.frame)
		}
		if !wordMeshFinite(instance.Offset.X) || !wordMeshFinite(instance.Offset.Y) {
			m.err = fmt.Errorf("effects: invalid word instance placement")
			return
		}
		m.instances[i] = instance
	}
	textures := m.config.Textures
	if len(textures) == 0 {
		if m.white == nil {
			m.white = ebiten.NewImage(1, 1)
			m.white.Fill(color.White)
		}
		textures = []*ebiten.Image{m.white}
	}
	paint := color.NRGBA{255, 255, 255, 255}
	if m.config.Color != nil {
		paint = m.config.Color(m.frame)
	}
	base := render.Vertex(0, 0, 0, 0, paint)
	faces := m.models[m.pose.Model].prepared
	if !m.config.PerFaceBatch {
		texture := textures[0]
		for _, face := range faces {
			if face.visible && textures[face.texture] != texture {
				m.err = fmt.Errorf("effects: mixed textures require per-face batching")
				return
			}
		}
		m.batch.Begin(dst, texture)
	}
	used := 0
	for instanceIndex := 0; instanceIndex < m.count; instanceIndex++ {
		instance := m.instances[instanceIndex]
		if instance.Hidden {
			continue
		}
		for fi, face := range faces {
			if !face.visible {
				continue
			}
			if m.config.PerFaceBatch {
				m.batch.Begin(dst, textures[face.texture])
				used = 0
			}
			for ci, points := range face.contours {
				count := face.counts[ci]
				if count < 3 {
					continue
				}
				if m.config.FillRule != ebiten.FillRuleFillAll && used+count-2 > m.config.BatchTriangles {
					m.batch.Discard()
					m.err = fmt.Errorf("effects: word parity run exceeds batch budget")
					return
				}
				used += count - 2
				for i := 0; i < count; i++ {
					point := points[i]
					screen := geometry.Vec2{X: point.X + m.config.Offset.X + instance.Offset.X, Y: point.Y + m.config.Offset.Y + instance.Offset.Y}
					uv := geometry.Vec2{}
					if m.config.UV != nil {
						uv = m.config.UV(WordMaterialSample{Face: fi, Material: face.material, Phase: m.pose.Phase, Point: point, Screen: screen, Center: face.center})
					}
					if !wordMeshFinite(screen.X) || !wordMeshFinite(screen.Y) || !wordMeshFinite(uv.X) || !wordMeshFinite(uv.Y) {
						m.batch.Discard()
						m.err = fmt.Errorf("effects: invalid word material coordinates")
						return
					}
					v := base
					v.DstX, v.DstY, v.SrcX, v.SrcY = float32(screen.X), float32(screen.Y), float32(uv.X), float32(uv.Y)
					m.vertices[i] = v
				}
				m.batch.Fan(count, func(i int) ebiten.Vertex { return m.vertices[i] })
			}
			if m.config.PerFaceBatch {
				m.batch.Flush()
			}
		}
	}
	if !m.config.PerFaceBatch {
		m.batch.Flush()
	}
}
func (m *WordMesh) Close() error {
	if m == nil || m.closed {
		return nil
	}
	m.closed = true
	if m.white != nil {
		m.white.Deallocate()
	}
	m.white = nil
	m.models = nil
	m.instances = nil
	m.clipA = nil
	m.clipB = nil
	m.vertices = nil
	m.config.Textures = nil
	return nil
}

var _ kit.Effect = (*WordMesh)(nil)
