package composite

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// CellWaveBank combines independent row and column harmonics on both axes.
// Empty banks contribute zero; multiple waves in one bank are added.
type CellWaveBank struct {
	XRows, XColumns motion.Waves
	YRows, YColumns motion.Waves
}

// HarmonicCellWarpConfig splits an image into rigid cells and moves them with
// separable row/column waves. UseTime chooses Frame.Time instead of Tick as the
// wave clock. Cell, Gap and Source follow CellWarpConfig conventions.
type HarmonicCellWarpConfig struct {
	Cell, Gap image.Point
	Source    image.Rectangle
	Filter    ebiten.Filter
	Blend     ebiten.Blend
	Waves     CellWaveBank
	UseTime   bool
}

// HarmonicCellWarp precomputes each row and column offset once per frame.
// Cropped source images are borrowed and cached by the underlying CellWarp.
type HarmonicCellWarp struct {
	config    HarmonicCellWarpConfig
	warp      *CellWarp
	xRows     []float64
	xColumns  []float64
	yRows     []float64
	yColumns  []float64
	lastClock float64
	prepared  bool
}

func NewHarmonicCellWarp(config HarmonicCellWarpConfig) (*HarmonicCellWarp, error) {
	w := &HarmonicCellWarp{config: config}
	if err := w.SetWaves(config.Waves); err != nil {
		return nil, err
	}
	base, err := NewCellWarp(CellWarpConfig{
		Cell: config.Cell, Gap: config.Gap, Source: config.Source,
		Filter: config.Filter, Blend: config.Blend, Sample: w.sample,
	})
	if err != nil {
		return nil, err
	}
	w.warp = base
	return w, nil
}

// SetWaves changes the waveform at a cue without replacing image crops.
// New banks are copied, so later caller edits cannot change a running effect.
func (w *HarmonicCellWarp) SetWaves(bank CellWaveBank) error {
	for _, waves := range [...]motion.Waves{bank.XRows, bank.XColumns, bank.YRows, bank.YColumns} {
		for _, wave := range waves {
			for _, value := range [...]float64{wave.Amplitude, wave.Spatial, wave.Speed, wave.Phase, wave.Offset} {
				if math.IsNaN(value) || math.IsInf(value, 0) {
					return fmt.Errorf("composite: nonfinite harmonic cell wave")
				}
			}
		}
	}
	w.config.Waves = CellWaveBank{
		XRows: append(motion.Waves(nil), bank.XRows...), XColumns: append(motion.Waves(nil), bank.XColumns...),
		YRows: append(motion.Waves(nil), bank.YRows...), YColumns: append(motion.Waves(nil), bank.YColumns...),
	}
	w.prepared = false
	return nil
}

// DrawAt samples all row/column waves, then draws cells with no trigonometry
// inside the per-cell loop. Repeated draws at the same clock reuse the offsets.
func (w *HarmonicCellWarp) DrawAt(dst, source *ebiten.Image, frame kit.Frame, x, y float64) {
	if w == nil || dst == nil || source == nil {
		return
	}
	bounds := source.Bounds()
	if !w.config.Source.Empty() {
		bounds = bounds.Intersect(w.config.Source)
	}
	if bounds.Empty() {
		return
	}
	clock := float64(frame.Tick)
	if w.config.UseTime {
		clock = frame.Time
	}
	if math.IsNaN(clock) || math.IsInf(clock, 0) {
		return
	}
	rows := 1 + (bounds.Dy()-1)/w.config.Cell.Y
	columns := 1 + (bounds.Dx()-1)/w.config.Cell.X
	if !w.prepared || w.lastClock != clock || len(w.xRows) != rows || len(w.xColumns) != columns {
		w.prepareOffsets(rows, columns, clock)
	}
	w.warp.DrawAt(dst, source, frame, x, y)
}

func (w *HarmonicCellWarp) prepareOffsets(rows, columns int, clock float64) {
	w.xRows = resizeCellOffsets(w.xRows, rows)
	w.yRows = resizeCellOffsets(w.yRows, rows)
	w.xColumns = resizeCellOffsets(w.xColumns, columns)
	w.yColumns = resizeCellOffsets(w.yColumns, columns)
	for row := 0; row < rows; row++ {
		w.xRows[row] = w.config.Waves.XRows.At(float64(row), clock)
		w.yRows[row] = w.config.Waves.YRows.At(float64(row), clock)
	}
	for column := 0; column < columns; column++ {
		w.xColumns[column] = w.config.Waves.XColumns.At(float64(column), clock)
		w.yColumns[column] = w.config.Waves.YColumns.At(float64(column), clock)
	}
	w.lastClock = clock
	w.prepared = true
}

func resizeCellOffsets(values []float64, size int) []float64 {
	if cap(values) < size {
		return make([]float64, size)
	}
	return values[:size]
}

func (w *HarmonicCellWarp) sample(row, column int, _ kit.Frame) CellTransform {
	pose := CellTransform{}
	pose.GeoM.Translate(w.xRows[row]+w.xColumns[column], w.yRows[row]+w.yColumns[column])
	return pose
}

func (w *HarmonicCellWarp) Close() error {
	if w == nil {
		return nil
	}
	w.prepared = false
	w.xRows, w.xColumns, w.yRows, w.yColumns = nil, nil, nil, nil
	return w.warp.Close()
}
