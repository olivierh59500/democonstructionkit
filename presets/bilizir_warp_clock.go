package presets

import (
	"fmt"

	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// BilizirStripWarpConfig adapts one live clock to any StripWarp instance.
// Text and logo may share the clock but set their own WarpVariation afterward.
func BilizirStripWarpConfig(clock *motion.WarpTableClock) (composite.StripWarpConfig, error) {
	if clock == nil {
		return composite.StripWarpConfig{}, fmt.Errorf("presets: nil Bilizir warp clock")
	}
	return composite.StripWarpConfig{
		RowHeight: 2, ColumnWidth: 16, VerticalBias: 35,
		SampleX: func(row int, _ kit.Frame) int { return clock.SampleX(row) },
		OffsetY: func(column int, _ kit.Frame) float64 { return clock.OffsetY(column) },
	}, nil
}
