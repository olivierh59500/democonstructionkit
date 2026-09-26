package presets

import "github.com/olivierh59500/democonstructionkit/motion"

// WindowedTextWrap starts a bitmap message one viewport width to the left and
// resets there after its right edge moves strictly past the viewport's right
// edge. Draw at At(0) before Step to retain the boundary frame. Speed is the
// positive leftward distance per update; the caller may change it at any time.
func WindowedTextWrap(contentWidth, viewportWidth, speed float64) motion.WrapBankConfig {
	return motion.WrapBankConfig{
		Start: []float64{-viewportWidth}, Velocity: []float64{-speed},
		Lower: &motion.WrapLimit{Boundary: viewportWidth - contentWidth, Restart: -viewportWidth},
	}
}

func UnionTNT2TextWrap(contentWidth float64) motion.WrapBankConfig {
	return WindowedTextWrap(contentWidth, 640, 4)
}

func UnionReplicantsTextWrap(contentWidth float64) motion.WrapBankConfig {
	return WindowedTextWrap(contentWidth, 640, 6)
}
